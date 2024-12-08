//go:build integration

package dto

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestInternal(t *testing.T) {
	assert.Equal(t, Dummy(), "foo")
}
