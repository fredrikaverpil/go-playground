package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("1 + 2 did not equal 3")
	}
}

func TestAddSubTestLevel1(t *testing.T) {
	// NOTE: this is OK.
	t.Run("Add1", func(t *testing.T) {
		if Add(1, 2) != 3 {
			t.Error("1 + 2 did not equal 3")
		}
	})
}

func TestAddSubTestLevel2(t *testing.T) {
	t.Run("Add1", func(t *testing.T) {
		// NOTE: this is OK.
		if Add(1, 2) != 3 {
			t.Error("1 + 2 did not equal 3")
		}

		t.Run("Add2", func(t *testing.T) {
			// FIXME: this causes JSON output in the Output tab, related to https://github.com/nvim-neotest/neotest-go/issues/52
			// FIXME: this test also crashes marshal_gotest_output if run with "nearest test", but passes if test is executed from top level.
			if Add(1, 2) != 3 {
				t.Error("1 + 2 did not equal 3")
			}
		})
	})
}

func TestAddSubTestLevelSkipping(t *testing.T) {
	// FIXME: this causes a crash in marshal_gotest_output unless the '!' is removed.
	for _, skip := range []bool{true, false} {
		t.Run("Subtest1", func(t *testing.T) {
			t.Run("Subtest2", func(t *testing.T) {
				if !skip { // NOTE: remove the '!' to make the test pass without a crash.
					t.Skip("skipping test")
				}
			})
		})
	}
}
