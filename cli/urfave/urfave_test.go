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
}

func TestRunDispatchesAndAppliesFlags(t *testing.T) {
	t.Cleanup(func() { clog.SetVerbose(false) })

	root := &clilib.Command{Commands: []*clilib.Command{greetCommand(nil)}}
	prog := newProgram(t, root)
	code := prog.Run(context.Background(), []string{"demo", "--verbose", "greet"})
	assert.Equal(t, 0, code)
	assert.True(t, clog.IsVerbose(), "verbose flag should reach clog")
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
