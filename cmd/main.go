package main

import (
	"fmt"
	"github.com/arf-rpc/arfc/arf"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:        "arfc",
		Usage:       "arf compiler",
		Version:     "v0.1.0",
		Description: "Compiles arf idl files into source files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "lang",
				Usage:    "The destination language",
				Required: true,
				Aliases:  []string{"l"},
			},
			&cli.StringFlag{
				Name:      "input",
				Usage:     "The input file to generate sources from",
				Required:  true,
				TakesFile: true,
				Aliases:   []string{"i"},
			},
			&cli.StringFlag{
				Name:     "output",
				Usage:    "The output directory to write to",
				Required: true,
				Aliases:  []string{"o"},
			},
			&cli.BoolFlag{
				Name: "ruby-flat",
				Usage: "Creates all files in the level of the output directory, without creating " +
					"subdirectories matching the output module path",
				Category: "Ruby",
			},
			&cli.StringSliceFlag{
				Name: "ruby-module",
				Usage: "When lang is set to \"ruby\", overrides the generated module name for a given package. Must " +
					"be in the format some.package.name=Module::Name",
				Category: "Ruby",
			},
			&cli.StringSliceFlag{
				Name: "golang-package",
				Usage: "When lang is set to \"go\", overrides the generated package name for a given arf package. Must " +
					"be in the format some.package.name=package",
				Category: "Go",
			},
		},
		Action: arf.Run,
		Authors: []*cli.Author{
			{Name: "Vito Sartori", Email: "hey@vito.io"},
		},
		Copyright: "Copyright (c) The arf Authors",
		Suggest:   true,
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
