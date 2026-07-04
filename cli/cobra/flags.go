package cobra

import (
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	"github.com/spf13/pflag"
)

// Flags are the standard persistent flags: verbosity and color. [New]
// registers them on the root command's persistent flag set and applies them
// in the chained PersistentPreRunE.
type Flags struct {
	Verbose bool
	Quiet   bool
	Color   clog.ColorMode
}

// Register adds --verbose/-v, --quiet/-q and --color to fs.
func (f *Flags) Register(fs *pflag.FlagSet) {
	fs.BoolVarP(&f.Verbose, "verbose", "v", false, "Show debug logs")
	fs.BoolVarP(&f.Quiet, "quiet", "q", false, "Only show errors")
	fs.TextVar(&f.Color, "color", clog.ColorAuto, "When to use color (auto, always, never)")
}

// ConductorFlags implements [conductor.FlagSource].
func (f *Flags) ConductorFlags() conductor.Flags {
	return conductor.Flags{Verbose: f.Verbose, Quiet: f.Quiet, Color: f.Color}
}
