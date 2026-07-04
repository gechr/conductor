// Package cobra adapts Conductor to cobra-based CLIs: one call wires themed
// help, shell completions, standard flags, version output and update
// notifications onto an existing root command, mirroring the conventions of
// clib's cli/cobra package.
package cobra

import (
	"context"
	"strings"

	clibcobra "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/help"
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	"github.com/gechr/conductor/internal/update"
	cobralib "github.com/spf13/cobra"
)

// Program is the assembled cobra CLI. All fields are exported and may be
// customised between [New] and [Program.Run].
type Program struct {
	Runtime    *conductor.Runtime
	Root       *cobralib.Command
	Completion *clibcobra.Completion
	Flags      *Flags

	cfg   config
	flush func()
}

type config struct {
	sections        []clibcobra.SectionsOption
	generator       []func(*complete.Generator)
	exitCode        func(error) int
	selfUpdate      bool
	noStandardFlags bool
}

// Option configures [New].
type Option func(*Program)

// WithSections configures the help section builder.
func WithSections(opts ...clibcobra.SectionsOption) Option {
	return func(p *Program) { p.cfg.sections = append(p.cfg.sections, opts...) }
}

// WithGenerator customises the completion generator before completion
// actions run, e.g. to add dynamic argument specs.
func WithGenerator(fn func(*complete.Generator)) Option {
	return func(p *Program) { p.cfg.generator = append(p.cfg.generator, fn) }
}

// WithExitCode maps command errors to exit codes; see [conductor.ExitCode].
func WithExitCode(fn func(error) int) Option {
	return func(p *Program) { p.cfg.exitCode = fn }
}

// WithVersionCommand adds the standard `version` subcommand.
func WithVersionCommand() Option {
	return func(p *Program) { p.Root.AddCommand(VersionCommand(p.Runtime)) }
}

// WithUpdateCommand adds the standard Homebrew `update` subcommand.
func WithUpdateCommand() Option {
	return func(p *Program) { p.Root.AddCommand(UpdateCommand(p.Runtime)) }
}

// WithoutStandardFlags skips registering --verbose/--quiet/--color, for
// tools whose own flags or output policy already cover those semantics (a
// subcommand's local shorthand may also conflict with the persistent ones).
func WithoutStandardFlags() Option {
	return func(p *Program) { p.cfg.noStandardFlags = true }
}

// WithSelfUpdate adds the hidden --self-update flag for CLIs without
// subcommands, which have no room for an update command. [Program.Run]
// intercepts the flag before cobra executes and performs the self-update via
// App.Updater; it is mutually exclusive with every other argument.
func WithSelfUpdate() Option {
	return func(p *Program) {
		p.cfg.selfUpdate = true
		fs := p.Root.Flags()
		fs.Bool("self-update", false, "Update to the latest version")
		_ = fs.MarkHidden("self-update")
	}
}

// New wires Conductor onto root: identity defaults from the App, themed help,
// persistent standard flags, a -V/--version flag, completion flags, and a
// chained PersistentPreRunE that applies the standard flags and starts the
// update notification once the command is resolved. Add subcommands to root
// before or after New; the completion generator is built at Run time.
func New(app *conductor.Runtime, root *cobralib.Command, opts ...Option) *Program {
	if root.Use == "" {
		root.Use = app.App.Name
	}
	if root.Short == "" {
		root.Short = app.App.Description
	}
	root.SilenceUsage = true
	root.SilenceErrors = true
	root.Version = app.Version()
	root.SetVersionTemplate("{{.Version}}\n")
	// Pre-register the version flag so cobra's lazy init keeps our -V
	// shorthand and help text.
	root.Flags().BoolP("version", "V", false, "Print version information")

	root.SetHelpFunc(clibcobra.HelpFunc(
		app.Renderer,
		clibcobra.SectionsWithOptions(),
		help.WithHelpFlags(app.App.HelpShortDesc(), app.App.HelpLongDesc()),
	))

	p := &Program{
		Runtime:    app,
		Root:       root,
		Completion: clibcobra.NewCompletion(root),
		Flags:      &Flags{},
	}

	for _, opt := range opts {
		opt(p)
	}
	if !p.cfg.noStandardFlags {
		p.Flags.Register(root.PersistentFlags())
	}
	if len(p.cfg.sections) > 0 {
		root.SetHelpFunc(clibcobra.HelpFunc(
			app.Renderer,
			clibcobra.SectionsWithOptions(p.cfg.sections...),
			help.WithHelpFlags(app.App.HelpShortDesc(), app.App.HelpLongDesc()),
		))
	}

	prev := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobralib.Command, args []string) error {
		app.ApplyFlags(p.Flags.ConductorFlags())
		p.flush = app.Notify(commandVerb(root, cmd))
		if prev != nil {
			return prev(cmd, args)
		}
		return nil
	}

	return p
}

// Run is the one-call happy path: completion preflight, execute, deferred
// update-hint flush, and exit code mapping. The caller passes the result to
// os.Exit.
func (p *Program) Run(ctx context.Context, args []string) int {
	if handled, code := p.completion(); handled {
		return code
	}

	// Self-update runs before cobra executes, like completion; the flag is
	// mutually exclusive with every other argument.
	if p.cfg.selfUpdate {
		if requested, err := update.Requested(args); requested {
			if err == nil {
				err = update.Run(ctx, p.Runtime, update.Options{})
			}
			return conductor.ExitCode(err, p.cfg.exitCode)
		}
	}

	p.Root.SetArgs(args)
	err := p.Root.ExecuteContext(ctx)
	if p.flush != nil {
		p.flush()
	}
	return conductor.ExitCode(err, p.cfg.exitCode)
}

// Generator builds the completion generator from the assembled command tree.
func (p *Program) Generator() *complete.Generator {
	gen := complete.NewGenerator(p.Runtime.App.Name).FromFlags(clibcobra.FlagMeta(p.Root))
	gen.Specs = append(gen.Specs,
		complete.Spec{ShortFlag: "h", Terse: p.Runtime.App.HelpShortDesc()},
		complete.Spec{LongFlag: "help", Terse: p.Runtime.App.HelpLongDesc()},
	)
	gen.Subs = clibcobra.Subcommands(p.Root)
	for _, fn := range p.cfg.generator {
		fn(gen)
	}
	return gen
}

// completion executes a pre-parse completion action when one was requested.
func (p *Program) completion() (bool, int) {
	cf, args, ok := clibcobra.Preflight()
	if !ok {
		return false, 0
	}
	if _, err := cf.Handle(p.Generator(), nil, complete.WithArgs(args)); err != nil {
		clog.Error().Err(err).Msg("Completion failed")
		return true, 1
	}
	return true, 0
}

// commandVerb returns the first subcommand segment of cmd's path, the verb
// Conductor's notify skip logic keys on.
func commandVerb(root, cmd *cobralib.Command) string {
	path := strings.TrimPrefix(cmd.CommandPath(), root.Name())
	return strings.TrimSpace(path)
}
