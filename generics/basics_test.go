package generics

import (
	"testing"

	"gotest.tools/v3/assert"
)

// Step 1: Generic functions and the `any` constraint
//
// reverse() doesn't care what the elements are, so it works for any type. Here
// asserting on the *whole* returned slice with assert.DeepEqual is exactly the
// idiom we want. Note we never write reverse[int](...) -- the type is *inferred*
// from the arguments.
func TestReverse(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name string
			in   []int
			want []int
		}{
			{name: "several", in: []int{1, 2, 3}, want: []int{3, 2, 1}},
			{name: "single", in: []int{42}, want: []int{42}},
			{name: "empty", in: []int{}, want: []int{}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := reverse(tt.in) // returns []int, not []any
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct {
			name string
			in   []string
			want []string
		}{
			{name: "two", in: []string{"a", "b"}, want: []string{"b", "a"}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := reverse(tt.in) // returns []string
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})
}

// Step 2: The `comparable` constraint
//
// contains() needs ==, which `any` cannot provide -- so the element type is
// constrained to `comparable`.
func TestContains(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name string
			in   []int
			v    int
			want bool
		}{
			{name: "present", in: []int{1, 2, 3}, v: 2, want: true},
			{name: "absent", in: []int{1, 2, 3}, v: 9, want: false},
			{name: "empty", in: []int{}, v: 1, want: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := contains(tt.in, tt.v)
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct {
			name string
			in   []string
			v    string
			want bool
		}{
			{name: "present", in: []string{"a", "b"}, v: "b", want: true},
			{name: "absent", in: []string{"a", "b"}, v: "z", want: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := contains(tt.in, tt.v)
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})
}

// Step 3: The cmp.Ordered constraint
//
// maxOf() needs the ordering operator >, which only cmp.Ordered provides. The
// string subtest shows cmp.Ordered also orders strings lexicographically.
func TestMaxOf(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name string
			a, b int
			want int
		}{
			{name: "a bigger", a: 5, b: 3, want: 5},
			{name: "b bigger", a: 2, b: 9, want: 9},
			{name: "equal", a: 4, b: 4, want: 4},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := maxOf(tt.a, tt.b)
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct {
			name string
			a, b string
			want string
		}{
			{name: "lexicographic", a: "apple", b: "banana", want: "banana"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := maxOf(tt.a, tt.b)
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})
}

// Step 4: Custom constraints and type unions
//
// sum() works for any type in the Number union. We test the two ends we care
// about: an integer type and a floating-point type.
func TestSum(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name string
			in   []int
			want int
		}{
			{name: "several", in: []int{1, 2, 3}, want: 6},
			{name: "empty", in: []int{}, want: 0},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := sum(tt.in)
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})

	t.Run("float64", func(t *testing.T) {
		tests := []struct {
			name string
			in   []float64
			want float64
		}{
			{name: "halves", in: []float64{0.5, 0.5, 1}, want: 2},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := sum(tt.in)
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})
}

// Step 5: The ~ operator
//
// double() accepts a defined type (Celsius) because RealNumber uses ~. With the
// strict Number constraint from step 4, the Celsius case would not compile.
func TestDouble(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		tests := []struct {
			name string
			in   float64
			want float64
		}{
			{name: "positive", in: 2.5, want: 5},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := double(tt.in)
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})

	t.Run("Celsius", func(t *testing.T) {
		tests := []struct {
			name string
			in   Celsius
			want Celsius
		}{
			{name: "room temp", in: Celsius(20), want: Celsius(40)},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := double(tt.in)
				assert.DeepEqual(t, got, tt.want)
			})
		}
	})
}

