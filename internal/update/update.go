// Package update holds the self-update logic shared by the framework
// adapters' update commands. It dispatches on the install method configured
// in App.Updater: Homebrew, `go install`, or GitHub release download.
package update

import (
	"context"
	"fmt"

	"github.com/gechr/clive/updater/brew"
	"github.com/gechr/clive/updater/github"
	"github.com/gechr/clive/updater/goinstall"
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

// Run checks for or installs an update using the install method of the
// runtime's App.Updater: [brew.Config], [goinstall.Config] or
// [github.Config]. Tools distributed another way provide their own update
// command instead.
func Run(ctx context.Context, rt *conductor.Runtime, opts Options) error {
	switch cfg := rt.App.Updater.(type) {
	case brew.Config:
		if opts.NoUninstall {
			brew.WithOnConflict(brew.ConflictIgnore)(&cfg)
		}
		if opts.Check {
			return brew.Check(ctx, cfg)
		}
		return brew.Update(ctx, cfg, brew.ChannelFor(opts.Dev, opts.Stable))
	case goinstall.Config:
		if opts.Check {
			return goinstall.Check(ctx, cfg)
		}
		return goinstall.Update(ctx, cfg, goinstall.ChannelFor(opts.Dev))
	case github.Config:
		if opts.Check {
			return github.Check(ctx, cfg)
		}
		return github.Update(ctx, cfg, github.ChannelFor(opts.Dev))
	default:
		return fmt.Errorf(
			"update command supports brew, goinstall and github updaters, got %T",
			rt.App.Updater,
		)
	}
}
