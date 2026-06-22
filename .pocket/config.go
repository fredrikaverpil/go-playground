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
			// The generics/ directory is a hand-written tutorial; skip go-fix
			// there so modernizers (e.g. slicescontains) don't rewrite examples
			// that intentionally spell out == loops, etc.
			pk.WithSkipTask(golang.Fix, "^generics$"),
		),
		markdown.Tasks(),
		pk.WithOptions(
			github.Tasks(),
			pk.WithFlags(github.WorkflowFlags{
				PerPocketTaskJob:      new(true),
				ReleasePleaseWorkflow: new(false),
				StaleWorkflow:         new(false),
				Platforms:             []github.Platform{github.Ubuntu},
			}),
		),
	),
}
