package b2

import (
	"errors"
	"net/url"
	"strings"

	"github.com/graymeta/stow"
	"gopkg.in/kothar/go-backblaze.v0"
)

type location struct {
	config stow.Config
	client *backblaze.B2
}

func (l *location) Close() error {
	return nil // nothing to close
}

func (l *location) CreateContainer(name string) (stow.Container, error) {
	bucket, err := l.client.CreateBucket(name, backblaze.AllPrivate)
	if err != nil {
		return nil, err
	}
	container := &container{
		bucket: bucket,
	}
	return container, nil
}

func (l *location) Containers(prefix, cursor string) ([]stow.Container, string, error) {

	// At this time, the api and/or library don't seem to support pagination or prefixes
	response, err := l.client.ListBuckets()

	if err != nil {
		return nil, "", err
	}

	containers := make([]stow.Container, 0, len(response))
	for _, cont := range response {
		if prefix == "" || (prefix != "" && strings.HasPrefix(cont.Name, prefix)) {
			containers = append(containers, &container{
				bucket: cont,
			})
		}
	}

	return containers, "", nil
}

// Note: in this case, the 'id' is really the bucket name
func (l *location) Container(id string) (stow.Container, error) {
	bucket, err := l.client.Bucket(id)
	if err != nil || bucket == nil {
		return nil, stow.ErrNotFound
	}

	c := &container{
		bucket: bucket,
	}

	return c, nil
}

func (l *location) ItemByURL(u *url.URL) (stow.Item, error) {

	if u.Scheme != Kind {
		return nil, errors.New("not valid swift URL")
	}

	// b2://f001.backblaze.com/file/<container_name>/<path_to_object>
	pieces := strings.SplitN(u.Path, "/", 4)

	c, err := l.Container(pieces[2])
	if err != nil {
		return nil, err
	}

	filename, err := url.QueryUnescape(pieces[3])
	if err != nil {
		return nil, err
	}
	response, err := c.(*container).bucket.ListFileNames(filename, 1)
	if err != nil {
		return nil, stow.ErrNotFound
	}

	if len(response.Files) != 1 {
		return nil, errors.New("unexpected number of responses from ListFileNames")
	}

	return c.Item(response.Files[0].ID)
}

// Note: in this case, the 'id' is really the bucket name
func (l *location) RemoveContainer(id string) error {
	stowCont, err := l.Container(id)
	if err != nil {
		return err
	}

	return stowCont.(*container).bucket.Delete()
}
