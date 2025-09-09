package queue

import (
	"strconv"
	"sync"

	"github.com/EutropicAI/FinalRip/common/version"
	"github.com/EutropicAI/FinalRip/module/config"
	"github.com/hibiken/asynq"
)

var (
	once sync.Once
	Qc   *asynq.Client
	Qs   *asynq.Server
	Isp  *asynq.Inspector
)

const (
	CUT_QUEUE             = "cut_" + version.FINALRIP_VERSION
	ENCODE_QUEUE_DEFAULT  = "encode_default_" + version.FINALRIP_VERSION
	ENCODE_QUEUE_PRIORITY = "encode_priority_" + version.FINALRIP_VERSION
	MERGE_QUEUE           = "merge_" + version.FINALRIP_VERSION
)

func getRedisClientOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     config.RedisConfig.Host + ":" + strconv.Itoa(config.RedisConfig.Port),
		Password: config.RedisConfig.Password,
		PoolSize: config.RedisConfig.PoolSize,
		DB:       0,
	}
}

func InitServer() {
	once.Do(func() {
		Qc = asynq.NewClient(getRedisClientOpt())

		Isp = asynq.NewInspector(getRedisClientOpt())
	})
}

func InitCutWorker() {
	once.Do(func() {
		Qs = asynq.NewServer(
			getRedisClientOpt(),
			asynq.Config{
				Concurrency: 1,
				Queues: map[string]int{
					CUT_QUEUE: 1,
				},
			},
		)
	})
}

func InitEncodeWorker() {
	once.Do(func() {
		Qs = asynq.NewServer(
			getRedisClientOpt(),
			asynq.Config{
				Concurrency: 1,
				Queues: map[string]int{
					ENCODE_QUEUE_DEFAULT:  1,
					ENCODE_QUEUE_PRIORITY: 9,
				},
				StrictPriority: true, // enable strict priority, higher priority queues are always processed first
			},
		)
	})
}

func InitMergeWorker() {
	once.Do(func() {
		Qs = asynq.NewServer(
			getRedisClientOpt(),
			asynq.Config{
				Concurrency: 1,
				Queues: map[string]int{
					MERGE_QUEUE: 1,
				},
			},
		)
	})
}
