package main

import (
	"github.com/fredrikaverpil/pocket/pk"
	"github.com/fredrikaverpil/pocket/tasks/golang"
)

// Config is the Pocket configuration for this project.
// Edit this file to define your tasks and composition.
var Config = &pk.Config{
	Auto: pk.Parallel(
		pk.WithOptions(
			golang.Tasks(),
			pk.WithDetect(golang.Detect()),
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
