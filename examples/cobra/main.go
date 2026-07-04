// Command demo shows the conductor cobra integration: one App struct plus one
// Program call replaces the usual hand-rolled logger, help, completion,
// version and update-check bootstrap.
package main

import (
	"context"
	"os"

	"github.com/gechr/clive"
	"github.com/gechr/clive/updater/brew"
	"github.com/gechr/clog"
	"github.com/gechr/conductor"
	ccobra "github.com/gechr/conductor/cli/cobra"
	"github.com/spf13/cobra"
)

func greetCommand() *cobra.Command {
	var shout bool
	cmd := &cobra.Command{
		Use:   "greet [name]",
		Short: "Print a greeting",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := "world"
			if len(args) > 0 {
				name = args[0]
			}
			greeting := "hello"
			if shout {
				greeting = "HELLO"
			}
			clog.Info().Str("name", name).Msg(greeting)
			return nil
		},
	}
	cmd.Flags().BoolVar(&shout, "shout", false, "Greet loudly")
	return cmd
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

	root := &cobra.Command{}
	root.AddCommand(greetCommand())
	prog := ccobra.New(app, root, ccobra.WithVersionCommand(), ccobra.WithUpdateCommand())
	os.Exit(prog.Run(context.Background(), os.Args[1:]))
}
