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

// Register adds --verbose/-v, --quiet/-q and --color to fs. Flags the
// command already defines are left untouched - a tool with its own --quiet
// semantics keeps them, and the corresponding Flags field simply stays zero.
// A shorthand already claimed by another flag registers the long form only.
func (f *Flags) Register(fs *pflag.FlagSet) {
	registerBool := func(target *bool, name, shorthand, usage string) {
		if fs.Lookup(name) != nil {
			return
		}
		if fs.ShorthandLookup(shorthand) != nil {
			shorthand = ""
		}
		fs.BoolVarP(target, name, shorthand, false, usage)
	}
	registerBool(&f.Verbose, "verbose", "v", "Show debug logs")
	registerBool(&f.Quiet, "quiet", "q", "Only show errors")
	if fs.Lookup("color") == nil {
		fs.TextVar(&f.Color, "color", clog.ColorAuto, "When to use color (auto, always, never)")
	}
}

// ConductorFlags implements [conductor.FlagSource].
func (f *Flags) ConductorFlags() conductor.Flags {
	return conductor.Flags{Verbose: f.Verbose, Quiet: f.Quiet, Color: f.Color}
}
