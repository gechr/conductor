package urfave_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	curfave "github.com/gechr/conductor/cli/urfave"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clilib "github.com/urfave/cli/v3"
)

func newProgram(t *testing.T, root *clilib.Command, opts ...curfave.Option) *curfave.Program {
	t.Helper()
	app := conductor.New(conductor.App{Name: "demo", Description: "Test app."})
	return curfave.New(app, root, opts...)
}

func greetCommand(err error) *clilib.Command {
	return &clilib.Command{
		Name:   "greet",
		Usage:  "Greet",
		Action: func(context.Context, *clilib.Command) error { return err },
	}
}

func TestNewDefaultsIdentity(t *testing.T) {
	root := &clilib.Command{}
	prog := newProgram(t, root)
	assert.Equal(t, "demo", prog.Root.Name)
	assert.Equal(t, "Test app.", prog.Root.Usage)
	assert.True(t, prog.Root.UseShortOptionHandling)
}

func TestRunDispatchesAndAppliesFlags(t *testing.T) {
	t.Cleanup(func() { clog.SetVerbose(false) })

	root := &clilib.Command{Commands: []*clilib.Command{greetCommand(nil)}}
	prog := newProgram(t, root)
	code := prog.Run(context.Background(), []string{"demo", "--verbose", "greet"})
	assert.Equal(t, 0, code)
	assert.True(t, clog.IsVerbose(), "verbose flag should reach clog")
}

func TestRunAppliesRepeatableVerbosity(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "separate short flags", args: []string{"demo", "-v", "-v", "-v", "greet"}},
		{name: "combined short flags", args: []string{"demo", "-vvv", "greet"}},
		{
			name: "long flags",
			args: []string{"demo", "--verbose", "--verbose", "--verbose", "greet"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() { clog.SetLevel(clog.LevelInfo) })

			root := &clilib.Command{Commands: []*clilib.Command{greetCommand(nil)}}
			prog := newProgram(t, root)
			require.Equal(t, 0, prog.Run(context.Background(), tt.args))
			assert.True(t, prog.Flags.Verbose)
			assert.Equal(t, 3, prog.Flags.Verbosity)
			assert.Equal(t, clog.LevelTrace, clog.GetLevel())
		})
	}
}

func TestRunRejectsExcessiveVerbosity(t *testing.T) {
	root := &clilib.Command{Commands: []*clilib.Command{greetCommand(nil)}}
	prog := newProgram(t, root)
	err := prog.Root.Run(context.Background(), []string{"demo", "-vvvv", "greet"})
	require.EqualError(
		t,
		err,
		"verbosity 4 out of range (0-3) - use -v for debug or -vv for trace",
	)
	assert.Equal(t, 4, prog.Flags.Verbosity)
}

func TestVerboseFlagDescription(t *testing.T) {
	root := &clilib.Command{}
	prog := newProgram(t, root)
	for _, flag := range prog.Root.Flags {
		names := flag.Names()
		if len(names) > 0 && names[0] == "verbose" {
			verbose, ok := flag.(*clilib.BoolFlag)
			require.True(t, ok)
			assert.Equal(t, "Increase log verbosity (repeatable)", verbose.Usage)
			assert.Same(t, &prog.Flags.Verbosity, verbose.Config.Count)
			return
		}
	}
	t.Fatal("verbose flag not registered")
}

func TestRunCustomExitCode(t *testing.T) {
	root := &clilib.Command{Commands: []*clilib.Command{greetCommand(errors.New("boom"))}}
	prog := newProgram(t, root, curfave.WithExitCode(func(error) int { return 42 }))
	assert.Equal(t, 42, prog.Run(context.Background(), []string{"demo", "greet"}))
}

func TestRunChainsExistingBefore(t *testing.T) {
	ran := false
	root := &clilib.Command{
		Commands: []*clilib.Command{greetCommand(nil)},
		Before: func(ctx context.Context, _ *clilib.Command) (context.Context, error) {
			ran = true
			return ctx, nil
		},
	}
	prog := newProgram(t, root)
	require.Equal(t, 0, prog.Run(context.Background(), []string{"demo", "greet"}))
	assert.True(t, ran, "existing Before hook should still run")
}

func TestGeneratorIncludesSubcommands(t *testing.T) {
	root := &clilib.Command{Commands: []*clilib.Command{greetCommand(nil)}}
	prog := newProgram(t, root, curfave.WithVersionCommand())
	assert.NotEmpty(t, prog.Generator().Subs)
}

func TestWithoutStandardFlags(t *testing.T) {
	root := &clilib.Command{}
	newProgram(t, root, curfave.WithoutStandardFlags())
	for _, flag := range root.Flags {
		assert.NotContains(t, flag.Names(), "verbose")
		assert.NotContains(t, flag.Names(), "quiet")
		assert.NotContains(t, flag.Names(), "color")
	}
}

func TestWithSelfUpdate(t *testing.T) {
	root := &clilib.Command{Action: func(context.Context, *clilib.Command) error { return nil }}
	var got error
	prog := newProgram(t, root, curfave.WithSelfUpdate(), curfave.WithExitCode(func(err error) int {
		got = err
		return 42
	}))
	require.Equal(t, 42, prog.Run(context.Background(), []string{"demo", "--self-update"}))
	require.EqualError(t, got, "update command requires a self-updating updater, got <nil>")

	// The flag is mutually exclusive with every other argument.
	require.Equal(
		t,
		42,
		prog.Run(context.Background(), []string{"demo", "--self-update", "--verbose"}),
	)
	require.EqualError(t, got, "--self-update cannot be combined with other arguments")
}

func TestUpdateCommandRequiresSupportedUpdater(t *testing.T) {
	app := conductor.New(conductor.App{Name: "demo"})
	root := &clilib.Command{Commands: []*clilib.Command{curfave.UpdateCommand(app)}}
	err := root.Run(context.Background(), []string{"demo", "update"})
	require.EqualError(
		t,
		err,
		"update command requires a self-updating updater, got <nil>",
	)
}
