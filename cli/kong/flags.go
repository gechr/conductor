package kong

import (
	konglib "github.com/alecthomas/kong"
	clibkong "github.com/gechr/clib/cli/kong"
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
)

// Flags is the standard embeddable flag block: shell-completion management
// plus the verbosity/color/version trio every tool shares. Embedding it makes
// the CLI a [conductor.FlagSource], so [Program.Run] applies the flags to
// clog automatically after parsing.
type Flags struct {
	clibkong.CompletionFlags

	Verbose int            `help:"Increase log verbosity" short:"v"        type:"counter"  xor:"verbosity"`
	Quiet   bool           `help:"Only show errors"       short:"q"        xor:"verbosity"`
	Color   clog.ColorMode `help:"When to use color"      aliases:"colour" default:"auto"  enum:"auto,always,never"`

	VersionFlag konglib.VersionFlag `name:"version" help:"Print version information" short:"V" hidden:""`
}

// ConductorFlags implements [conductor.FlagSource].
func (f Flags) ConductorFlags() conductor.Flags {
	return conductor.Flags{Verbosity: f.Verbose, Quiet: f.Quiet, Color: f.Color}
}

// SelfUpdateFlag is an embeddable hidden --self-update flag for flag-only
// CLIs that have no subcommands to hang an update command off. When set,
// [Program.Run] performs the self-update via App.Updater instead of
// dispatching, and exits.
type SelfUpdateFlag struct {
	SelfUpdate bool `name:"self-update" help:"Update to the latest version" hidden:"" clib:"complete-hidden"`
}

// SelfUpdateRequested reports whether the self-update flag was set.
func (f SelfUpdateFlag) SelfUpdateRequested() bool { return f.SelfUpdate }
