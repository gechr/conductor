package kong_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	ckong "github.com/gechr/conductor/cli/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCLI struct {
	ckong.Flags

	Greet greetCmd `help:"Greet" cmd:""`
}

type greetCmd struct{}

func (greetCmd) Run() error { return nil }

func newProgram(t *testing.T, cli any, opts ...ckong.Option) *ckong.Program {
	t.Helper()
	app := conductor.New(conductor.App{Name: "demo", Description: "Test app."})
	opts = append(opts, ckong.WithKongOptions(konglib.Writers(io.Discard, io.Discard)))
	prog, err := ckong.New(app, cli, opts...)
	require.NoError(t, err)
	return prog
}

func TestNewBuildsProgram(t *testing.T) {
	var cli testCLI
	prog := newProgram(t, &cli)
	require.NotNil(t, prog.Parser)
	require.NotNil(t, prog.Gen)
	assert.NotEmpty(t, prog.Gen.Subs, "subcommand specs should be populated")
}

func TestRunDispatchesAndAppliesFlags(t *testing.T) {
	t.Cleanup(func() { clog.SetVerbose(false) })

	var cli testCLI
	prog := newProgram(t, &cli)
	code := prog.Run([]string{"greet", "--verbose"})
	assert.Equal(t, 0, code)
	assert.True(t, clog.IsVerbose(), "verbose flag should reach clog")
}

func TestRunParseErrorReturnsWithoutExiting(t *testing.T) {
	var cli testCLI
	prog := newProgram(t, &cli)
	var stderr bytes.Buffer
	prog.Parser.Stderr = &stderr
	code := prog.Run([]string{"--no-such-flag"})
	assert.Equal(t, 80, code, "kong usage errors exit with code 80")
	assert.NotEmpty(t, stderr.String())
}

func TestRunCustomExitCode(t *testing.T) {
	var cli struct {
		ckong.Flags

		Boom boomCmd `help:"Fail" cmd:""`
	}
	prog := newProgram(t, &cli, ckong.WithExitCode(func(error) int { return 42 }))
	assert.Equal(t, 42, prog.Run([]string{"boom"}))
}

type boomCmd struct{}

func (boomCmd) Run() error { return errors.New("boom") }

func TestSelfUpdateFlag(t *testing.T) {
	t.Cleanup(func() { clog.SetLevel(clog.LevelInfo) })
	clog.SetLevel(clog.LevelFatal) // suppress the expected error log

	var cli struct {
		ckong.Flags
		ckong.SelfUpdateFlag

		Greet greetCmd `help:"Greet" cmd:"" default:"1"`
	}
	var got error
	prog := newProgram(t, &cli, ckong.WithExitCode(func(err error) int {
		got = err
		return 42
	}))
	// Nil App.Updater: the self-update path runs and fails with the
	// requires-updater error, proving the flag short-circuits dispatch.
	require.Equal(t, 42, prog.Run([]string{"--self-update"}))
	require.EqualError(t, got, "update command requires a self-updating updater, got <nil>")

	// The flag is mutually exclusive with every other argument.
	require.Equal(t, 42, prog.Run([]string{"--self-update", "--verbose"}))
	require.EqualError(t, got, "--self-update cannot be combined with other arguments")
}

func TestSelfUpdateFlagHiddenFromHelp(t *testing.T) {
	var cli struct {
		ckong.Flags
		ckong.SelfUpdateFlag

		Greet greetCmd `help:"Greet" cmd:""`
	}
	prog := newProgram(t, &cli)
	node := prog.Parser.Model.Node
	for _, f := range node.Flags {
		if f.Name == "self-update" {
			require.True(t, f.Hidden, "--self-update must be hidden")
		}
	}
}

func TestUpdateCmdRequiresSupportedUpdater(t *testing.T) {
	app := conductor.New(conductor.App{Name: "demo"})
	cmd := &ckong.UpdateCmd{}
	err := cmd.Run(app)
	require.EqualError(
		t,
		err,
		"update command requires a self-updating updater, got <nil>",
	)
}
