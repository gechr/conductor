// Package kong adapts Conductor to kong-based CLIs: one call builds the
// parser with themed help, shell completions, version wiring and update
// notifications, mirroring the conventions of clib's cli/kong package.
package kong

import (
	"errors"

	konglib "github.com/alecthomas/kong"
	clibkong "github.com/gechr/clib/cli/kong"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/help"
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
)

// Program is the assembled kong CLI. All fields are exported and may be
// customised between [New] and [Program.Run].
type Program struct {
	Runtime *conductor.Runtime
	Parser  *konglib.Kong
	Gen     *complete.Generator

	cli any
	cfg config
}

type config struct {
	kongOptions  []konglib.Option
	nodeSections []clibkong.NodeSectionsOption
	generator    []func(*complete.Generator)
	exitCode     func(error) int
	noDispatch   bool
}

// Option configures [New].
type Option func(*config)

// WithKongOptions appends extra kong options after Conductor's own, so they
// can override any default.
func WithKongOptions(opts ...konglib.Option) Option {
	return func(c *config) { c.kongOptions = append(c.kongOptions, opts...) }
}

// WithNodeSections configures the help section builder.
func WithNodeSections(opts ...clibkong.NodeSectionsOption) Option {
	return func(c *config) { c.nodeSections = append(c.nodeSections, opts...) }
}

// WithBind binds extra values into the kong context for command Run methods.
func WithBind(values ...any) Option {
	return WithKongOptions(konglib.Bind(values...))
}

// WithGenerator customises the completion generator before the parser is
// built, e.g. to add dynamic argument specs.
func WithGenerator(fn func(*complete.Generator)) Option {
	return func(c *config) { c.generator = append(c.generator, fn) }
}

// WithExitCode maps command errors to exit codes; see [conductor.ExitCode].
func WithExitCode(fn func(error) int) Option {
	return func(c *config) { c.exitCode = fn }
}

// WithNoDefaultCommand skips kong dispatch in [Program.Run] for flag-only
// CLIs that inspect the parse result themselves.
func WithNoDefaultCommand() Option {
	return func(c *config) { c.noDispatch = true }
}

// New builds the parser, completion generator and help wiring for cli, a
// pointer to a kong grammar struct (usually embedding [Flags]).
func New(app *conductor.Runtime, cli any, opts ...Option) (*Program, error) {
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}

	flags, err := clibkong.Reflect(cli)
	if err != nil {
		return nil, err
	}
	gen := complete.NewGenerator(app.App.Name).FromFlags(flags)
	gen.Specs = append(gen.Specs,
		complete.Spec{ShortFlag: "h", Terse: app.App.HelpShortDesc()},
		complete.Spec{LongFlag: "help", Terse: app.App.HelpLongDesc()},
	)
	for _, fn := range cfg.generator {
		fn(gen)
	}

	kongOpts := []konglib.Option{
		konglib.Name(app.App.Name),
		konglib.Description(app.App.Description),
		konglib.UsageOnError(),
		konglib.Vars{"version": app.Version()},
		konglib.Help(clibkong.HelpPrinterFunc(
			app.Renderer,
			clibkong.NodeSectionsFunc(cfg.nodeSections...),
			help.WithHelpFlags(app.App.HelpShortDesc(), app.App.HelpLongDesc()),
		)),
		konglib.Bind(app),
		konglib.Bind(gen),
	}
	kongOpts = append(kongOpts, cfg.kongOptions...)

	parser, err := konglib.New(cli, kongOpts...)
	if err != nil {
		return nil, err
	}
	// Populate subcommand specs from the parser model so completion scripts
	// list subcommands and their flags.
	gen.Subs = clibkong.Subcommands(parser)

	return &Program{Runtime: app, Parser: parser, Gen: gen, cli: cli, cfg: cfg}, nil
}

// Run is the one-call happy path: completion preflight, parse, standard flag
// application, update notification, dispatch, deferred hint flush, and exit
// code mapping. The caller passes the result to os.Exit.
func (p *Program) Run(args []string) int {
	if handled, code := p.completion(); handled {
		return code
	}

	kctx, err := p.Parser.Parse(args)
	if err != nil {
		// Report via kong (error plus usage-on-error) but return the exit
		// code rather than letting kong terminate the process.
		exit := p.Parser.Exit
		p.Parser.Exit = func(int) {}
		p.Parser.FatalIfErrorf(err)
		p.Parser.Exit = exit
		if parseErr, ok := errors.AsType[*konglib.ParseError](err); ok {
			return parseErr.ExitCode()
		}
		return 1
	}

	if src, ok := p.cli.(conductor.FlagSource); ok {
		p.Runtime.ApplyFlags(src.ConductorFlags())
	}

	flush := p.Runtime.Notify(kctx.Command())
	var runErr error
	if !p.cfg.noDispatch {
		runErr = kctx.Run()
	}
	flush()
	return conductor.ExitCode(runErr, p.cfg.exitCode)
}

// Parse is the granular path for tools that dispatch themselves: it handles
// completion preflight (the bool result reports a completion action ran and
// the caller should exit with the int code) and parses args, leaving flag
// application, notify and dispatch to the caller.
func (p *Program) Parse(args []string) (*konglib.Context, bool, int, error) {
	if handled, code := p.completion(); handled {
		return nil, true, code, nil
	}
	kctx, err := p.Parser.Parse(args)
	return kctx, false, 0, err
}

// completion executes a pre-parse completion action when one was requested.
func (p *Program) completion() (bool, int) {
	cf, args, ok := clibkong.Preflight()
	if !ok {
		return false, 0
	}
	if _, err := cf.Handle(p.Gen, nil, clibkong.WithArgs(args)); err != nil {
		clog.Error().Err(err).Msg("Completion failed")
		return true, 1
	}
	return true, 0
}
