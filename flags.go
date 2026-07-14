package conductor

import (
	"fmt"

	"github.com/gechr/clog"
)

// The -v counter steps: 1 is debug, 2 is trace (the lowest clog level).
// maxVerbosity leaves one step of headroom above trace; anything beyond is a
// typo worth rejecting rather than silently clamping.
const (
	minVerbosity   = iota // 0: default level
	verbosityDebug        // 1: -v
	verbosityTrace        // 2: -vv
	maxVerbosity          // 3: headroom above trace
)

// Flags are the framework-neutral post-parse flag values shared by every
// tool: verbosity and color mode.
type Flags struct {
	// Verbose is the legacy boolean debug switch, kept for tools that have
	// not moved to the Verbosity counter. Verbosity takes precedence when
	// non-zero.
	Verbose bool
	// Verbosity is the -v counter: 0 leaves the default level, 1 is debug,
	// 2 or more is trace. Adapters map their repeatable -v flag onto it.
	Verbosity int
	Quiet     bool
	Color     clog.ColorMode
}

// FlagSource is implemented by CLI values that expose the standard flags,
// usually by embedding a framework adapter's flag struct. Adapters apply the
// flags automatically after parsing when the CLI value implements it.
type FlagSource interface {
	ConductorFlags() Flags
}

// SelfUpdater is implemented by CLI grammars that opt into the hidden
// --self-update flag, usually by embedding an adapter's SelfUpdateFlag.
type SelfUpdater interface {
	SelfUpdateRequested() bool
}

// ApplyFlags is the post-parse phase: it maps the standard flags onto clog.
// Tools with bespoke verbosity semantics skip FlagSource and call clog
// directly instead. It errors on an out-of-range verbosity so a typo like
// -vvvv fails loudly rather than resolving to trace.
func (r *Runtime) ApplyFlags(f Flags) error {
	if f.Verbosity < minVerbosity || f.Verbosity > maxVerbosity {
		return fmt.Errorf(
			"verbosity %d out of range (0-%d) - use -v for debug or -vv for trace",
			f.Verbosity, maxVerbosity,
		)
	}
	clog.SetColorMode(f.Color)
	switch {
	case f.Quiet:
		clog.SetLevel(clog.LevelError)
	case f.Verbosity >= verbosityTrace:
		clog.SetLevel(clog.LevelTrace)
		clog.SetReportTimestamp(true)
	case f.Verbosity == verbosityDebug || f.Verbose:
		clog.SetVerbose(true)
	default:
		clog.SetVerbose(false)
	}
	return nil
}
