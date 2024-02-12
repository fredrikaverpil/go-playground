package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_Level_1(t *testing.T) {
	t.Parallel()
	t.Run("Level 2", func(t *testing.T) {
		t.Parallel()
		t.Run("Level 3", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, 1, 1)
		})
	})
}
