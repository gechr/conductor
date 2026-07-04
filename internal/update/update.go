// Package update holds the self-update logic shared by the framework
// adapters' update commands, driven through the [updater.Updater] interface
// so any install method (Homebrew, `go install`, GitHub release) works.
package update

import (
	"context"
	"errors"
	"fmt"

	"github.com/gechr/clive/updater"
	"github.com/gechr/clive/updater/brew"
	"github.com/gechr/conductor"
)

// Options are the standard update-command flags. Stable and NoUninstall only
// influence the Homebrew method; the other methods treat stable as the
// default channel and manage no conflicting copies.
type Options struct {
	Check       bool
	Stable      bool
	Dev         bool
	NoUninstall bool
}

// Requested reports whether args contains the hidden --self-update flag,
// scanning only up to a "--" terminator. Like completion preflight, the scan
// runs before parsing so self-update works even in grammars whose validation
// would otherwise demand arguments or a subcommand. The flag is mutually
// exclusive with everything else: when it appears alongside any other
// argument, the first result is still true and the error explains the
// conflict.
func Requested(args []string) (bool, error) {
	requested := false
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "--self-update" {
			requested = true
		}
	}
	if requested && len(args) > 1 {
		return true, errors.New("--self-update cannot be combined with other arguments")
	}
	return requested, nil
}

// Run checks for or installs an update via App.Updater, which must implement
// [updater.Updater] (every clive updater config does). Tools distributed
// another way provide their own update command instead.
func Run(ctx context.Context, app *conductor.Runtime, opts Options) error {
	u, ok := app.App.Updater.(updater.Updater)
	if !ok {
		return fmt.Errorf(
			"update command requires a self-updating updater, got %T",
			app.App.Updater,
		)
	}
	// Leaving a conflicting install in place is inherently Homebrew-specific.
	if cfg, ok := u.(brew.Config); ok && opts.NoUninstall {
		brew.WithOnConflict(brew.ConflictIgnore)(&cfg)
		u = cfg
	}
	if opts.Check {
		return u.Check(ctx)
	}
	return u.Update(ctx, opts.Dev, opts.Stable)
}
