package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/errgroup"
)

var (
	// multipartUploadThreshold is the size from which a body that can be read
	// at any offset is uploaded in parts, the threshold of the transfer
	// manager. A smaller body is uploaded with one PutObject request.
	multipartUploadThreshold int64 = 16 << 20
	// uploadPartSize is the smallest part size of such a multipart upload. It
	// grows when the body would need more than uploadMaxParts parts.
	uploadPartSize int64 = 8 << 20
	// uploadConcurrency is how many parts of one upload are sent at once.
	uploadConcurrency = 5
)

const (
	// uploadMaxParts is the most parts S3 accepts in one multipart upload.
	uploadMaxParts = 10000
	// abortTimeout bounds the abort of a failed multipart upload.
	abortTimeout = 30 * time.Second
)

// readerAtSeeker is a body that can be read at any offset, such as an
// *os.File, so its requests can read their bytes straight from it.
type readerAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// putStreamed uploads the length bytes of r from offset, as found by
// bodyRange, without holding them in memory. The transfer manager reads every
// body up to its multipart threshold into memory, and allocates the buffers of
// all its parts up front for a bigger one, even when the body is a file.
func (c *container) putStreamed(ctx context.Context, name string, r io.ReaderAt, offset, length int64, metadata map[string]string) error {
	if length >= multipartUploadThreshold {
		return c.multipartUpload(ctx, name, r, offset, length, metadata)
	}
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:            new(c.name),
		Key:               new(name),
		Body:              io.NewSectionReader(r, offset, length),
		ContentLength:     new(length),
		Metadata:          metadata,
		ChecksumAlgorithm: c.checksumAlgorithm(),
	})
	return err
}

// multipartUpload uploads length bytes of r from offset in parts, each read
// from r as it is sent.
func (c *container) multipartUpload(ctx context.Context, name string, r io.ReaderAt, offset, length int64, metadata map[string]string) error {
	algorithm := c.checksumAlgorithm()
	upload, err := c.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:            new(c.name),
		Key:               new(name),
		Metadata:          metadata,
		ChecksumAlgorithm: algorithm,
	})
	if err != nil {
		return fmt.Errorf("put, creating the multipart upload: %w", err)
	}

	abort := func() {
		// The caller's context may be the reason for the failure, and the
		// parts are billed until the upload is aborted. Keep its values, such
		// as the trace, but not its cancellation, and bound the abort so it
		// can't hang the failing Put.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), abortTimeout)
		defer cancel()
		_, _ = c.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   new(c.name),
			Key:      new(name),
			UploadId: upload.UploadId,
		})
	}

	partSize := partSizeFor(length)
	parts := make([]types.CompletedPart, (length+partSize-1)/partSize)
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(uploadConcurrency)
	for i := range parts {
		start := int64(i) * partSize
		size := min(partSize, length-start)
		number := int32(i + 1)
		g.Go(func() error {
			res, err := c.client.UploadPart(gctx, &s3.UploadPartInput{
				Bucket:            new(c.name),
				Key:               new(name),
				UploadId:          upload.UploadId,
				PartNumber:        new(number),
				Body:              io.NewSectionReader(r, offset+start, size),
				ContentLength:     new(size),
				ChecksumAlgorithm: algorithm,
			})
			if err != nil {
				return fmt.Errorf("put, uploading part %d: %w", number, err)
			}
			parts[i] = types.CompletedPart{
				ETag:              res.ETag,
				PartNumber:        new(number),
				ChecksumCRC32:     res.ChecksumCRC32,
				ChecksumCRC32C:    res.ChecksumCRC32C,
				ChecksumCRC64NVME: res.ChecksumCRC64NVME,
				ChecksumSHA1:      res.ChecksumSHA1,
				ChecksumSHA256:    res.ChecksumSHA256,
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		abort()
		return err
	}

	_, err = c.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          new(c.name),
		Key:             new(name),
		UploadId:        upload.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: parts},
	})
	if err != nil {
		abort()
		return fmt.Errorf("put, completing the multipart upload: %w", err)
	}
	return nil
}

// checksumAlgorithm is the algorithm of the checksums sent with an upload,
// picked as the transfer manager picks it: CRC32, unless the client only
// calculates the checksums an operation requires.
func (c *container) checksumAlgorithm() types.ChecksumAlgorithm {
	if c.client.Options().RequestChecksumCalculation == aws.RequestChecksumCalculationWhenRequired {
		return ""
	}
	return types.ChecksumAlgorithmCrc32
}

// partSizeFor is the part size of a multipart upload of length bytes.
func partSizeFor(length int64) int64 {
	return max(uploadPartSize, (length+uploadMaxParts-1)/uploadMaxParts)
}

// bodyRange returns the current position of r and the number of bytes from
// there to its end, leaving the position where it was.
func bodyRange(r io.Seeker) (offset, length int64, err error) {
	offset, err = r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, 0, err
	}
	end, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, 0, err
	}
	if _, err := r.Seek(offset, io.SeekStart); err != nil {
		return 0, 0, err
	}
	return offset, end - offset, nil
}
