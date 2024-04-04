# Spanner playground

A playground for learning Spanner using the emulator.

## Quickstart :rocket:

```bash
go test -v -count=1 -race ./...
```

The `-count=1` avoids caching.

## TestMain

The `main_test.go` contains the bulk of the setup, which spins up
a docker container with the emulator and then tears it down after
having executed a test.

The test setup also creates the Spanner instance as well as the db.
