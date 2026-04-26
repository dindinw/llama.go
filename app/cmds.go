package app

import (
	"context"
	"fmt"

	"github.com/Qitmeer/llama.go/api"
	"github.com/Qitmeer/llama.go/app/embedding"
	econfig "github.com/Qitmeer/llama.go/app/embedding/config"
	"github.com/Qitmeer/llama.go/app/pull"
	"github.com/Qitmeer/llama.go/app/run"
	"github.com/Qitmeer/llama.go/common"
	"github.com/Qitmeer/llama.go/config"
	"github.com/Qitmeer/llama.go/server"
	"github.com/Qitmeer/llama.go/system"
	"github.com/Qitmeer/llama.go/system/limits"
	"github.com/Qitmeer/llama.go/version"
	"github.com/Qitmeer/llama.go/wrapper"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

func commands() []*cli.Command {
	cmds := []*cli.Command{}
	cmds = append(cmds, versionCmd())
	cmds = append(cmds, serveCmd())
	cmds = append(cmds, runCmd())
	cmds = append(cmds, pullCmd())
	cmds = append(cmds, embeddingCmd())
	cmds = append(cmds, whisperCmd())
	return cmds
}

func OnBefore(ctx *cli.Context) error {
	err := initLog(config.Conf)
	if err != nil {
		return err
	}
	return checkServerHeartbeat(ctx.Context)
}

func OnBeforeForServe(ctx *cli.Context) error {
	err := initLog(config.Conf)
	if err != nil {
		return err
	}
	log.Info("Before init")
	err = config.Conf.Load()
	if err != nil {
		return err
	}
	return nil
}

func versionCmd() *cli.Command {
	return &cli.Command{
		Name:        "version",
		Aliases:     []string{"v"},
		Category:    "llama",
		Usage:       "Show llama.go version",
		Description: "Show llama.go version",
		Action: func(ctx *cli.Context) error {
			print(version.String())
			return nil
		},
	}
}

func runCmd() *cli.Command {
	return &cli.Command{
		Name:        "run",
		Aliases:     []string{"r"},
		Category:    "llama",
		Usage:       "llama.go run [prompt]",
		Description: "llama.go run - Run llama with optional prompt as positional argument",
		ArgsUsage:   "[prompt]",
		Flags:       run.AppFlags,
		Before:      OnBefore,
		Action:      run.RunHandler,
	}
}

func serveCmd() *cli.Command {
	return &cli.Command{
		Name:        "serve",
		Aliases:     []string{"s"},
		Category:    "llama",
		Usage:       "llama.go server",
		Description: "llama.go server",
		Before:      OnBeforeForServe,
		Action: func(ctx *cli.Context) error {
			err := limits.SetLimits()
			if err != nil {
				return err
			}
			interrupt := system.InterruptListener()
			cfg := config.Conf

			ser := server.New(ctx, cfg)

			err = ser.Start()
			defer func() {
				err = ser.Stop()
				if err != nil {
					log.Error(err.Error())
				}
			}()

			if err != nil {
				return err
			}
			<-interrupt

			return nil
		},
	}
}

func pullCmd() *cli.Command {
	return &cli.Command{
		Name:        "pull",
		Aliases:     []string{"pu"},
		Category:    "llama",
		Usage:       "llama.go pull MODEL",
		Description: "Download a model from a registry",
		ArgsUsage:   "MODEL",
		Before:      OnBefore,
		Action:      pull.PullHandler,
	}
}

func embeddingCmd() *cli.Command {
	return &cli.Command{
		Name:        "embedding",
		Aliases:     []string{"e"},
		Category:    "llama",
		Usage:       "llama.go embedding [PROMPT]",
		Description: "Generate high-dimensional embedding vector of a given text",
		Flags:       econfig.AppFlags,
		Before:      OnBeforeForServe,
		Action:      embedding.EmbeddingHandler,
	}
}

func whisperCmd() *cli.Command {
	return &cli.Command{
		Name:        "whisper",
		Aliases:     []string{"w"},
		Category:    "whisper",
		Usage:       "Generate text by whisper model",
		Description: "Generate text by whisper model",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "Input file path for generate.",
			},
		},
		Action: func(ctx *cli.Context) error {
			cfg := config.Conf
			log.Info("Start whisper")
			if !ctx.IsSet("input") {
				return fmt.Errorf("No input file")
			}
			ret, err := wrapper.WhisperGenerate(cfg, ctx.Value("input").(string))
			if err != nil {
				return err
			}
			if len(cfg.OutputFile) > 0 {
				return common.SaveOutputToFile(cfg.OutputFile, ret)
			} else {
				fmt.Println("result:")
				fmt.Println(ret)
			}
			return nil
		},
	}
}

func checkServerHeartbeat(ctx context.Context) error {
	client := api.DefaultClient()
	err := client.Heartbeat(ctx)
	if err != nil {
		return err
	}
	return nil
}
