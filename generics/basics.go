// Package generics is a step-by-step tour of Go generics (type parameters),
// added in Go 1.18.
//
// Read it top-to-bottom: each step is a little more advanced than the last.
// Every example is exercised by a table test in basics_test.go
// (from this directory, run: go test -v).
package generics

import (
	"cmp"
	"fmt"
)

// Step 1: Generic functions and the `any` constraint -- "any type at all"
//
// A generic function is parameterised by a TYPE, not just by values. The type
// parameter list goes in square brackets *before* the value parameters:
//
//	func name[TypeParam Constraint](valueParams) returnType
//
// Without generics you'd either copy reverse() once per element type, or use
// `any` and lose type safety. With a type parameter, one definition serves many
// types.
//
// `any` (an alias for interface{}, Go 1.18) is the loosest constraint:
// satisfied by every type. The trade-off is that, knowing nothing about T, you
// may only do what works for *all* types -- copy it, store it, index slices of
// it, return it. You may NOT use ==, <, or + on a `T any`, because not every
// type supports those (comparable and cmp.Ordered unlock those; next steps).
//
// Why not just take []any? Because that throws away the element type:
// reverse([]int{...}) would return []any, forcing boxing and type assertions at
// every call site. With [T any] the concrete type is preserved -- feed it a
// []int and you get a []int back.
//
// Note we call reverse([]int{...}), never reverse[int]([]int{...}): the type
// argument is *inferred* from the value arguments. There is one place inference
// can't help you -- you'll meet it at step 7.
func reverse[T any](s []T) []T {
	out := make([]T, len(s))
	for i, v := range s {
		out[len(s)-1-i] = v
	}
	return out
}

// Step 2: The `comparable` constraint -- unlocks == and !=
//
// `comparable` is a built-in constraint (Go 1.18). It permits any type whose
// values can be compared with == and != -- numbers, strings, bools, pointers,
// channels, and arrays/structs built from comparable types. (Slices, maps and
// funcs are NOT comparable and won't satisfy it.)
//
// This is the piece `any` was missing: reverse() never compared elements, but
// contains() must, so `any` is too weak and `comparable` is exactly enough.
// (It's the same reason map keys must be comparable.)
//
// Boundary example that intentionally does NOT compile:
//
//	_ = contains([][]int{{1}}, []int{1}) // []int is a slice, and slices are not comparable
func contains[T comparable](s []T, v T) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}

// Step 3: The cmp.Ordered constraint -- unlocks < <= > >=
//
// cmp.Ordered (package cmp, Go 1.21) is the third ready-made constraint and the
// most powerful: it permits the ordered types (every numeric type and string)
// and unlocks the ordering operators < <= > >=. It also includes == and !=,
// because every ordered type is comparable -- so cmp.Ordered is a *superset of
// comparable in power, but accepts fewer types* (a struct can be comparable yet
// is not ordered). The rule of thumb: more operations, fewer types.
//
// cmp.Ordered is itself a union of types with a ~ on every line -- exactly the
// ingredients we assemble by hand in steps 4 and 5. (Before Go 1.21 it lived in
// golang.org/x/exp/constraints as constraints.Ordered; the stdlib cmp version
// now supersedes it.)
func maxOf[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Step 4: Custom constraints and type unions
//
// A constraint is just an interface used in the [T ...] position. So far we've
// borrowed pre-made ones (any, comparable, cmp.Ordered); now we write our own.
// (For common numeric sets you often needn't: golang.org/x/exp/constraints
// already defines Integer, Float, Signed, Unsigned and Complex -- an external
// golang.org/x module. We hand-roll Number here to show what those are made of.)
//
// Inside an interface, listing concrete types separated by `|` forms a *union*:
// Number reads as "T must be int, int64 or float64". Because every type in the
// union supports +, sum() is allowed to use + on values of type T.
//
// One catch: an interface that lists types (this union, or the ~ forms in step
// 5) is CONSTRAINT-ONLY -- it may appear in [T ...] but never as an ordinary
// type. `var n Number` does not compile ("interface contains type constraints").
// Plain method-only interfaces like fmt.Stringer (step 6) have no such limit.
//
// Note the union matches those EXACT types only. A defined type such as
// `type Celsius float64` has underlying type float64 but is NOT float64, so it
// does not satisfy Number. Step 5 introduces `~`, which removes this restriction.
//
// Boundary examples that intentionally do NOT compile:
//
//	var n Number             // constraint-only interface; cannot use as a value type
//	_ = sum([]Celsius{1, 2}) // Celsius is not exactly float64
//	_ = sum([]uint{1, 2})    // uint is not listed in Number
type Number interface {
	int | int64 | float64
}

// sum adds up every element of s. The zero value of T (var total T) is a valid
// starting accumulator for every numeric type in the union.
func sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

// Step 5: The ~ operator -- matching underlying types
//
// Celsius is a *defined type*: its underlying type is float64, but it is a
// distinct type. With step 4's strict Number, a Celsius is rejected, because the
// union listed float64 exactly -- not "things based on float64".
//
// Putting `~` in front of a type means "any type whose UNDERLYING type is this".
// So ~float64 accepts float64 AND Celsius (and any other `type X float64`).
// RealNumber is identical to Number except for the ~ -- that single change is
// what lets double() below accept a Celsius.
//
// This is exactly why cmp.Ordered is written with a ~ on every line: it wants to
// accept your defined types (type UserID int, type Money int64, ...), not just
// the predeclared ones.
type Celsius float64 // NOTE: not a type alias!

type RealNumber interface {
	~int | ~int64 | ~float64
}

// double returns x + x. Thanks to the ~ in RealNumber it works on defined types
// like Celsius, not only the predeclared numeric types.
func double[T RealNumber](x T) T {
	return x + x
}

// Step 6: Constraints that require methods
//
// Every constraint so far has been a *type set*: it answers "which concrete
// types may T be?" (any, comparable, cmp.Ordered, a union, a ~union). But a
// constraint is just an interface, and interfaces have a second, older job --
// listing METHODS. So a constraint can also answer "which methods must T have?",
// and inside the function you may CALL those methods on values of type T.
//
// Any interface you already have works as a constraint. Here T is constrained to
// the standard library's fmt.Stringer (interface { String() string }), so the
// compiler guarantees every element has a String method to call.
//
// Why constrain [T fmt.Stringer] instead of just taking a []fmt.Stringer? Same
// reason as step 1's []any: a []fmt.Stringer boxes every element and discards
// the concrete type, whereas [T fmt.Stringer] keeps your []Celsius a []Celsius
// (and lets several parameters share the one concrete T).
func describe[T fmt.Stringer](items []T) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.String()
	}
	return out
}

