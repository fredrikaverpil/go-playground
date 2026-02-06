# Functional Options in Go

Four variants of the functional options pattern for configuring structs via
optional, composable arguments.

## Variants

### 1. Simple function (`func_option_test.go`, `server.go`)

The most common variant, popularized by [Dave Cheney][cheney]. The option type
is a plain function that mutates the target.

```go
type Option func(*Server)

func WithAddr(addr string) Option {
    return func(s *Server) { s.addr = addr }
}

s := NewServer(WithAddr(":9090"), WithMaxConns(500))
```

### 2. Interface-based (`interface_option_test.go`)

Used by gRPC (`grpc.DialOption`) and Google Cloud client libraries. An
unexported `apply` method prevents external implementations, giving the library
full control over valid options.

```go
type clientOption interface { apply(*client) }

type clientOptionFunc func(*client)
func (f clientOptionFunc) apply(c *client) { f(c) }

c := newClient(withEndpoint("https://staging.example.com"))
```

Supports both function adapters and dedicated structs for options that carry
multiple values or need their own validation.

### 3. Error-returning (`error_option_test.go`)

Options return an error, enabling validation at construction time. The
constructor short-circuits on the first invalid option.

```go
type databaseOption func(*database) error

func withPort(port int) databaseOption {
    return func(db *database) error {
        if port < 1 || port > 65535 {
            return fmt.Errorf("port %d out of range", port)
        }
        db.port = port
        return nil
    }
}

db, err := newDatabase(withPort(3306))
```

### 4. Self-referential (`selfref_option_test.go`)

[Rob Pike's original][pike] design. Each option returns the previous value,
enabling save-and-restore for temporary overrides.

```go
type cacheOption func(*cache) cacheOption

func withMaxItems(n int) cacheOption {
    return func(c *cache) cacheOption {
        prev := c.maxItems
        c.maxItems = n
        return withMaxItems(prev)
    }
}

undo := withMaxItems(42)(c)
undo(c) // restores previous value
```

## When to use which

| Variant          | Use when                                                    |
| ---------------- | ----------------------------------------------------------- |
| Simple function  | Default choice. Covers most cases.                          |
| Interface-based  | Library code that needs to restrict who can create options. |
| Error-returning  | Invalid config should fail fast at construction time.       |
| Self-referential | You need to undo/restore option changes (e.g. in tests).    |

## References

- [Dave Cheney: Functional options for friendly APIs][cheney]
- [Rob Pike: Self-referential functions and the design of options][pike]

[cheney]:
  https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
[pike]:
  https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
