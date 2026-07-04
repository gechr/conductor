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
