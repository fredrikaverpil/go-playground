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
