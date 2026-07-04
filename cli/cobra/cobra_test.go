package cobra_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	ccobra "github.com/gechr/conductor/cli/cobra"
	cobralib "github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProgram(t *testing.T, root *cobralib.Command, opts ...ccobra.Option) *ccobra.Program {
	t.Helper()
	app := conductor.New(conductor.App{Name: "demo", Description: "Test app."})
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	return ccobra.New(app, root, opts...)
}

func greetCommand(err error) *cobralib.Command {
	return &cobralib.Command{
		Use:   "greet",
		Short: "Greet",
		RunE:  func(*cobralib.Command, []string) error { return err },
	}
}

func TestNewDefaultsIdentity(t *testing.T) {
	root := &cobralib.Command{}
	prog := newProgram(t, root)
	assert.Equal(t, "demo", prog.Root.Use)
	assert.Equal(t, "Test app.", prog.Root.Short)
	assert.True(t, prog.Root.SilenceUsage)
	assert.True(t, prog.Root.SilenceErrors)
}

func TestRunDispatchesAndAppliesFlags(t *testing.T) {
	t.Cleanup(func() { clog.SetVerbose(false) })

	root := &cobralib.Command{}
	root.AddCommand(greetCommand(nil))
	prog := newProgram(t, root)
	code := prog.Run(context.Background(), []string{"greet", "--verbose"})
	assert.Equal(t, 0, code)
	assert.True(t, clog.IsVerbose(), "verbose flag should reach clog")
}

func TestRunCustomExitCode(t *testing.T) {
	root := &cobralib.Command{}
	root.AddCommand(greetCommand(errors.New("boom")))
	prog := newProgram(t, root, ccobra.WithExitCode(func(error) int { return 42 }))
	assert.Equal(t, 42, prog.Run(context.Background(), []string{"greet"}))
}

func TestRunChainsExistingPreRun(t *testing.T) {
	ran := false
	root := &cobralib.Command{
		PersistentPreRunE: func(*cobralib.Command, []string) error {
			ran = true
			return nil
		},
	}
	root.AddCommand(greetCommand(nil))
	prog := newProgram(t, root)
	require.Equal(t, 0, prog.Run(context.Background(), []string{"greet"}))
	assert.True(t, ran, "existing PersistentPreRunE should still run")
}

func TestGeneratorIncludesSubcommands(t *testing.T) {
	root := &cobralib.Command{}
	root.AddCommand(greetCommand(nil))
	prog := newProgram(t, root, ccobra.WithVersionCommand())
	gen := prog.Generator()
	assert.NotEmpty(t, gen.Subs)
}

func TestUpdateCommandRequiresSupportedUpdater(t *testing.T) {
	app := conductor.New(conductor.App{Name: "demo"})
	cmd := ccobra.UpdateCommand(app)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	err := cmd.ExecuteContext(context.Background())
	require.EqualError(
		t,
		err,
		"update command supports brew, goinstall and github updaters, got <nil>",
	)
}
