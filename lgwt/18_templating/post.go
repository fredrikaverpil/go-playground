package blogrenderer

import (
	"strings"
)

// Post is a representation of a post
type Post struct {
	Title, Description, Body string
	Tags                     []string
}

// SanitisedTitle returns the title of the post with spaces replaced by dashes for pleasant URLs.
// Without this, the post would e.g. contain %20 for spaces.
func (p Post) SanitisedTitle() string {
	return strings.ToLower(strings.ReplaceAll(p.Title, " ", "-"))
}
