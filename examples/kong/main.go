// Command demo shows the conductor kong integration: one App struct plus one
// Program call replaces the usual hand-rolled logger, help, completion,
// version and update-check bootstrap.
package main

import (
	"errors"
	"os"

	"github.com/gechr/clive"
	"github.com/gechr/clive/updater/brew"
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	ckong "github.com/gechr/conductor/cli/kong"
	xstrings "github.com/gechr/x/strings"
)

// CLI is the kong grammar. Embedding [ckong.Flags] adds completion management
// plus -v/-q/--color/-V, applied to clog automatically after parsing.
type CLI struct {
	ckong.Flags

	Greet   cmdGreet         `help:"Print a greeting"                  cmd:""`
	Update  ckong.UpdateCmd  `help:"Update demo to the latest version" cmd:""`
	Version ckong.VersionCmd `help:"Print version information"         cmd:""`
}

type cmdGreet struct {
	Name  string `help:"Who to greet" arg:"" default:"world"`
	Shout bool   `help:"Greet loudly"`
}

// Run greets, demonstrating clog structured fields and error → exit-code
// mapping (try `demo greet ""`).
func (c *cmdGreet) Run() error {
	if xstrings.IsBlank(c.Name) {
		return errors.New("name must not be blank")
	}
	greeting := "hello"
	if c.Shout {
		greeting = "HELLO"
	}
	clog.Info().Str("name", c.Name).Msg(greeting)
	return nil
}

func main() {
	app := conductor.New(conductor.App{
		Name:        "demo",
		Description: "Conductor demo application.",
		Module:      "github.com/gechr/conductor",
		Updater: brew.New(
			clive.Info{Module: "github.com/gechr/conductor"},
			brew.WithFormula("demo"),
			brew.WithTap("gechr/tap"),
		),
		// Escape hatch: direct clog customisation on top of the defaults.
		ConfigureLog: func() {
			clog.SetSmartQuotes(true)
		},
	})

	var cli CLI
	prog, err := ckong.New(app, &cli)
	if err != nil {
		clog.Fatal().Err(err).Msg("Failed to build CLI")
	}
	os.Exit(prog.Run(os.Args[1:]))
}
