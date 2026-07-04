package conductor

import "github.com/gechr/clog"

// Flags are the framework-neutral post-parse flag values shared by every
// tool: verbosity and color mode.
type Flags struct {
	Verbose bool
	Quiet   bool
	Color   clog.ColorMode
}

// FlagSource is implemented by CLI values that expose the standard flags,
// usually by embedding a framework adapter's flag struct. Adapters apply the
// flags automatically after parsing when the CLI value implements it.
type FlagSource interface {
	ConductorFlags() Flags
}

// ApplyFlags is the post-parse phase: it maps the standard flags onto clog.
// Tools with bespoke verbosity semantics skip FlagSource and call clog
// directly instead.
func (r *Runtime) ApplyFlags(f Flags) {
	clog.SetColorMode(f.Color)
	clog.SetVerbose(f.Verbose)
	if f.Quiet {
		clog.SetLevel(clog.LevelError)
	}
}
