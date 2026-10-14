package s3

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/stow"
)

func TestCopySource(t *testing.T) {
	assert.Equal(t, "bucket/a/b%20c/d%2Be.pb", copySource("bucket", "a/b c/d+e.pb"))
}

// TestCopy runs against a real S3 endpoint. Set S3ENDPOINT to use an
// S3-compatible server instead of AWS.
func TestCopy(t *testing.T) {
	accessKeyID := os.Getenv("S3ACCESSKEYID")
	secretKey := os.Getenv("S3SECRETKEY")
	region := os.Getenv("S3REGION")
	if accessKeyID == "" || secretKey == "" || region == "" {
		t.Skip("skipping test because missing one or more of S3ACCESSKEYID S3SECRETKEY S3REGION")
	}

	config := stow.ConfigMap{
		ConfigAccessKeyID: accessKeyID,
		ConfigSecretKey:   secretKey,
		ConfigRegion:      region,
	}
	if endpoint := os.Getenv("S3ENDPOINT"); endpoint != "" {
		config[ConfigEndpoint] = endpoint
		config[ConfigDisableSSL] = "true"
	}

	location, err := stow.Dial(Kind, config)
	require.NoError(t, err)
	defer location.Close()

	srcContainer, err := location.CreateContainer("stow-copy-src-" + randomSuffix(t))
	require.NoError(t, err)
	dstContainer, err := location.CreateContainer("stow-copy-dst-" + randomSuffix(t))
	require.NoError(t, err)
	defer func() {
		for _, c := range []stow.Container{srcContainer, dstContainer} {
			items, _, _ := c.Items("", stow.CursorStart, 100)
			for _, i := range items {
				_ = c.RemoveItem(i.ID())
			}
			_ = location.RemoveContainer(c.ID())
		}
	}()

	// 11 MiB, so the multipart case below needs three parts.
	content := make([]byte, 11<<20)
	_, err = rand.Read(content)
	require.NoError(t, err)
	metadata := map[string]any{"stow": "copy"}

	_, err = srcContainer.Put("dir/src item+1.pb", bytes.NewReader(content), int64(len(content)), metadata)
	require.NoError(t, err)

	check := func(t *testing.T, name string) {
		src, err := srcContainer.Item("dir/src item+1.pb")
		require.NoError(t, err)

		copied, err := dstContainer.(stow.Copier).Copy(t.Context(), src, name)
		require.NoError(t, err)
		assert.Equal(t, name, copied.ID())
		size, err := copied.Size()
		require.NoError(t, err)
		assert.Equal(t, int64(len(content)), size)

		dst, err := dstContainer.Item(name)
		require.NoError(t, err)
		r, err := dst.Open()
		require.NoError(t, err)
		defer r.Close()
		got, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(content, got), "copied content differs")
		md, err := dst.Metadata()
		require.NoError(t, err)
		assert.Equal(t, metadata, md)
	}

	t.Run("single request", func(t *testing.T) {
		check(t, "dir/single.pb")
	})

	t.Run("multipart", func(t *testing.T) {
		oldMax, oldPart := copyObjectMaxSize, copyPartSize
		copyObjectMaxSize, copyPartSize = 5<<20, 5<<20
		defer func() { copyObjectMaxSize, copyPartSize = oldMax, oldPart }()
		check(t, "dir/multipart.pb")
	})

	t.Run("missing source", func(t *testing.T) {
		src, err := srcContainer.Item("dir/src item+1.pb")
		require.NoError(t, err)
		require.NoError(t, srcContainer.RemoveItem(src.ID()))
		_, err = dstContainer.(stow.Copier).Copy(t.Context(), src, "dir/missing.pb")
		assert.Error(t, err)
	})
}

func TestSameService(t *testing.T) {
	aws := &container{}
	assert.True(t, sameService(aws, &container{}))
	assert.True(t, sameService(&container{customEndpoint: "http://a:9000"}, &container{customEndpoint: "http://a:9000"}))
	assert.False(t, sameService(aws, &container{customEndpoint: "http://a:9000"}))
	assert.False(t, sameService(&container{customEndpoint: "http://a:9000"}, &container{customEndpoint: "http://b:9000"}))
}

func TestPartition(t *testing.T) {
	for region, want := range map[string]string{
		"us-east-1":      "aws",
		"eu-central-1":   "aws",
		"cn-north-1":     "aws-cn",
		"us-gov-west-1":  "aws-us-gov",
		"us-iso-east-1":  "aws-iso",
		"us-isob-east-1": "aws-iso-b",
	} {
		client, _, err := newS3Client(stow.ConfigMap{ConfigAccessKeyID: "id", ConfigSecretKey: "key"}, region)
		require.NoError(t, err)
		assert.Equal(t, want, partition(&container{client: client}), region)
	}
}

func TestCopyNilItem(t *testing.T) {
	_, err := (&container{}).Copy(t.Context(), nil, "name")
	assert.Error(t, err)
}

func randomSuffix(t *testing.T) string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	require.NoError(t, err)
	const hex = "0123456789abcdef"
	out := make([]byte, 0, 8)
	for _, v := range b {
		out = append(out, hex[v>>4], hex[v&0xf])
	}
	return string(out)
}
