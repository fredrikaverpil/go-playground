//go:build integration

package dto_test

import (
	"testing"

	"github.com/fredrikaverpil/go-playground/bugs/neotest-golang/internal/core/dto"
	"gotest.tools/v3/assert"
)

func TestPublic(t *testing.T) {
	assert.Equal(t, dto.Dummy(), "foo")
}
