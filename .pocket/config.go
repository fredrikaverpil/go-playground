package main

import (
	"github.com/fredrikaverpil/pocket/pk"
	"github.com/fredrikaverpil/pocket/tasks/github"
	"github.com/fredrikaverpil/pocket/tasks/golang"
	"github.com/fredrikaverpil/pocket/tasks/markdown"
)

// Config is the Pocket configuration for this project.
// Edit this file to define your tasks and composition.
var Config = &pk.Config{
	Auto: pk.Parallel(
		pk.WithOptions(
			golang.Tasks(),
			pk.WithDetect(golang.Detect()),
		),
		markdown.Tasks(),
		pk.WithOptions(
			github.Tasks(),
			pk.WithFlag(github.Workflows, "skip-pocket", true),
			pk.WithFlag(github.Workflows, "skip-release", true),
			pk.WithFlag(github.Workflows, "skip-stale", true),
			pk.WithFlag(github.Workflows, "include-pocket-matrix", true),
			pk.WithContextValue(github.MatrixConfigKey{}, github.MatrixConfig{
				DefaultPlatforms: []string{"ubuntu-latest"},
			}),
		),
	),

	// Plan configuration: shims, directories, and CI settings.
	// Use ./pok -g to run git diff check after execution.
	Plan: &pk.PlanConfig{
		Shims: &pk.ShimConfig{
			Posix: true,
		},
	},
}
