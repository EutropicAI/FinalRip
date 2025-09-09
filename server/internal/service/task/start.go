package task

import (
	"errors"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/EutropicAI/FinalRip/common/constant"
	"github.com/EutropicAI/FinalRip/common/db"
	"github.com/EutropicAI/FinalRip/common/task"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/queue"
	"github.com/EutropicAI/FinalRip/module/resp"
	"github.com/EutropicAI/FinalRip/module/util"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type StartRequest struct {
	EncodeParam string  `form:"encode_param" binding:"required"`
	Script      string  `form:"script" binding:"required"`
	VideoKey    string  `form:"video_key" binding:"required"`
	Slice       *bool   `form:"slice"`
	Timeout     *int    `form:"timeout"`
	Queue       *string `form:"queue"`
}

// Start 开始压制 (POST /start)
func Start(c *gin.Context) {
	// 绑定参数
	var req StartRequest
	if err := c.ShouldBind(&req); err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	// 去除末尾空格，换行符等
	req.EncodeParam = strings.TrimRightFunc(req.EncodeParam, unicode.IsSpace)
	// 检查传入的 Script 和 EncodeParam 是否合法
	err := constant.CheckVSScriptAndEncodeParam(req.Script, req.EncodeParam)
	if err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	// 检查任务是否 new 上传
	if !db.CheckTaskExist(req.VideoKey) {
		resp.AbortWithMsg(c, "Task not found, please upload video first.")
		return
	}

	// 检查任务是否已经开始
	if db.CheckTaskStart(req.VideoKey) {
		resp.AbortWithMsg(c, "Task already started.")
		return
	}

	// 更新任务
	err = db.UpdateTask(db.Task{Key: req.VideoKey}, db.Task{
		EncodeParam: req.EncodeParam,
		Script:      req.Script,
	})
	if err != nil {
		log.Logger.Error("Failed to update task: " + err.Error())
		resp.AbortWithMsg(c, err.Error())
		return
	}

	resp.OK(c)
	go HandleStart(req)
}

// HandleStart 处理开始压制请求
func HandleStart(req StartRequest) {
	payload, err := sonic.Marshal(task.CutTaskPayload{
		VideoKey: req.VideoKey,
		Slice:    (req.Slice == nil) || *req.Slice, // 默认为 true
	})
	if err != nil {
		log.Logger.Error("Failed to marshal payload: " + err.Error())
		return
	}

	// 视频切片，上传到 OSS
	cut := asynq.NewTask(task.VIDEO_CUT, payload)

	info, err := queue.Qc.Enqueue(cut, asynq.Queue(queue.CUT_QUEUE))
	if err != nil {
		log.Logger.Error("Failed to enqueue task: " + err.Error())
		return
	}

	// 等待任务完成
	for {
		_, err := queue.Isp.GetTaskInfo(queue.CUT_QUEUE, info.ID)
		if err != nil {
			if errors.Is(err, asynq.ErrTaskNotFound) {
				break
			} else {
				log.Logger.Error("Unexpected error: " + err.Error())
				return
			}
		}

		time.Sleep(1 * time.Second)
	}

	log.Logger.Info("Cut task completed!")

	// 获取视频 clips
	clips, err := db.GetVideoClips(req.VideoKey)
	if err != nil {
		log.Logger.Error("Failed to get video clips: " + err.Error())
		return
	}

	encodeQueue := task.GetEncodeQueueName(req.Queue)

	var wg sync.WaitGroup
	// 开始压制任务
	for _, clip := range clips {
		payload, err := sonic.Marshal(task.EncodeTaskPayload{
			EncodeParam: req.EncodeParam,
			Script:      req.Script,
			Clip:        clip,
		})
		if err != nil {
			log.Logger.Error("Failed to marshal payload: " + err.Error())
			return
		}

		encode := asynq.NewTask(task.VIDEO_ENCODE, payload)

		info, err := queue.Qc.Enqueue(encode, asynq.Queue(encodeQueue), task.GetTaskTimeout(len(clips), req.Timeout))
		if err != nil {
			log.Logger.Error("Failed to enqueue task: " + err.Error())
			return
		}

		err = db.UpdateVideoClip(db.VideoClipInfo{Key: req.VideoKey, ClipKey: clip.ClipKey},
			db.VideoClipInfo{TaskID: info.ID})
		if err != nil {
			log.Logger.Error("Failed to enqueue task: " + err.Error())
			return
		}

		log.Logger.Info("Successfully enqueued task: " + util.StructToString(clip))

		wg.Add(1)
		go func(i *asynq.TaskInfo) {
			defer wg.Done()
			// 等待任务完成
			for {
				_, err := queue.Isp.GetTaskInfo(encodeQueue, i.ID)
				if err != nil {
					if errors.Is(err, asynq.ErrTaskNotFound) {
						break
					} else {
						log.Logger.Error("Unexpected error: " + err.Error())
						break
					}
				}

				time.Sleep(1 * time.Second)
			}
		}(info)
	}

	// 等待所有任务完成
	wg.Wait()
	log.Logger.Info("All Encode tasks completed!")

	// 开始合并任务
	clips, err = db.GetVideoClips(req.VideoKey)
	if err != nil {
		log.Logger.Error("Failed to get video clips: " + err.Error())
		return
	}

	// 如果已经clear，不再合并
	if len(clips) == 0 {
		log.Logger.Info("No clips to merge.")
		return
	}

	payload, err = sonic.Marshal(task.MergeTaskPayload{
		VideoKey: req.VideoKey,
	})
	if err != nil {
		log.Logger.Error("Failed to marshal payload: " + err.Error())
		return
	}

	merge := asynq.NewTask(task.VIDEO_MERGE, payload)

	info, err = queue.Qc.Enqueue(merge, asynq.Queue(queue.MERGE_QUEUE))
	if err != nil {
		log.Logger.Error("Failed to enqueue task: " + err.Error())
		return
	}

	log.Logger.Info("Successfully enqueued task: merge")

	// 等待任务完成
	for {
		_, err := queue.Isp.GetTaskInfo(queue.MERGE_QUEUE, info.ID)
		if err != nil {
			if errors.Is(err, asynq.ErrTaskNotFound) {
				break
			} else {
				log.Logger.Error("Unexpected error: " + err.Error())
				return
			}
		}

		time.Sleep(1 * time.Second)
	}

	log.Logger.Info("Merge task completed!")
}
