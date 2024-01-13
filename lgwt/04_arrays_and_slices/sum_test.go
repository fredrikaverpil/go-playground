package arrays_and_slices

import (
	"reflect"
	"slices"
	"testing"
)

func TestSumArray(t *testing.T) {
	numbers := [5]int{1, 2, 3, 4, 5} // array becayse of fixed size of 5

	got := SumArray(numbers)
	want := 15

	if got != want {
		t.Errorf("got %d want %d given, %v", got, want, numbers)
	}
}

func TestSumSlice(t *testing.T) {
	t.Run("collection of 5 numbers", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}

		got := SumSlice(numbers)
		want := 15

		if got != want {
			t.Errorf("got %d want %d given, %v", got, want, numbers)
		}
	})

	t.Run("collection of any size", func(t *testing.T) {
		numbers := []int{1, 2, 3}

		got := SumSlice(numbers)
		want := 6

		if got != want {
			t.Errorf("got %d want %d given, %v", got, want, numbers)
		}
	})
}

func TestSumAll(t *testing.T) {
	got := SumAll([]int{1, 2}, []int{0, 9})
	want := []int{3, 9}

	// got != want does not work for slices
	if !reflect.DeepEqual(got, want) { // reflect.DeepEqual is not type safe
		t.Errorf("got %v want %v", got, want)
	}

	// since go 1.21, this might be better, requires elements to be comparable
	if !slices.Equal(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestSumAllTails(t *testing.T) {
	checkSums := func(t testing.TB, got, want []int) {
		t.Helper()
		if !slices.Equal(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	}

	t.Run("more than one element", func(t *testing.T) {
		got := SumAllTails([]int{1, 2}, []int{0, 9})
		want := []int{2, 9}
		checkSums(t, got, want)
	})

	t.Run("one element", func(t *testing.T) {
		got := SumAllTails([]int{1}, []int{9})
		want := []int{0, 0}
		checkSums(t, got, want)
	})

	t.Run("no elements", func(t *testing.T) {
		got := SumAllTails([]int{}, []int{})
		want := []int{0, 0}
		checkSums(t, got, want)
	})

	t.Run("safely sum empty slices", func(t *testing.T) {
		got := SumAllTails([]int{}, []int{3, 4, 5})
		want := []int{0, 9}
		checkSums(t, got, want)
	})
}
