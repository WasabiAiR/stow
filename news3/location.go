package s3

import (
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

type location struct {
	client *s3.S3
	cfg    cfg
}

var _ stow.Location = (*location)(nil)

func (l *location) CreateContainer(containerName string) (stow.Container, error) {

	params := &s3.CreateBucketInput{Bucket: aws.String(containerName)}

	_, err := l.client.CreateBucket(params)
	if err != nil {
		return nil, errors.Wrap(err, "creating container")
	}

	c := &container{
		name:   containerName,
		cfg:    l.cfg,
		client: l.client,
	}

	return c, nil
}

func (l *location) Close() error { return nil }

func (l *location) Containers(prefix, cursor string, count int) ([]stow.Container, string, error) {

	var containers []stow.Container

	list, err := l.client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, "", errors.Wrap(err, "listing containers")
	}

	for _, b := range list.Buckets {

		if !strings.HasPrefix(*b.Name, prefix) {
			continue
		}

		c := &container{
			name:   *b.Name,
			cfg:    l.cfg,
			client: l.client,
		}

		containers = append(containers, c)
	}

	return containers, "", nil

}

func (l *location) Container(id string) (stow.Container, error) {

	_, err := l.client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(id)})
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == s3.ErrCodeNoSuchBucket {
			return nil, stow.ErrNotFound
		}
		return nil, errors.Wrap(err, "getting container")
	}

	c := &container{
		name:   id,
		cfg:    l.cfg,
		client: l.client,
	}

	return c, nil
}

func (l *location) RemoveContainer(id string) error {

	_, err := l.client.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(id)})
	if err != nil {
		return errors.Wrap(err, "deleting a container")
	}

	return nil
}

func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {
	panic("not implemented")
}