// Celsius (from step 5) gains a String method, so it now also satisfies
// fmt.Stringer -- the same defined type can sit in a type-set constraint
// (RealNumber) AND a method-set constraint.
func (c Celsius) String() string {
	return fmt.Sprintf("%g°C", float64(c))
}

// THE FULL PICTURE: one constraint may require BOTH a type set and methods. The
// types it permits are the INTERSECTION -- a type qualifies only if its
// underlying type is in the union AND it has the methods. FloatStringer accepts
// a type only when it is based on float64 *and* has String(): Celsius qualifies;
// a plain float64 does not (no String method); a `type Name string` with a
// String method does not (wrong underlying type).
type FloatStringer interface {
	~float64
	fmt.Stringer
}

// scale uses both halves of its constraint at once: + comes from the ~float64
// type set, and String() comes from the embedded fmt.Stringer.
func scale[T FloatStringer](x T) string {
	doubled := x + x        // operator: allowed by ~float64
	return doubled.String() // method: allowed by fmt.Stringer
}

// Step 7: Generic types
//
// Type parameters aren't only for functions -- a *type* can have them too. Slice
// is a generic type: E is its type parameter, `any` its constraint. The
// definition reads "a Slice of E is a []E".
//
// KEY DIFFERENCE from functions: a generic type has no arguments to infer E
// from, so you must INSTANTIATE it explicitly with a type argument -- Slice[int]
// or Slice[string], never a bare `Slice`. (This is the teaser from step 1.)
//
// The underlying rule is one: inference reads type arguments from VALUE
// arguments. A generic type has none, so it always needs them spelled out -- and
// so does a function whose type parameter appears only in its RESULTS, e.g.
// `func Zero[T any]() T`: nothing in the call pins down T, so you must write
// Zero[int](). Whenever a value argument does determine T (every function up to
// now), inference fills it in for you.
//
// One version wrinkle: Go's inference rules have improved over time, and newer
// Go versions can infer in a few contexts older versions could not. If an
// example from recent Go docs fails on an older toolchain, try spelling out the
// type argument first: f[int](x).
//
// Underneath, a Slice[int] is exactly a []int, so all slice built-ins (len,
// append, range, indexing) work on it unchanged.
type Slice[E any] []E

