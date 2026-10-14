package s3

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/stow"
)

// fakeS3 is enough of S3 for PutObject, the multipart upload calls and
// HeadObject on one bucket, with path-style addressing.
type fakeS3 struct {
	t       *testing.T
	mu      sync.Mutex
	objects map[string][]byte
	parts   map[int][]byte
	puts    int // PutObject requests
	creates int // CreateMultipartUpload requests
	aborts  int // AbortMultipartUpload requests
	// onPart, if set, is called with the number of each part uploaded.
	onPart func(n int)
}

func newFakeS3(t *testing.T) *fakeS3 {
	return &fakeS3{t: t, objects: map[string][]byte{}, parts: map[int][]byte{}}
}

type completeRequest struct {
	Parts []struct {
		ETag       string
		PartNumber int
	} `xml:"Part"`
}

func (f *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := r.URL.Path
	q := r.URL.Query()
	body, err := io.ReadAll(r.Body)
	require.NoError(f.t, err)

	switch {
	case r.Method == http.MethodHead:
		obj, ok := f.objects[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("ETag", `"etag"`)
		w.Header().Set("Content-Length", strconv.Itoa(len(obj)))
	case r.Method == http.MethodPost && q.Has("uploads"):
		f.creates++
		fmt.Fprint(w, `<InitiateMultipartUploadResult><UploadId>upload</UploadId></InitiateMultipartUploadResult>`)
	case r.Method == http.MethodPut && q.Has("partNumber"):
		n, err := strconv.Atoi(q.Get("partNumber"))
		require.NoError(f.t, err)
		f.parts[n] = body
		if f.onPart != nil {
			f.onPart(n)
		}
		w.Header().Set("ETag", fmt.Sprintf(`"part-%d"`, n))
	case r.Method == http.MethodPost && q.Has("uploadId"):
		var req completeRequest
		require.NoError(f.t, xml.Unmarshal(body, &req))
		var obj []byte
		for i, p := range req.Parts {
			assert.Equal(f.t, i+1, p.PartNumber, "parts out of order")
			assert.Equal(f.t, fmt.Sprintf(`"part-%d"`, p.PartNumber), p.ETag)
			obj = append(obj, f.parts[p.PartNumber]...)
		}
		f.objects[key] = obj
		fmt.Fprint(w, `<CompleteMultipartUploadResult><ETag>"etag"</ETag></CompleteMultipartUploadResult>`)
	case r.Method == http.MethodDelete && q.Has("uploadId"):
		f.aborts++
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPut:
		f.puts++
		f.objects[key] = body
		w.Header().Set("ETag", `"etag"`)
	default:
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func fakeContainer(t *testing.T, f *fakeS3) stow.Container {
	server := httptest.NewServer(f)
	t.Cleanup(server.Close)
	location, err := stow.Dial(Kind, stow.ConfigMap{
		ConfigAccessKeyID: "access-key",
		ConfigSecretKey:   "secret-key",
		ConfigRegion:      "us-east-1",
		ConfigEndpoint:    server.URL,
	})
	require.NoError(t, err)
	container, err := location.Container("bucket")
	require.NoError(t, err)
	return container
}

// readAtOnly is a body that can only be read at an offset: reading it from
// the start, as a buffering uploader does, fails.
type readAtOnly struct {
	*bytes.Reader
}

func (readAtOnly) Read([]byte) (int, error) {
	return 0, errors.New("the body was read sequentially")
}

func content(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i % 251)
	}
	return b
}

func TestPutStreamsSmallBodyInOneRequest(t *testing.T) {
	f := newFakeS3(t)
	c := fakeContainer(t, f)
	data := content(1000)

	body := readAtOnly{bytes.NewReader(data)}
	// Put sends the body from its current position, as a retry after seeking
	// back expects.
	_, err := body.Seek(10, io.SeekStart)
	require.NoError(t, err)
	_, err = c.Put("dir/small", body, int64(len(data)-10), nil)
	require.NoError(t, err)

	assert.Equal(t, 1, f.puts)
	assert.Equal(t, 0, f.creates)
	assert.Equal(t, data[10:], f.objects["/bucket/dir/small"])
}

func TestPutStreamsLargeBodyInParts(t *testing.T) {
	oldThreshold, oldPart := multipartUploadThreshold, uploadPartSize
	multipartUploadThreshold, uploadPartSize = 1000, 300
	defer func() { multipartUploadThreshold, uploadPartSize = oldThreshold, oldPart }()

	f := newFakeS3(t)
	c := fakeContainer(t, f)
	data := content(2000)

	_, err := c.Put("dir/large", readAtOnly{bytes.NewReader(data)}, int64(len(data)), nil)
	require.NoError(t, err)

	assert.Equal(t, 0, f.puts)
	assert.Equal(t, 1, f.creates)
	assert.Equal(t, 0, f.aborts)
	parts := make([]int, 0, len(f.parts))
	for n := range f.parts {
		parts = append(parts, n)
	}
	sort.Ints(parts)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, parts)
	assert.Equal(t, data, f.objects["/bucket/dir/large"])
}

