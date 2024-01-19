package maps

import "errors"

// Never initialize a map with nil:
// var m map[string]string
// Instead, initialize it with an empty map:
// m := map[string]string{}
// or
// m := make(map[string]string)
type Dictionary map[string]string

var (
	ErrNotFound   = errors.New("could not find the word you were looking for")
	ErrWordExists = errors.New("cannot add word because it already exists")
)

func (d Dictionary) Search(word string) (string, error) {
	value, ok := d[word]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

func (d Dictionary) Add(key, value string) error {
	// Note that it is the pointer to the map that is passed in, so we can modify the map directly.
	d[key] = value
	return nil
}
