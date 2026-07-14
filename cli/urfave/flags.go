package urfave

import (
	"context"

	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	clilib "github.com/urfave/cli/v3"
)

// Flags are the standard flags: verbosity and color. [New] prepends them to
// the root command's flags and applies them in the chained Before hook.
type Flags struct {
	Verbose bool
	Quiet   bool
	Color   clog.ColorMode
}

// flags builds the urfave flag definitions bound to f.
func (f *Flags) flags() []clilib.Flag {
	return []clilib.Flag{
		&clilib.BoolFlag{
			Name:        "quiet",
			Aliases:     []string{"q"},
			Usage:       "Only show errors",
			Destination: &f.Quiet,
		},
		&clilib.BoolFlag{
			Name:        "verbose",
			Aliases:     []string{"v"},
			Usage:       "Show debug logs",
			Destination: &f.Verbose,
		},
		&clilib.StringFlag{
			Name:  "color",
			Usage: "When to use color (auto, always, never)",
			Value: "auto",
			Action: func(_ context.Context, _ *clilib.Command, value string) error {
				return f.Color.UnmarshalText([]byte(value))
			},
		},
	}
}

// ConductorFlags implements [conductor.FlagSource].
func (f *Flags) ConductorFlags() conductor.Flags {
	return conductor.Flags{Verbose: f.Verbose, Quiet: f.Quiet, Color: f.Color}
}
