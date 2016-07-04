package util

import "github.com/graymeta/stow"

// Walker walks the stow.Item objects inside a
// stow.Container.
type Walker struct {
	container stow.Container
	itemChan  chan stow.Item
	errChan   chan error
	doneChan  chan struct{}
}

// NewWalker makes a new walker for the
// given Container.
func NewWalker(container stow.Container) *Walker {
	return &Walker{
		container: container,
		itemChan:  make(chan stow.Item),
		errChan:   make(chan error),
		doneChan:  make(chan struct{}),
	}
}

// ItemChan gets a channel on which walked stow.Item
// types are returned.
func (w *Walker) ItemChan() <-chan stow.Item {
	return w.itemChan
}

// ErrChan gets a channel on which any errors are
// returned.
func (w *Walker) ErrChan() <-chan error {
	return w.errChan
}

// DoneChan is closed when the walker has finished walking.
func (w *Walker) DoneChan() <-chan struct{} {
	return w.doneChan
}

// Start begins walking. DoneChan() channel is closed
// when walking is complete.
func (w *Walker) Start() {
	go func() {
		defer close(w.doneChan)
		page := 0
		more := true
		for more {
			var items []stow.Item
			var err error
			items, more, err = w.container.Items(page)
			if err != nil {
				w.errChan <- err
				continue
			}
			for _, item := range items {
				w.itemChan <- item
			}
			page++
		}
	}()
}