// Step 6: Constraints that require methods
//
// describe is constrained to fmt.Stringer, so it can call String() on each
// element. We pass a []Celsius -- a concrete type kept concrete, not boxed into
// []fmt.Stringer.
func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		in   []Celsius
		want []string
	}{
		{name: "several", in: []Celsius{0, 20, 100}, want: []string{"0°C", "20°C", "100°C"}},
		{name: "empty", in: []Celsius{}, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := describe(tt.in)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

// Step 6 (continued): combining a type set with methods
//
// scale needs BOTH + (from ~float64) and String() (from fmt.Stringer), so only a
// type satisfying both -- like Celsius -- can instantiate FloatStringer.
func TestScale(t *testing.T) {
	tests := []struct {
		name string
		in   Celsius
		want string
	}{
		{name: "room temp", in: Celsius(20), want: "40°C"},
		{name: "freezing", in: Celsius(0), want: "0°C"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scale(tt.in)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

// Step 7: Generic types
//
// Note the explicit instantiation everywhere -- Slice[int], Slice[string] --
// because a generic type can't infer its element type. Underneath it's an
// ordinary slice, so the built-ins behave exactly as they would on []E.
func TestSlice(t *testing.T) {
	t.Run("is a slice underneath", func(t *testing.T) {
		// Must instantiate explicitly: Slice[int], never a bare Slice.
		got := Slice[int]{1, 2, 3}
		assert.DeepEqual(t, got, Slice[int]{1, 2, 3})
		assert.DeepEqual(t, len(got), 3)
	})

	t.Run("append keeps the named type", func(t *testing.T) {
		// append on a named slice type returns that same named type.
		got := Slice[string]{"a"} //nolint:prealloc // teaching: show append on a named slice type
		got = append(got, "b")
		assert.DeepEqual(t, got, Slice[string]{"a", "b"})
	})
}

// Step 8: Methods on a generic type
//
// Add uses the receiver's type parameter E and returns the same Slice[E] type,
// so calls chain.
func TestSliceAdd(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		got := Slice[int]{1}.Add(2).Add(3)
		assert.DeepEqual(t, got, Slice[int]{1, 2, 3})
	})

	t.Run("string", func(t *testing.T) {
		got := Slice[string]{}.Add("a").Add("b")
		assert.DeepEqual(t, got, Slice[string]{"a", "b"})
	})
}

// Step 9: Multiple type parameters
//
// Pair holds two values of independent types T and U. A composite literal needs
// explicit type arguments; the MakePair constructor infers them from its values.
func TestPair(t *testing.T) {
	t.Run("explicit instantiation", func(t *testing.T) {
		got := Pair[int, string]{First: 1, Second: "a"}
		assert.DeepEqual(t, got, Pair[int, string]{First: 1, Second: "a"})
	})

	t.Run("constructor infers the types", func(t *testing.T) {
		got := MakePair(1, "a") // T=int, U=string inferred
		assert.DeepEqual(t, got, Pair[int, string]{First: 1, Second: "a"})
	})

	t.Run("both fields the same type", func(t *testing.T) {
		// T and U are independent, but nothing stops both being string.
		got := MakePair("hello", "world") // T=string, U=string
		assert.DeepEqual(t, got, Pair[string, string]{First: "hello", Second: "world"})
	})
}

// Step 10: Constraining the shape of a type
//
// Filter preserves the named slice type: a Slice[int] (step 7) goes in and a
// Slice[int] comes back, not a bare []int. S and E are both inferred from the
// argument, so there are no explicit type arguments at the call site.
func TestFilter(t *testing.T) {
	t.Run("keeps the named type", func(t *testing.T) {
		got := Filter(Slice[int]{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
		assert.DeepEqual(t, got, Slice[int]{2, 4})
	})

	t.Run("no matches returns non-nil empty slice", func(t *testing.T) {
		// make(S, 0, len(s)) means Filter returns empty, never nil. assert.DeepEqual
		// already distinguishes []int{} from a nil slice; got != nil says so plainly.
		got := Filter([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
		assert.DeepEqual(t, got, []int{})
		assert.Assert(t, got != nil)
	})

	t.Run("plain slice", func(t *testing.T) {
		got := Filter([]string{"a", "bb", "ccc"}, func(s string) bool { return len(s) > 1 })
		assert.DeepEqual(t, got, []string{"bb", "ccc"})
	})
}
