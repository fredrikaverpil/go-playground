# http

The code from the
["HTTP from scratch" series](https://kmcd.dev/series/http-from-scratch/).

## v0.9

[Original blog post](https://kmcd.dev/posts/http0.9-from-scratch/)

### Run the server

```bash
cd v0.9
go run ./cmd/server
```

### Run the client

```bash
go run ./cmd/client
```

or use `curl`:

```bash
curl --http0.9 http://127.0.0.1:9000/this/is/a/test
```

## v1.0

[Original blog post](https://kmcd.dev/posts/http1.0-from-scratch/)

- Headers: Metadata that added context and control to requests and responses
  (RFC 1945 4.2).
- Methods: A diverse set of actions (POST, HEAD, PUT, DELETE, etc.) beyond just
  retrieving documents. The web was no longer read-only. (RFC 1945 8).
- Status Codes: Clear signals about the outcome of requests, paving the way for
  better error handling and redirection (RFC 1945 6.1.1).
- Content Negotiation: The ability to request specific formats (RFC 1945 10.5),
  encoding (RFC 1945 10.3) or languages (RFC 1945 D.2.5) for content.

### TODO

- [ ] Add `public/` directory with HTML files for the static file server
- [ ] Implement `cmd/client/main.go`
