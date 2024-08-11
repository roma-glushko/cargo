package main

import (
	"github.com/roma-glushko/cargo/cmd"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r", "serve", "srv"},
				Usage:   "run Cargo API server",
				Action:  cmd.Run,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
