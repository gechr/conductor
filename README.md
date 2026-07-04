# Conductor

One `App` struct wires up:

- [clib](https://github.com/gechr/clib) - themed help, shell completions
- [clog](https://github.com/gechr/clog) - structured logging
- [clive](https://github.com/gechr/clive) - versioning, self-update

Direct use of the underlying libraries is never blocked.

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

See the runnable [kong](examples/kong/main.go), [cobra](examples/cobra/main.go) and [urfave](examples/urfave/main.go) demo apps.

## Lifecycle

1. **`conductor.New(App)`** - env prefix on clog *and* clib, [`LogDefaults`](log.go) (stderr, auto color, human-friendly durations), your `ConfigureLog` hook, theme detection, help renderer.
2. **Adapter `New(app, ...)`** - parser/command tree with themed help, standard flags, `-V`/`version`, completions, update hook.
3. **`Program.Run(...)`** - completion preflight, parse, apply `--verbose`/`--quiet`/`--color`, update check, dispatch, hint flush, error → exit code. Pass the result to `os.Exit`.

## Standard flags

Every adapter provides the same trio plus completion management:

| Flag              | Effect                              |
| ----------------- | ----------------------------------- |
| `-v`, `--verbose` | `clog.SetVerbose(true)`             |
| `-q`, `--quiet`   | `clog.SetLevel(clog.LevelError)`    |
| `--color <mode>`  | `clog.SetColorMode` (auto default)  |
| `-V`, `--version` | Print the version                   |

- [kong](cli/kong): embed [`ckong.Flags`](cli/kong/flags.go) in your grammar
- [cobra](cli/cobra)/[urfave](cli/urfave): registered as persistent/root flags automatically

## Version stamping

No per-app `version` var - stamp clive's linker variables (unstamped builds fall back to Go module build info, then the VCS revision):

```text
go build -ldflags "-X github.com/gechr/clive.version=$(VERSION) -X github.com/gechr/clive.buildTime=$(BUILD_TIME)"
```

## Update notifications

- Set `App.Updater` to any `updater.Tool` (`brew.New`, `goinstall.New`, `github.New`) to enable the passive "update available" hint, cached and rate-limited by clive; `<APP>_NO_UPDATE_CHECK` disables it.
- The standard `update` subcommand ([`UpdateCmd`](cli/kong/commands.go) for kong; [`UpdateCommand`](cli/cobra/commands.go) for cobra/urfave) self-updates via whichever install method `App.Updater` uses, with `--check`, `--stable`, `--dev` and `--no-uninstall` (the latter two are Homebrew-specific).

## Escape hatches

A thin layer over the same public APIs you would call yourself:

- `App.ConfigureLog` runs after `LogDefaults` - the place for direct clog customisation (symbols, styles, parts, custom levels).
- `App.SkipLogDefaults` opts out of the house style; the standalone `conductor.LogDefaults()` opts into it without the rest.
- `Runtime.Theme`/`Runtime.Renderer` and every `Program` field are exported and replaceable before `Run`.
- Adapter options (`WithKongOptions`, `WithSections`, `WithGenerator`, `WithExitCode`, ...) append after conductor's defaults, so they can override anything.
- `conductor.SignalContext()` - opt-in SIGINT/SIGTERM context for commands that should abort cleanly on Ctrl-C.
- Tools with bespoke needs skip the relevant piece (don't embed the flags, provide your own `update` command) and keep the rest.
