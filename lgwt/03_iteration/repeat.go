package iteration

import "strings"

func Repeat(c string) string {
	var result strings.Builder
	for range 5 {
		result.WriteString(c)
	}
	return result.String()
}
