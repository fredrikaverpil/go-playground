# protobuf-duration-nilness

Demonstrates that `google.protobuf.Duration` (a message type) preserves the
distinction between **unset (nil)** and **zero duration** in Go — even when sent
over the wire via gRPC.

An echo gRPC server receives a `Task` with a `Duration` field and returns it
unchanged. The tests verify that nil, zero, and non-zero durations all survive
the round-trip.

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
