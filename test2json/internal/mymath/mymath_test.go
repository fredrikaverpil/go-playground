// mymath_test.go
package mymath

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	expected := 5
	if result != expected {
		t.Errorf("Add(2, 3) = %d; want %d", result, expected)
	}
}

func TestSubtract(t *testing.T) {
	// A dummy failing test for demonstration
	t.Run("PositiveNumbers", func(t *testing.T) {
		if (5 - 2) != 3 {
			t.Error("5 - 2 should be 3")
		}
	})
	t.Run("NegativeResult", func(t *testing.T) {
		if (2 - 5) != -3 {
			t.Error("2 - 5 should be -3")
		}
	})
	// A deliberately failing sub-test
	t.Run("FailingSubTest", func(t *testing.T) {
		t.Error("This sub-test is designed to fail")
	})
}
