package cmd

import (
	"github.com/EutropicAI/FinalRip/module/config"
	"github.com/EutropicAI/FinalRip/module/db"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/oss"
	"github.com/EutropicAI/FinalRip/module/queue"
	"github.com/EutropicAI/FinalRip/worker/internal/merge"
	"github.com/urfave/cli/v2"
)

var MergeWorker = &cli.Command{
	Name:        "merge",
	Usage:       "Start FinalRip Merge Worker",
	Description: "Start FinalRip Merge Worker",
	Action:      runMergeWorker,
}

func runMergeWorker(ctx *cli.Context) error {
	config.Init()
	log.Init()
	db.Init()
	oss.Init()
	queue.InitMergeWorker()
	merge.Start()
	return nil
}
