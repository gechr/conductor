package kong

import (
	"context"

	"github.com/gechr/conductor"
	"github.com/gechr/conductor/internal/update"
)

// VersionCmd is an embeddable `version` subcommand:
//
//	Version kong.VersionCmd `cmd:"" help:"Print version information"`
type VersionCmd struct {
	Detailed bool `help:"Show detailed build information" short:"d"`
}

// Run prints the version via the bound [conductor.Runtime].
func (c *VersionCmd) Run(app *conductor.Runtime) error {
	app.PrintVersion(c.Detailed)
	return nil
}

// UpdateCmd is an embeddable self-update subcommand:
//
//	Update kong.UpdateCmd `cmd:"" help:"Update <app> to the latest version"`
//
// It dispatches on App.Updater's install method (brew, goinstall or github);
// tools distributed another way provide their own update command instead.
type UpdateCmd struct {
	Check       bool `help:"Report whether an update is available without installing"`
	Stable      bool `help:"Install the latest stable release"                        clib:"group='Channel/0'" xor:"channel"`
	Dev         bool `help:"Install the latest source build"                          clib:"group='Channel/0'" xor:"channel"`
	NoUninstall bool `help:"Keep a conflicting installation in place"                 clib:"group='Cleanup/0'"`
}

// Run checks for or installs the update.
func (c *UpdateCmd) Run(app *conductor.Runtime) error {
	return update.Run(context.Background(), app, update.Options{
		Check:       c.Check,
		Stable:      c.Stable,
		Dev:         c.Dev,
		NoUninstall: c.NoUninstall,
	})
}
