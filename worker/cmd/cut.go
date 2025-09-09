package cmd

import (
	"github.com/EutropicAI/FinalRip/module/config"
	"github.com/EutropicAI/FinalRip/module/db"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/oss"
	"github.com/EutropicAI/FinalRip/module/queue"
	"github.com/EutropicAI/FinalRip/worker/internal/cut"
	"github.com/urfave/cli/v2"
)

var CutWorker = &cli.Command{
	Name:        "cut",
	Usage:       "Start FinalRip Cut Worker",
	Description: "Start FinalRip Cut Worker",
	Action:      runCutWorker,
}

func runCutWorker(ctx *cli.Context) error {
	config.Init()
	log.Init()
	db.Init()
	oss.Init()
	queue.InitCutWorker()
	cut.Start()
	return nil
}
