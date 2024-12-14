package blogposts_test

import (
	"testing"
	"testing/fstest"

	"github.com/fredrikaverpil/go-playground/lgwt/blogposts"
	"gotest.tools/v3/assert"
)

func TestNewBlogPosts(t *testing.T) {
	// Arrange
	const (
		firstBody = `Title: Post 1
Description: Description 1
Tags: tdd, go
---
Hello
World`
		secondBody = `Title: Post 2
Description: Description 2
Tags: rust, borrow-checker
---
B
L
M`
	)
	fs := fstest.MapFS{
		"hello world.md":  {Data: []byte(firstBody)},
		"hello-world2.md": {Data: []byte(secondBody)},
	}
	expectedPost1 := blogposts.Post{
		Title:       "Post 1",
		Description: "Description 1",
		Tags:        []string{"tdd", "go"},
		Body: `Hello
World`,
	}
	expectedPost2 := blogposts.Post{
		Title:       "Post 2",
		Description: "Description 2",
		Tags:        []string{"rust", "borrow-checker"},
		Body: `B
L
M`,
	}

	// Act
	posts, err := blogposts.NewPostsFromFS(fs)

	// Assert
	assert.NilError(t, err)
	assert.Equal(t, len(posts), len(fs))
	assert.DeepEqual(t, posts, []blogposts.Post{expectedPost1, expectedPost2})
}
