package stow

// tests for this are in test/test.go

// WalkFunc is the type of the function called for
// each Item visited by Walk.
// If there was a problem,
// the incoming error will describe the problem and
// the function can decide how to handle that error.
// If an error is returned, processing stops.
type WalkFunc func(item Item, err error) error

// Walk walks all Items in the Container.
// Returns the first error returned by the WalkFunc or
// nil if no errors were returned.
func Walk(container Container, prefix string, fn WalkFunc) error {
	var (
		err    error
		items  []Item
		cursor = "start"
	)
	for len(cursor) > 0 {
		items, cursor, err = container.Items(prefix, cursor)
		if err != nil {
			err = fn(nil, err)
			if err != nil {
				return err
			}
		}
		for _, item := range items {
			err = fn(item, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