// Step 8: Methods on a generic type
//
// To attach a method to a generic type, the receiver restates the type parameter
// by NAME (no constraint): `(s Slice[E])`. Inside the method E is the element
// type, so Add can accept and return values of the right type. Returning
// Slice[E] lets calls chain.
//
// THE BIG LIMITATION TODAY (Go 1.26): no method -- on ANY type, generic or not -- may
// declare type parameters OF ITS OWN. A method may only use the type parameters
// it inherits from its receiver (that is what Add's E is above). So this is
// illegal:
//
//	func (s Slice[E]) Map[U any](f func(E) U) Slice[U] // ERROR: methods cannot have type parameters
//
// This isn't specific to slice-based types: a generic struct like Pair (step 9)
// equally cannot give a method its own type parameter. There are no "generic
// methods" in Go 1.26. When you need an extra type parameter, write a standalone
// generic FUNCTION instead: move the receiver into the parameter list and add the
// type parameter there.
//
// Looking ahead: in Go 1.27 this restriction is being lifted. The proposal
// "generic methods for Go" (golang/go#77273) is accepted and milestoned for
// 1.27, so the Map method above will become legal. Two caveats to remember:
//   - It only applies to CONCRETE (non-interface) methods. Interface methods
//     still can't have type parameters, so a generic method does NOT satisfy any
//     interface (e.g. a generic Read[E any] would not implement io.Reader).
//     The mental model: "a generic method is a generic function with a receiver".
//   - Calls work like generic functions: s.Map[string](f), or inferred s.Map(f).
func (s Slice[E]) Add(item E) Slice[E] {
	return append(s, item)
}

// Step 9: Multiple type parameters
//
// A generic may declare more than one type parameter. The clearest minimal
// example is Pair: a little container holding two values of independent types T
// and U. There's deliberately no algorithm here, so nothing competes with the
// point -- T and U are simply two separate type parameters. The shorthand
// [T, U any] gives both the `any` constraint; list them separately (e.g.
// [T any, U comparable]) when their constraints differ.
//
// Like any generic type (step 7), Pair must be instantiated explicitly in a
// composite literal: Pair[int, string]{...}. But a constructor FUNCTION can
// infer the type arguments from its values -- the same inference we saw for
// functions earlier.
type Pair[T, U any] struct {
	First  T
	Second U
}

// MakePair builds a Pair, inferring T and U from its arguments, so you can write
// MakePair(1, "a") instead of Pair[int, string]{1, "a"}.
func MakePair[T, U any](first T, second U) Pair[T, U] {
	return Pair[T, U]{First: first, Second: second}
}

// Step 10: Constraining the SHAPE of a type (+ constraint type inference)
//
// Every constraint so far has restricted what T *is* (which types, which
// methods). With `~` you can also restrict T's STRUCTURE: `~[]E` means "any type
// whose underlying type is a slice of E" (likewise ~map[K]V, ~chan E). Expressing
// it needs two type parameters working together (step 9): S for the container and
// E for its element, with S constrained in terms of E.
//
// Why bother, when `func([]E) []E` already works on slices? It's the
// preserve-the-concrete-type theme again -- step 1's []any vs [T any], step 6's
// []fmt.Stringer vs [T fmt.Stringer] -- now one level up: a plain []E parameter
// accepts a Slice[int] (step 7) but hands you back a bare []int, losing the named
// type. With [S ~[]E] the function returns the SAME named type S it was given.
//
// CONSTRAINT TYPE INFERENCE: you never spell these out. E is inferred from S, and
// S is inferred from the argument, so you write Filter(xs, keep), never
// Filter[Slice[int], int](xs, keep). This is how the standard library's slices
// and maps packages are written.
func Filter[S ~[]E, E any](s S, keep func(E) bool) S {
	// make (not var out S) preallocates and guarantees a non-nil result: Filter
	// returns an empty slice, never nil, even when nothing matches.
	out := make(S, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Aside: naming convention -- why T, E, K, V, ...?
//
// By now you've seen a few names: T (steps 1-6, 9), E (steps 7-8, 10), U (step
// 9), and S (step 10). Type parameter names can be ANY identifier -- whole words
// included -- but
// by convention they are short and UPPERCASE, with length tracking scope (the
// same idea behind `i` for a loop index). The common letters:
//
//	T = a general Type    E = Element        S = Slice (or type Set)
//	K = map Key           V = map Value      R = Result
//
// When no letter is meaningful, the convention is to continue the alphabet from
// T: T, U, V, W. Pair above is the textbook case -- its two components are just
// "some type" and "some other type", so T and U fit perfectly.
//
// Reach for a single letter by default; use a descriptive word when the role
// isn't obvious -- e.g. a key/value Pair reads well as Pair[Key, Value any].
// (The standard library uses words too: maps.Keys names one type parameter `Map`.)

// Aside: when NOT to use generics
//
// Generics are best when the implementation is genuinely the same for every
// allowed type: reverse, contains, sum, maxOf and Filter all have one algorithm
// that does not care about the concrete type beyond its constraint.
//
// Do NOT reach for generics just to avoid writing two short functions, or when
// different types need different behavior. In those cases ordinary functions,
// methods, or small interfaces are often clearer. A useful rule of thumb: if the
// body mostly switches on the concrete type, uses reflection, or has many
// type-specific branches, generics probably are not buying much.
//
// Also remember that constraints are part of the API. A very broad constraint
// can promise too little to be useful (`any` cannot be compared or added), while
// a very narrow one may reject perfectly good callers. Prefer the weakest
// constraint that still permits the operations the function actually needs.
