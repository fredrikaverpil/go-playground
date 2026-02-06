package main

import (
	"encoding/base32"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// NewSystemGeneratedBase32 does not always produce an AIP-122 compliant ID.
// For more details, see https://google.aip.dev/122#resource-id-segments
func NewSystemGeneratedBase32() string {
	base32Encoding := base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)
	regexp := regexp.MustCompile(`^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$`)
	// Retry creating id until it matches the regexp
	for {
		id := uuid.New()
		encodedID := base32Encoding.EncodeToString(id[:])
		if regexp.MatchString(encodedID) {
			fmt.Printf("Valid: %s\n", encodedID)
			return encodedID
		} else {
			fmt.Printf("Not valid: %s\n", encodedID)
			panic("whoops")
		}
	}
}

func main() {
	for range 100 {
		NewSystemGeneratedBase32()
	}
}
