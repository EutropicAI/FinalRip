package cmd

import (
	"github.com/EutropicAI/FinalRip/module/config"
	"github.com/EutropicAI/FinalRip/module/db"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/oss"
	"github.com/EutropicAI/FinalRip/module/queue"
	"github.com/EutropicAI/FinalRip/server/internal/router"
	"github.com/urfave/cli/v2"
)

// CmdWeb api 子命令
var CmdWeb = &cli.Command{
	Name:        "server",
	Usage:       "Start FinalRip Api Server",
	Description: "Start FinalRip Api Server",
	Action:      runWeb,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   "3000",
			Usage:   "Temporary port number to prevent conflict",
		},
	},
}

func runWeb(ctx *cli.Context) error {
	config.Init()
	log.Init()
	db.Init()
	oss.Init()
	queue.InitServer()
	router.Init()
	return nil
}
