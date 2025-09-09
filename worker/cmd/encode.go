package cmd

import (
	"github.com/EutropicAI/FinalRip/module/config"
	"github.com/EutropicAI/FinalRip/module/db"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/oss"
	"github.com/EutropicAI/FinalRip/module/queue"
	"github.com/EutropicAI/FinalRip/worker/internal/encode"
	"github.com/urfave/cli/v2"
)

var EncodeWorker = &cli.Command{
	Name:        "encode",
	Usage:       "Start FinalRip Enocde Worker",
	Description: "Start FinalRip Enocde Worker",
	Action:      runEncodeWorker,
}

func runEncodeWorker(ctx *cli.Context) error {
	config.Init()
	log.Init()
	db.Init()
	oss.Init()
	queue.InitEncodeWorker()
	encode.Start()
	return nil
}
