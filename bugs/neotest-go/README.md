# neotest-go bugs

## JSON output

- Original issue: https://github.com/nvim-neotest/neotest-go/issues/52
- See [jsonoutput_test.go](jsonoutput_test.go) for example code. I left the `t.Parallel()` in, as this is how my production tests are written. But this is not required to trigger the JSON output.

<img width="3390" alt="Screenshot" src="https://github.com/fredrikaverpil/go-playground/assets/994357/db90a22e-da8a-4768-9d96-3d0e0520b46e">

## Crash in `marshal_gotest_output`

- Original issue: https://github.com/nvim-neotest/neotest-go/issues/80
- See [marshaloutput_test.go](marshaloutput_test.go) for example code, see the `// FIXME` comments.