func TestPutBuffersReaderWithoutOffsets(t *testing.T) {
	f := newFakeS3(t)
	c := fakeContainer(t, f)
	data := content(1000)

	// A plain reader still goes through the transfer manager.
	_, err := c.Put("dir/stream", io.MultiReader(bytes.NewReader(data)), int64(len(data)), nil)
	require.NoError(t, err)

	assert.Equal(t, 1, f.puts)
	assert.Equal(t, data, f.objects["/bucket/dir/stream"])
}

func TestPutBuffersFileThatCannotSeek(t *testing.T) {
	f := newFakeS3(t)
	c := fakeContainer(t, f)
	data := content(1000)

	// The read end of a pipe is an *os.File, but seeking it fails, so it is
	// read to the end through the transfer manager.
	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	defer pr.Close()
	go func() {
		_, _ = pw.Write(data)
		_ = pw.Close()
	}()
	_, err = c.Put("dir/pipe", pr, int64(len(data)), nil)
	require.NoError(t, err)

	assert.Equal(t, 1, f.puts)
	assert.Equal(t, data, f.objects["/bucket/dir/pipe"])
}

func TestPutBuffersLargeFileThatCannotSeek(t *testing.T) {
	f := newFakeS3(t)
	c := fakeContainer(t, f)
	// Above the transfer manager's 16 MiB multipart threshold.
	data := content(17 << 20)

	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	defer pr.Close()
	go func() {
		_, _ = pw.Write(data)
		_ = pw.Close()
	}()
	_, err = c.Put("dir/pipe", pr, int64(len(data)), nil)
	require.NoError(t, err)

	assert.Equal(t, 0, f.puts)
	assert.Equal(t, 1, f.creates)
	assert.Equal(t, 0, f.aborts)
	assert.True(t, bytes.Equal(data, f.objects["/bucket/dir/pipe"]), "uploaded object differs")
}

func TestPutAbortsAfterCallerCancels(t *testing.T) {
	oldThreshold, oldPart, oldConcurrency := multipartUploadThreshold, uploadPartSize, uploadConcurrency
	multipartUploadThreshold, uploadPartSize, uploadConcurrency = 1000, 300, 1
	defer func() {
		multipartUploadThreshold, uploadPartSize, uploadConcurrency = oldThreshold, oldPart, oldConcurrency
	}()

	f := newFakeS3(t)
	c := fakeContainer(t, f).(*container)
	data := content(2000)

	// The caller gives up while the parts are uploading. The upload must
	// still be aborted, or its parts stay billed.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f.onPart = func(n int) {
		if n == 2 {
			cancel()
		}
	}
	_, err := c.PutContext(ctx, "dir/large", readAtOnly{bytes.NewReader(data)}, int64(len(data)), nil)
	require.ErrorIs(t, err, context.Canceled)

	assert.Equal(t, 1, f.creates)
	assert.Equal(t, 1, f.aborts)
	assert.NotContains(t, f.objects, "/bucket/dir/large")
}

func TestPartSizeFor(t *testing.T) {
	assert.Equal(t, uploadPartSize, partSizeFor(1))
	assert.Equal(t, uploadPartSize, partSizeFor(uploadPartSize*uploadMaxParts))
	// Past that the parts grow so that there are never more than uploadMaxParts.
	length := uploadPartSize*uploadMaxParts + 1
	size := partSizeFor(length)
	assert.Greater(t, size, uploadPartSize)
	assert.LessOrEqual(t, (length+size-1)/size, int64(uploadMaxParts))
}
