package blogrenderer_test

import (
	"bytes"
	"io"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/fredrikaverpil/go-playground/lgwt/blogrenderer"
)

// func TestRender_String(t *testing.T) {
// 	t.Skip("skipping because string would be way too long to keep inline")
// 	aPost := blogrenderer.Post{
// 		Title:       "hello world",
// 		Body:        "This is a post",
// 		Description: "This is a description",
// 		Tags:        []string{"go", "tdd"},
// 	}
//
// 	t.Run("it converts a single post into HTML", func(t *testing.T) {
// 		buf := bytes.Buffer{}
// 		// err := blogrenderer.RenderManual(&buf, aPost)
// 		err := blogrenderer.Render(&buf, aPost)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
//
// 		got := buf.String()
// 		want := `<h1>hello world</h1>
//
// <p>This is a description</p>
//
// Tags: <ul><li>go</li><li>tdd</li></ul>
// `
//
// 		if got != want {
// 			t.Errorf("got '%s' want '%s'", got, want)
// 		}
// 	})
// }

func TestRender(t *testing.T) {
	aPost := blogrenderer.Post{
		Title:       "hello world",
		Body:        "This is a post",
		Description: "This is a description",
		Tags:        []string{"go", "tdd"},
	}

	postRenderer, err := blogrenderer.NewPostRenderer()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("it converts a single post into HTML", func(t *testing.T) {
		buf := bytes.Buffer{}

		if err := postRenderer.Render(&buf, aPost); err != nil {
			t.Fatal(err)
		}

		// Approval test; use file on disk to assert on large text output instead of keeping inline string in here.
		approvals.VerifyString(t, buf.String())
	})

	t.Run("it renders an index of posts", func(t *testing.T) {
		buf := bytes.Buffer{}
		posts := []blogrenderer.Post{{Title: "Hello World"}, {Title: "Hello World 2"}}

		if err := postRenderer.RenderIndex(&buf, posts); err != nil {
			t.Fatal(err)
		}

		approvals.VerifyString(t, buf.String())
	})
}

func BenchmarkRender(b *testing.B) {
	aPost := blogrenderer.Post{
		Title:       "hello world",
		Body:        "This is a post",
		Description: "This is a description",
		Tags:        []string{"go", "tdd"},
	}

	postRenderer, err := blogrenderer.NewPostRenderer()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = postRenderer.Render(io.Discard, aPost)
	}
}
