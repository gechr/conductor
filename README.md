# conductor

🎼 One-shot initialization for Go CLIs.

conductor glues together [clib](https://github.com/gechr/clib) (themed help, shell completions), [clog](https://github.com/gechr/clog) (structured logging) and [clive](https://github.com/gechr/clive) (versioning, self-update): fill in one `App` struct at program start and your CLI inherits the whole stack - house logging style, unified env-var prefixes, themed help, completion management, version plumbing and passive update notifications - while remaining free to call any of the underlying libraries directly.

## Packages

| Package      | Description                                                    |
| ------------ | -------------------------------------------------------------- |
| `conductor`  | `App`/`Runtime` core: logging defaults, flags, notify, version |
| `cli/cobra`  | [Cobra](https://github.com/spf13/cobra) integration            |
| `cli/kong`   | [Kong](https://github.com/alecthomas/kong) integration         |
| `cli/urfave` | [urfave/cli](https://github.com/urfave/cli) integration        |

## Installation

```text
go get github.com/gechr/conductor
```

## Usage

```go
app := conductor.New(conductor.App{
    Name:        "demo",
    Description: "Demo application.",
    Module:      "github.com/you/demo",
    Updater: brew.New(
        clive.Info{Module: "github.com/you/demo"},
        brew.WithFormula("demo"),
        brew.WithTap("you/tap"),
    ),
})

var cli CLI // your kong grammar, embedding ckong.Flags
prog, err := ckong.New(app, &cli)
if err != nil {
    clog.Fatal().Err(err).Msg("Failed to build CLI")
}
os.Exit(prog.Run(os.Args[1:]))
```

That one `New` + `Program.Run` pair replaces the bootstrap every tool otherwise hand-rolls: clog output/format setup, `clog.SetEnvPrefix` + `theme.SetEnvPrefix`, `theme.Auto()` → `help.NewRenderer` → help-printer wiring, the `Reflect` → completion-generator → preflight block, post-parse `--color`/`--verbose`/`--quiet` application, version flag/subcommand, and the passive update check with its deferred hint flush and skip-on-`update` logic.

See [`examples/`](examples/) for runnable kong, cobra and urfave demo apps.

## Lifecycle

1. **`conductor.New(App)`** - pre-parse phase. Sets the env prefix on clog *and* clib, applies [`LogDefaults`](log.go) (stderr output with auto color, space-separated slices, human-friendly duration gradients), runs your `ConfigureLog` hook, resolves the theme (before any output, so background detection is clean) and builds the help renderer.
2. **Adapter `New(app, ...)`** - builds/wires the parser or command tree: themed help, standard flags, `-V`/`version`, completion flags and the update notification hook.
3. **`Program.Run(...)`** - completion preflight, parse, post-parse flag application (`--verbose`/`--quiet`/`--color` → clog), passive update check (skipped for `update`, `version`, `help` and completion verbs), dispatch, deferred update-hint flush, and error → exit-code mapping. Pass the result to `os.Exit`.

## Standard flags

Every adapter provides the same trio plus completion management:

| Flag              | Effect                              |
| ----------------- | ----------------------------------- |
| `-v`, `--verbose` | `clog.SetVerbose(true)`             |
| `-q`, `--quiet`   | `clog.SetLevel(clog.LevelError)`    |
| `--color <mode>`  | `clog.SetColorMode` (auto default)  |
| `-V`, `--version` | Print the version                   |

Kong CLIs embed `ckong.Flags`; cobra and urfave register them as persistent/root flags automatically.

## Version stamping

conductor standardises on clive's linker variables - no per-app `version` var needed:

```text
go build -ldflags "-X github.com/gechr/clive.version=$(VERSION) -X github.com/gechr/clive.buildTime=$(BUILD_TIME)"
```

Without stamping, clive falls back to Go module build info, then the VCS revision.

## Update notifications

Set `App.Updater` to any `updater.Tool` (`brew.New`, `goinstall.New`, `github.New`) to enable the passive "update available" hint, cached and rate-limited by clive. The standard `update` subcommand (`UpdateCmd` for kong; `UpdateCommand` for cobra/urfave) self-updates via whichever install method `App.Updater` uses - Homebrew, `go install`, or GitHub release download - with `--check`, `--stable`, `--dev` and `--no-uninstall` (the latter two are Homebrew-specific: other methods treat stable as the default channel and manage no conflicting copies). The `<APP>_NO_UPDATE_CHECK` env var disables the hint entirely.

## Escape hatches

conductor is a thin layer over the same public APIs you would call yourself, and never blocks direct use of them:

- `App.ConfigureLog` runs after `LogDefaults` - the place for direct clog customisation (symbols, styles, parts, custom levels).
- `App.SkipLogDefaults` opts out of the house style; the standalone `conductor.LogDefaults()` opts into it without the rest.
- `Runtime.Theme`/`Runtime.Renderer` and every `Program` field are exported and replaceable before `Run`.
- Adapter options (`WithKongOptions`, `WithSections`, `WithGenerator`, `WithExitCode`, ...) append after conductor's defaults, so they can override anything.
- Tools with bespoke needs skip the relevant piece (don't embed the flags, provide your own `update` command) and keep the rest.
