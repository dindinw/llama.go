// Copyright (c) 2017-2025 The qitmeer developers

package app

import (
	"github.com/Qitmeer/llama.go/config"
	"github.com/Qitmeer/llama.go/version"
	"github.com/urfave/cli/v2"
	"os"
)

func Run() error {
	app := &cli.App{
		Name:    "llama.go",
		Version: version.String(),
		Authors: []*cli.Author{
			&cli.Author{
				Name: "Qitmeer",
			},
		},
		Copyright:            "(c) 2026 Qitmeer",
		Usage:                "Llama",
		Flags:                config.AppFlags,
		EnableBashCompletion: true,
		Commands:             commands(),
		Action: func(c *cli.Context) error {
			print(version.String())
			return nil
		},
	}

	return app.Run(os.Args)
}
