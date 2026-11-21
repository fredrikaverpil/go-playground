# go-playground 🌱

This is my personal playground for experimenting and learning new things in
[Go](https://go.dev).

## Development

Run tests, linting etc with [mise](https://mise.jdx.dev):

```sh
mise tasks  # list tasks
mise run    # run all checks
```

Go tasks run in every module at or below the current directory, so run them from
a folder to target a specific experiment:

```sh
cd <folder>
mise run go:test
```
