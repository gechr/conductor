// Package urfave adapts conductor to urfave/cli v3 CLIs: one call wires
// themed help, shell completions, standard flags, version output and update
// notifications onto an existing root command, mirroring the conventions of
// clib's cli/urfave package.
package urfave

import (
	"context"

	cliburfave "github.com/gechr/clib/cli/urfave"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/help"
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	clilib "github.com/urfave/cli/v3"
)

// Program is the assembled urfave CLI. All fields are exported and may be
// customised between [New] and [Program.Run].
type Program struct {
	Runtime    *conductor.Runtime
	Root       *clilib.Command
	Completion *cliburfave.Completion
	Flags      *Flags

	cfg   config
	flush func()
}

type config struct {
	sections  []cliburfave.SectionsOption
	generator []func(*complete.Generator)
	exitCode  func(error) int
}

// Option configures [New].
type Option func(*Program)

// WithSections configures the help section builder.
func WithSections(opts ...cliburfave.SectionsOption) Option {
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
	return func(p *Program) { p.Root.Commands = append(p.Root.Commands, VersionCommand(p.Runtime)) }
}

// WithUpdateCommand adds the standard Homebrew `update` subcommand.
func WithUpdateCommand() Option {
	return func(p *Program) { p.Root.Commands = append(p.Root.Commands, UpdateCommand(p.Runtime)) }
}

// New wires conductor onto root: identity defaults from the App, themed help
// (via the package-global [clilib.HelpPrinter]), prepended standard flags, a
// -V/--version flag (replacing urfave's -v alias, which the verbose flag
// claims), completion flags, and a chained Before hook that applies the
// standard flags and starts the update notification.
func New(app *conductor.Runtime, root *clilib.Command, opts ...Option) *Program {
	if root.Name == "" {
		root.Name = app.App.Name
	}
	if root.Usage == "" {
		root.Usage = app.App.Description
	}
	root.Version = app.Version()
	// Print the bare version, matching the kong and cobra adapters.
	clilib.VersionPrinter = func(*clilib.Command) { app.PrintVersion(false) }
	// urfave's default version flag aliases -v, which --verbose claims.
	clilib.VersionFlag = &clilib.BoolFlag{
		Name:        "version",
		Aliases:     []string{"V"},
		Usage:       "Print version information",
		HideDefault: true,
		Local:       true,
	}

	p := &Program{
		Runtime: app,
		Root:    root,
		Flags:   &Flags{},
	}
	root.Flags = append(p.Flags.flags(), root.Flags...)
	p.Completion = cliburfave.NewCompletion(root)

	for _, opt := range opts {
		opt(p)
	}

	clilib.HelpPrinter = cliburfave.HelpPrinter(
		app.Renderer,
		cliburfave.SectionsWithOptions(p.cfg.sections...),
		help.WithHelpFlags(app.App.HelpShortDesc(), app.App.HelpLongDesc()),
	)

	prev := root.Before
	root.Before = func(ctx context.Context, cmd *clilib.Command) (context.Context, error) {
		app.ApplyFlags(p.Flags.ConductorFlags())
		p.flush = app.Notify(cmd.Args().First())
		if prev != nil {
			return prev(ctx, cmd)
		}
		return ctx, nil
	}

	return p
}

// Run is the one-call happy path: completion preflight, run, deferred
// update-hint flush, and exit code mapping. args is the full argument vector
// including the program name (os.Args, urfave's convention). The caller
// passes the result to os.Exit.
func (p *Program) Run(ctx context.Context, args []string) int {
	if handled, code := p.completion(); handled {
		return code
	}

	err := p.Root.Run(ctx, args)
	if p.flush != nil {
		p.flush()
	}
	return conductor.ExitCode(err, p.cfg.exitCode)
}

// Generator builds the completion generator from the assembled command tree.
func (p *Program) Generator() *complete.Generator {
	gen := complete.NewGenerator(p.Runtime.App.Name).FromFlags(cliburfave.FlagMeta(p.Root))
	gen.Specs = append(gen.Specs,
		complete.Spec{ShortFlag: "h", Terse: p.Runtime.App.HelpShortDesc()},
		complete.Spec{LongFlag: "help", Terse: p.Runtime.App.HelpLongDesc()},
	)
	gen.Subs = cliburfave.Subcommands(p.Root)
	for _, fn := range p.cfg.generator {
		fn(gen)
	}
	return gen
}

// completion executes a pre-parse completion action when one was requested.
func (p *Program) completion() (bool, int) {
	cf, args, ok := cliburfave.Preflight()
	if !ok {
		return false, 0
	}
	if _, err := cf.Handle(p.Generator(), nil, complete.WithArgs(args)); err != nil {
		clog.Error().Err(err).Msg("Completion failed")
		return true, 1
	}
	return true, 0
}
