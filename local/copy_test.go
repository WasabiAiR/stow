package local_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cheekybits/is"
	"github.com/flyteorg/stow"
	"github.com/flyteorg/stow/local"
)

func TestCopy(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()
	l, err := stow.Dial(local.Kind, stow.ConfigMap{"path": testDir})
	is.NoErr(err)

	containers, _, err := l.Containers("", stow.CursorStart, 10)
	is.NoErr(err)
	is.True(len(containers) > 2)
	srcContainer, dstContainer := containers[1], containers[2]

	const content = "the content to copy"
	src, err := srcContainer.Put("dir/src.txt", strings.NewReader(content), int64(len(content)), nil)
	is.NoErr(err)

	copier, ok := dstContainer.(stow.Copier)
	is.True(ok)
	copied, err := copier.Copy(t.Context(), src, "other/dst.txt")
	is.NoErr(err)
	size, err := copied.Size()
	is.NoErr(err)
	is.Equal(size, int64(len(content)))

	dst, err := dstContainer.Item(copied.ID())
	is.NoErr(err)
	r, err := dst.Open()
	is.NoErr(err)
	defer r.Close()
	got, err := io.ReadAll(r)
	is.NoErr(err)
	is.Equal(string(got), content)

	// The source is untouched.
	r, err = src.Open()
	is.NoErr(err)
	defer r.Close()
	got, err = io.ReadAll(r)
	is.NoErr(err)
	is.Equal(string(got), content)
}

func TestCopyErrors(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()
	l, err := stow.Dial(local.Kind, stow.ConfigMap{"path": testDir})
	is.NoErr(err)
	containers, _, err := l.Containers("", stow.CursorStart, 10)
	is.NoErr(err)
	container := containers[1]
	copier := container.(stow.Copier)

	_, err = copier.Copy(t.Context(), nil, "dst.txt")
	is.Err(err)

	src, err := container.Put("src.txt", strings.NewReader("x"), 1, nil)
	is.NoErr(err)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = copier.Copy(ctx, src, "dst.txt")
	is.Equal(err, context.Canceled)
}
