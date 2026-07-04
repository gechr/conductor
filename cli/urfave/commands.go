package urfave

import (
	"context"

	"github.com/gechr/conductor"
	"github.com/gechr/conductor/internal/update"
	clilib "github.com/urfave/cli/v3"
)

// VersionCommand returns the standard `version` subcommand.
func VersionCommand(app *conductor.Runtime) *clilib.Command {
	var detailed bool
	return &clilib.Command{
		Name:  "version",
		Usage: "Print version information",
		Flags: []clilib.Flag{
			&clilib.BoolFlag{
				Name:        "detailed",
				Aliases:     []string{"d"},
				Usage:       "Show detailed build information",
				Destination: &detailed,
			},
		},
		Action: func(context.Context, *clilib.Command) error {
			app.PrintVersion(detailed)
			return nil
		},
	}
}

// UpdateCommand returns the standard self-update subcommand. It dispatches
// on App.Updater's install method (brew, goinstall or github); tools
// distributed another way provide their own update command instead.
func UpdateCommand(app *conductor.Runtime) *clilib.Command {
	var opts update.Options
	return &clilib.Command{
		Name:  "update",
		Usage: "Update " + app.App.DisplayName + " to the latest version",
		Flags: []clilib.Flag{
			&clilib.BoolFlag{
				Name:        "check",
				Usage:       "Report whether an update is available without installing",
				Destination: &opts.Check,
			},
			&clilib.BoolFlag{
				Name:        "no-uninstall",
				Usage:       "Keep a conflicting installation in place",
				Destination: &opts.NoUninstall,
			},
		},
		MutuallyExclusiveFlags: []clilib.MutuallyExclusiveFlags{
			{Flags: [][]clilib.Flag{
				{&clilib.BoolFlag{
					Name:        "stable",
					Usage:       "Install the latest stable release",
					Destination: &opts.Stable,
				}},
				{&clilib.BoolFlag{
					Name:        "dev",
					Usage:       "Install the latest source build",
					Destination: &opts.Dev,
				}},
			}},
		},
		Action: func(ctx context.Context, _ *clilib.Command) error {
			return update.Run(ctx, app, opts)
		},
	}
}
