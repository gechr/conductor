package cobra

import (
	"github.com/gechr/conductor"
	"github.com/gechr/conductor/internal/update"
	cobralib "github.com/spf13/cobra"
)

// VersionCommand returns the standard `version` subcommand.
func VersionCommand(rt *conductor.Runtime) *cobralib.Command {
	var detailed bool
	cmd := &cobralib.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobralib.NoArgs,
		RunE: func(*cobralib.Command, []string) error {
			rt.PrintVersion(detailed)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed build information")
	return cmd
}

// UpdateCommand returns the standard self-update subcommand. It dispatches
// on App.Updater's install method (brew, goinstall or github); tools
// distributed another way provide their own update command instead.
func UpdateCommand(rt *conductor.Runtime) *cobralib.Command {
	var opts update.Options
	cmd := &cobralib.Command{
		Use:   "update",
		Short: "Update " + rt.App.Name + " to the latest version",
		Args:  cobralib.NoArgs,
		RunE: func(cmd *cobralib.Command, _ []string) error {
			return update.Run(cmd.Context(), rt, opts)
		},
	}
	fs := cmd.Flags()
	fs.BoolVar(
		&opts.Check,
		"check",
		false,
		"Report whether an update is available without installing",
	)
	fs.BoolVar(&opts.Stable, "stable", false, "Install the latest stable release")
	fs.BoolVar(&opts.Dev, "dev", false, "Install the latest source build")
	fs.BoolVar(&opts.NoUninstall, "no-uninstall", false, "Keep a conflicting installation in place")
	cmd.MarkFlagsMutuallyExclusive("stable", "dev")
	return cmd
}
