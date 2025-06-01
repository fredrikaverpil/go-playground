# Learning about [test2json](https://pkg.go.dev/cmd/test2json)

## Conclusion

There are no significant inherent benefits to switching to the
`go tool test2json` workflow if you have access to the source code and can
execute the tests via `go test -json`.

The `go tool test2json` utility is more useful in scenarios such as:

- Integrating with build systems that produce compiled test binaries as
  artifacts, and you then need to run these binaries and convert their output.
- Running tests on a different machine than where they were compiled.
- Legacy systems or specific CI setups that might already work with compiled
  test executables.

## Quickstart

Compile test binary:

```sh
go test -c github.com/fredrikaverpil/go-playground/test2json/internal/mymath -o mymath.test
```

Execute test binary and process output with test2json:

```sh
./mymath.test -test.v=test2json | go tool test2json
```

> [!NOTE]
>
> It's crucial to pass the `-test.v=test2json` flag to your compiled test
> binary. This tells the test binary to produce output in a format that
> test2json can reliably parse. While just `-test.v` might work,
> `-test.v=test2json` provides higher fidelity results.

Output to JSON file:

```sh
./mytests.test -test.v=test2json | go tool test2json > test_results.json
```

## Options

Noteworthy:

- `-p <pkgname>`: You can tell test2json what package name to report in the JSON
  events. This is useful if the information isn't automatically inferred or if
  you want to override it.
- `-t`: This flag requests that timestamps be added to each JSON test event.

```sh
❯ ./mymath.test -h
Usage of ./mymath.test:
  -test.bench regexp
        run only benchmarks matching regexp
  -test.benchmem
        print memory allocations for benchmarks
  -test.benchtime d
        run each benchmark for duration d or N times if `d` is of the form Nx (default 1s)
  -test.blockprofile file
        write a goroutine blocking profile to file
  -test.blockprofilerate rate
        set blocking profile rate (see runtime.SetBlockProfileRate) (default 1)
  -test.count n
        run tests and benchmarks n times (default 1)
  -test.coverprofile file
        write a coverage profile to file
  -test.cpu list
        comma-separated list of cpu counts to run each test with
  -test.cpuprofile file
        write a cpu profile to file
  -test.failfast
        do not start new tests after the first test failure
  -test.fullpath
        show full file names in error messages
  -test.fuzz regexp
        run the fuzz test matching regexp
  -test.fuzzcachedir string
        directory where interesting fuzzing inputs are stored (for use only by cmd/go)
  -test.fuzzminimizetime value
        time to spend minimizing a value after finding a failing input (default 1m0s)
  -test.fuzztime value
        time to spend fuzzing; default is to run indefinitely
  -test.fuzzworker
        coordinate with the parent process to fuzz random values (for use only by cmd/go)
  -test.gocoverdir string
        write coverage intermediate files to this directory
  -test.list regexp
        list tests, examples, and benchmarks matching regexp then exit
  -test.memprofile file
        write an allocation profile to file
  -test.memprofilerate rate
        set memory allocation profiling rate (see runtime.MemProfileRate)
  -test.mutexprofile string
        write a mutex contention profile to the named file after execution
  -test.mutexprofilefraction int
        if >= 0, calls runtime.SetMutexProfileFraction() (default 1)
  -test.outputdir dir
        write profiles to dir
  -test.paniconexit0
        panic on call to os.Exit(0)
  -test.parallel n
        run at most n tests in parallel (default 8)
  -test.run regexp
        run only tests and examples matching regexp
  -test.short
        run smaller test suite to save time
  -test.shuffle string
        randomize the execution order of tests and benchmarks (default "off")
  -test.skip regexp
        do not list or run tests matching regexp
  -test.testlogfile file
        write test action log to file (for use only by cmd/go)
  -test.timeout d
        panic test binary after duration d (default 0, timeout disabled)
  -test.trace file
        write an execution trace to file
  -test.v
        verbose: print additional output
```
