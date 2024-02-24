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

type Skipper struct {
	name string
	skip bool
}

func Test_Level_1_Skipper(t *testing.T) {
	for _, c := range []Skipper{{name: "Level 2a", skip: false}, {name: "Level 2b", skip: true}} {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if c.skip {
				t.Skip()
			}
			t.Run("Level 3", func(t *testing.T) {
				if c.skip {
					t.Skip()
				}
			})
		})
	}
}
