package blogposts

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Post represents a post on a blog
//
// Example frontmatter to handle:
//
// Title: Hello, TDD world!
// Description: First post on our wonderful blog
// Tags: tdd, go
// ---
// Hello world!
//
// The body of posts starts after the `---`.
type Post struct {
	Title       string
	Description string
	Tags        []string
	Body        string
}

const (
	titleSeparator       = "Title: "
	descriptionSeparator = "Description: "
	tagsSeparator        = "Tags: "
)

func newPost(postBody io.Reader) (Post, error) {
	// Implementation as part of excercise. I'm not using it as it depends on the order of calls.
	//
	// readMetaLine := func(tagName string) string {
	// 	scanner.Scan() // advance the scanner to the next line
	// 	x := scanner.Text()
	// 	fmt.Println("x:", x)
	// 	return strings.TrimPrefix(scanner.Text(), tagName)
	// }

	readContents := func() ([]string, []string) {
		var metaLines []string
		var bodyLines []string
		recordBody := false
		scanner := bufio.NewScanner(postBody) // start from the top
		for scanner.Scan() {
			line := scanner.Text()
			if !recordBody {
				metaLines = append(metaLines, line)
			} else {
				bodyLines = append(bodyLines, line)
			}
			if line == "---" {
				recordBody = true
				// break
			}
		}
		return metaLines, bodyLines
	}

	metaLines, bodyLines := readContents()
	body := strings.Join(bodyLines, "\n")
	body = strings.TrimSuffix(body, "\n") // remove last newline

	readMetaLine := func(tagName string) string {
		for _, line := range metaLines {
			if strings.HasPrefix(line, tagName) {
				return strings.TrimPrefix(line, tagName)
			}
		}
		return "" // should return an error here...
	}

	return Post{
		Title:       readMetaLine(titleSeparator),
		Description: readMetaLine(descriptionSeparator),
		Tags:        strings.Split(readMetaLine(tagsSeparator), ", "),
		Body:        body,
	}, nil
}

func readBody(scanner *bufio.Scanner) string {
	scanner.Scan() // ignore a line
	buf := bytes.Buffer{}
	for scanner.Scan() {
		fmt.Fprintln(&buf, scanner.Text())
	}
	return strings.TrimSuffix(buf.String(), "\n")
}
