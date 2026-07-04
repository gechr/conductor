// Package conductor wires up the gechr CLI libraries (clib, clog, clive) in
// one shot. A consumer fills in an [App] at program start, calls [New], and
// inherits the house logging style, a themed help renderer, unified env-var
// prefixes, version plumbing, and passive update notifications - while
// remaining free to call any of the underlying libraries directly.
package conductor

import (
	"strings"

	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	"github.com/gechr/clive"
	"github.com/gechr/clive/notify"
	"github.com/gechr/clive/updater"
	"github.com/gechr/clog"
	xstrings "github.com/gechr/x/strings"
)

// EnvPrefixNone disables env-prefix unification, leaving clog and clib on
// their own defaults (CLOG and CLIB).
const EnvPrefixNone = "-"

// App declares everything Conductor needs to initialize a CLI tool. Name is
// required; every other field has a sensible zero value.
type App struct {
	// Name is the binary name, e.g. "demo". Required.
	Name string
	// Description is the one-line description used in help output.
	Description string

	// EnvPrefix unifies clog.SetEnvPrefix and clib theme.SetEnvPrefix under a
	// single application prefix. It defaults to Name upper-cased with dashes
	// mapped to underscores. Set [EnvPrefixNone] to leave both libraries on
	// their own defaults.
	EnvPrefix string

	// Module is the Go module path, e.g. "github.com/gechr/demo". It enables
	// clive release links and latest-version lookups.
	Module string
	// Repo is the GitHub "owner/name"; derived from Module when empty.
	Repo string
	// Private resolves latest-version lookups via GOPRIVATE-scoped version
	// control instead of the public module proxy.
	Private bool

	// Updater enables the passive update notification; nil disables it. Any
	// [updater.Tool] works, e.g. brew.New(...), goinstall.New(...) or
	// github.New(...).
	Updater updater.Tool
	// NotifyOptions are appended after Conductor's own notify options.
	NotifyOptions []notify.Option
	// NotifySkip lists extra command verbs that suppress the update check, in
	// addition to the built-in "update", "version", "help" and completion
	// verbs.
	NotifySkip []string
	// NotifyOnly, when non-empty, restricts the update check to exactly these
	// command verbs.
	NotifyOnly []string

	// Theme overrides the help theme; nil means [theme.Auto].
	Theme *theme.Theme
	// HelpOptions are passed to [help.NewRenderer].
	HelpOptions []help.RendererOption
	// HelpShort is the terse description of -h (default "Print short help").
	HelpShort string
	// HelpLong is the terse description of --help (default "Print long
	// help, including examples").
	HelpLong string

	// SkipLogDefaults skips [LogDefaults] for tools with a fully bespoke
	// logging setup.
	SkipLogDefaults bool
	// ConfigureLog runs after [LogDefaults] and before anything logs; the
	// place for direct clog customisation (symbols, styles, custom levels).
	ConfigureLog func()
}

// Runtime is everything [New] builds. All fields are exported and may be
// replaced before handing the Runtime to a framework adapter.
type Runtime struct {
	App      App
	Info     clive.Info
	Theme    *theme.Theme
	Renderer *help.Renderer
}

// New initializes the libraries in dependency order and returns the shared
// Runtime. It must run before anything logs so that env prefixes, output and
// theme detection settle first. It panics when app.Name is empty - a
// programmer error on par with an invalid CLI grammar.
func New(app App) *Runtime {
	if xstrings.IsBlank(app.Name) {
		panic("conductor: App.Name is required")
	}

	// clog re-reads the environment when the prefix changes, so this must
	// precede LogDefaults for <PREFIX>_* overrides to survive.
	if prefix := app.envPrefix(); prefix != EnvPrefixNone {
		clog.SetEnvPrefix(prefix)
		theme.SetEnvPrefix(prefix)
	}

	if !app.SkipLogDefaults {
		LogDefaults()
	}
	if app.ConfigureLog != nil {
		app.ConfigureLog()
	}

	th := app.Theme
	if th == nil {
		th = theme.Auto()
	}

	return &Runtime{
		App:      app,
		Info:     clive.Info{Module: app.Module, Repo: app.Repo, Private: app.Private},
		Theme:    th,
		Renderer: help.NewRenderer(th, app.HelpOptions...),
	}
}

// HelpShortDesc returns the terse -h description, defaulted.
func (a App) HelpShortDesc() string {
	if a.HelpShort != "" {
		return a.HelpShort
	}
	return "Print short help"
}

// HelpLongDesc returns the terse --help description, defaulted.
func (a App) HelpLongDesc() string {
	if a.HelpLong != "" {
		return a.HelpLong
	}
	return "Print long help, including examples"
}

// envPrefix resolves the effective prefix: EnvPrefix as given, or Name
// upper-cased with dashes mapped to underscores.
func (a App) envPrefix() string {
	if a.EnvPrefix != "" {
		return a.EnvPrefix
	}
	return strings.ToUpper(strings.ReplaceAll(a.Name, "-", "_"))
}
