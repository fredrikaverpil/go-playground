# protobuf-duration-nilness

## The problem

Should we allow `0` duration in the communication protocol between services?

Zero-value duration could be problematic or even dangerous in practice:

- **Division by zero** — `(current / duration.Minutes()) * 100` produces
  `Infinity` or `NaN`.
- **Ambiguous semantics** — does `0` mean "instant", "unset", or "don't do
  this"? It depends on context, and neither side of the contract can be sure.
- **Bad UX** — a user sees `0 minutes` in an input field and doesn't know if it
  means unset or intentional.
- **Database footguns** — some schemas evaluate `0` to `null`, leading to
  [Null Island](https://en.wikipedia.org/wiki/Null_Island)-style bugs.
- **Zero as default** — `0` is the default/initial value for duration in most
  languages, making it easy for uninitialized values to slip through.

The core issue: if the contract allows zero duration, both frontend and backend
must constantly guard against it. Neither side can trust the other.

## The solution

Two things make this a non-issue with protobuf in Go:

### 1. `google.protobuf.Duration` preserves nil vs zero

A common misconception is that proto3 collapses `0` into `nil`. This is only
true for **scalar** fields (`int32`, `string`, etc.). `Duration` is a **message
type**, so it generates as a pointer (`*durationpb.Duration`) in Go:

| Value       | `GetMaxDuration() == nil` | JSON (`protojson`) |
| ----------- | ------------------------- | ------------------ |
| Unset       | `true`                    | `null`             |
| Zero (`0s`) | `false`                   | `"0s"`             |
| `5m`        | `false`                   | `"300s"`           |

This distinction survives proto binary serialization and gRPC round-trips.

### 2. protovalidate rejects zero duration at the contract level

Using [protovalidate](https://github.com/bufbuild/protovalidate), we declare the
constraint directly in the proto schema:

```proto
google.protobuf.Duration max_duration = 2 [
  (buf.validate.field).duration.gt = {}  // must be > 0s
];
```

This means:

- **nil** — accepted (field not set, "I'm not specifying a duration")
- **> 0s** — accepted (valid duration)
- **0s** — rejected with `InvalidArgument` before the handler is even called

The protovalidate interceptor enforces this server-side as gRPC middleware, so
zero duration never reaches application code. Both sides can trust the contract.

## Prerequisites

- [Go](https://go.dev/)
- [buf](https://buf.build/)

## Code generation

```sh
buf generate
```

## Tests

```sh
go test -v ./...
```
