// Command demo shows the conductor urfave/cli integration: one App struct
// plus one Program call replaces the usual hand-rolled logger, help,
// completion, version and update-check bootstrap.
package main

import (
	"context"
	"os"

	"github.com/gechr/clive"
	"github.com/gechr/clive/updater/brew"
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	curfave "github.com/gechr/conductor/cli/urfave"
	clilib "github.com/urfave/cli/v3"
)

func greetCommand() *clilib.Command {
	var shout bool
	return &clilib.Command{
		Name:      "greet",
		Usage:     "Print a greeting",
		ArgsUsage: "[name]",
		Flags: []clilib.Flag{
			&clilib.BoolFlag{Name: "shout", Usage: "Greet loudly", Destination: &shout},
		},
		Action: func(_ context.Context, cmd *clilib.Command) error {
			name := cmd.Args().First()
			if name == "" {
				name = "world"
			}
			greeting := "hello"
			if shout {
				greeting = "HELLO"
			}
			clog.Info().Str("name", name).Msg(greeting)
			return nil
		},
	}
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
	})

	root := &clilib.Command{Commands: []*clilib.Command{greetCommand()}}
	prog := curfave.New(app, root, curfave.WithVersionCommand(), curfave.WithUpdateCommand())
	os.Exit(prog.Run(context.Background(), os.Args))
}
