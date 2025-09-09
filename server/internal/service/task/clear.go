package task

import (
	"sync"

	"github.com/EutropicAI/FinalRip/common/db"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/oss"
	"github.com/EutropicAI/FinalRip/module/queue"
	"github.com/EutropicAI/FinalRip/module/resp"
	"github.com/gin-gonic/gin"
)

type ClearRequest struct {
	VideoKey string `form:"video_key" binding:"required"`
}

// Clear 清理任务 (POST /new)
func Clear(c *gin.Context) {
	// 绑定参数
	var req NewRequest
	if err := c.ShouldBind(&req); err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	// 检查任务是否存在
	if !db.CheckTaskExist(req.VideoKey) {
		resp.AbortWithMsg(c, "Task not found, please check the video key")
		return
	}

	clips, err := db.GetVideoClips(req.VideoKey)
	if err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	// 清理 OSS
	ossDelObjKeys := make([]string, 0)
	task, _ := db.GetTask(req.VideoKey)
	ossDelObjKeys = append(ossDelObjKeys, task.Key)
	ossDelObjKeys = append(ossDelObjKeys, task.EncodeKey)

	for _, clip := range clips {
		if clip.ClipKey != "" {
			ossDelObjKeys = append(ossDelObjKeys, clip.ClipKey)
		}
		if clip.EncodeKey != "" {
			ossDelObjKeys = append(ossDelObjKeys, clip.EncodeKey)
		}
	}

	err = ossDelMulti(ossDelObjKeys)
	if err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}
	log.Logger.Infof("Deleted files from OSS: %v", ossDelObjKeys)

	// 清理数据库
	err = db.DeleteTask(req.VideoKey)
	if err != nil {
		log.Logger.Errorf("Failed to delete task from database: %s", err)
		resp.AbortWithMsg(c, err.Error())
		return
	}

	err = db.DeleteVideoClips(req.VideoKey)
	if err != nil {
		log.Logger.Errorf("Failed to delete video clips from database: %s", err)
		resp.AbortWithMsg(c, err.Error())
		return
	}
	log.Logger.Infof("Deleted task from database: %s", req.VideoKey)

	// 一定要在删除数据库记录和 OSS 之后再取消任务，否则会导致 Merge 任务无法正常取消，以及 Encode 任务异常下载 OSS
	// 检查任务是否处理完成
	if !db.CheckTaskComplete(req.VideoKey) {
		// 清理任务队列，倒序删除
		for i := len(clips) - 1; i >= 0; i-- {
			clip := clips[i]
			CancelTask(clip.TaskID)
		}
	}

	resp.OK(c)
}

// ossDelMulti 批量删除 OSS 对象，使用协程并发删除
func ossDelMulti(keys []string) error {
	var ossErr error
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			err := oss.Delete(k)
			if err != nil {
				log.Logger.Errorf("Failed to delete file from OSS: %s", err)
				ossErr = err
			}
			wg.Done()
		}(key)
	}
	wg.Wait()
	return ossErr
}

// CancelTask 取消任务队列中的任务
func CancelTask(taskID string) {
	err := queue.Isp.CancelProcessing(taskID)
	if err != nil {
		log.Logger.Warnf("Failed to cancel processing task: %s", err)
	}
	err = queue.Isp.DeleteTask(queue.ENCODE_QUEUE_DEFAULT, taskID)
	if err != nil {
		log.Logger.Warnf("Failed to delete task in default encode queue: %s", err)
	}
	err = queue.Isp.DeleteTask(queue.ENCODE_QUEUE_PRIORITY, taskID)
	if err != nil {
		log.Logger.Warnf("Failed to delete task in priority encode queue: %s", err)
	}
}
