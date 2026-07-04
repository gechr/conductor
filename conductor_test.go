package conductor

import (
	"errors"
	"testing"

	"github.com/gechr/clog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequiresName(t *testing.T) {
	assert.Panics(t, func() { New(App{}) })
}

func TestNewBuildsRuntime(t *testing.T) {
	rt := New(App{
		Name:   "demo",
		Module: "github.com/gechr/demo",
	})
	require.NotNil(t, rt.Theme)
	require.NotNil(t, rt.Renderer)
	assert.Equal(t, "github.com/gechr/demo", rt.Info.Module)
}

func TestEnvPrefix(t *testing.T) {
	tests := []struct {
		name string
		app  App
		want string
	}{
		{"derived from name", App{Name: "demo"}, "DEMO"},
		{"dashes mapped", App{Name: "my-tool"}, "MY_TOOL"},
		{"explicit", App{Name: "demo", EnvPrefix: "CUSTOM"}, "CUSTOM"},
		{"disabled", App{Name: "demo", EnvPrefix: EnvPrefixNone}, EnvPrefixNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.app.envPrefix())
		})
	}
}

func TestDisplayNameDefaultsToName(t *testing.T) {
	assert.Equal(t, "demo", New(App{Name: "demo"}).App.DisplayName)
	assert.Equal(t, "Demo", New(App{Name: "demo", DisplayName: "Demo"}).App.DisplayName)
}

func TestSignalContext(t *testing.T) {
	ctx, stop := SignalContext()
	require.NotNil(t, ctx)
	stop()
	<-ctx.Done()
}

func TestConfigureLogRunsAfterDefaults(t *testing.T) {
	ran := false
	New(App{Name: "demo", ConfigureLog: func() { ran = true }})
	assert.True(t, ran)
}

func TestApplyFlags(t *testing.T) {
	rt := New(App{Name: "demo"})

	rt.ApplyFlags(Flags{Verbose: true})
	assert.True(t, clog.IsVerbose())

	rt.ApplyFlags(Flags{})
	assert.False(t, clog.IsVerbose())

	rt.ApplyFlags(Flags{Quiet: true})
	assert.Equal(t, clog.LevelError, clog.GetLevel())

	// Restore a sane default level for other tests.
	rt.ApplyFlags(Flags{})
}

func TestNotifySkipped(t *testing.T) {
	tests := []struct {
		name    string
		app     App
		command string
		want    bool
	}{
		{"plain command checks", App{Name: "demo"}, "greet", false},
		{"subcommand uses leading verb", App{Name: "demo"}, "greet loudly", false},
		{"update verb skips", App{Name: "demo"}, "update", true},
		{"update subcommand skips", App{Name: "demo"}, "update --check", true},
		{"version skips", App{Name: "demo"}, "version", true},
		{"help skips", App{Name: "demo"}, "help", true},
		{"cobra completion skips", App{Name: "demo"}, "__complete foo", true},
		{"custom skip verb", App{Name: "demo", NotifySkip: []string{"login"}}, "login", true},
		{"allowlist match checks", App{Name: "demo", NotifyOnly: []string{"run"}}, "run", false},
		{"allowlist miss skips", App{Name: "demo", NotifyOnly: []string{"run"}}, "greet", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := New(tt.app)
			assert.Equal(t, tt.want, rt.notifySkipped(tt.command))
		})
	}
}

func TestNotifyNilUpdaterIsNoop(t *testing.T) {
	rt := New(App{Name: "demo"})
	flush := rt.Notify("greet")
	require.NotNil(t, flush)
	flush() // must be safe to call
}

func TestExitCode(t *testing.T) {
	assert.Equal(t, 0, ExitCode(nil, nil))

	custom := func(error) int { return 42 }
	assert.Equal(t, 42, ExitCode(errors.New("boom"), custom))

	clog.SetLevel(clog.LevelFatal) // suppress the default error log
	t.Cleanup(func() { clog.SetLevel(clog.LevelInfo) })
	assert.Equal(t, 1, ExitCode(errors.New("boom"), nil))
}

func TestHelpDescDefaults(t *testing.T) {
	app := App{Name: "demo"}
	assert.Equal(t, "Print short help", app.HelpShortDesc())
	assert.Equal(t, "Print long help, including examples", app.HelpLongDesc())

	app = App{Name: "demo", HelpShort: "s", HelpLong: "l"}
	assert.Equal(t, "s", app.HelpShortDesc())
	assert.Equal(t, "l", app.HelpLongDesc())
}
