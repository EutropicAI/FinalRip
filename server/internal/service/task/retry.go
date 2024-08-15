package task

import (
	"errors"

	"github.com/TensoRaws/FinalRip/common/db"
	"github.com/TensoRaws/FinalRip/common/task"
	"github.com/TensoRaws/FinalRip/module/log"
	"github.com/TensoRaws/FinalRip/module/queue"
	"github.com/TensoRaws/FinalRip/module/resp"
	"github.com/TensoRaws/FinalRip/module/util"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type RetryEncodeRequest struct {
	VideoKey    string `form:"video_key" binding:"required"`
	EncodeParam string `form:"encode_param" binding:"required"`
	Index       *int   `form:"index" binding:"required"`
	Script      string `form:"script" binding:"required"`
}

// RetryEncode 重试压制任务 (POST /retry/encode)
func RetryEncode(c *gin.Context) {
	// 绑定参数
	var req RetryEncodeRequest
	if err := c.ShouldBind(&req); err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	info := db.VideoClipInfo{
		Key:     req.VideoKey,
		ClipKey: util.GenerateClipKey(req.VideoKey, *req.Index),
	}

	// 检查任务是否存在
	if !db.CheckVideoExist(info) {
		log.Logger.Error("Encode clip not exist: " + util.StructToString(info))
		resp.AbortWithMsg(c, "Encode clip not exist")
		return
	}

	clip, err := db.GetVideoClip(info)
	if err != nil {
		log.Logger.Error("Get video clip info failed: " + err.Error())
		resp.AbortWithMsg(c, "Get video clip info failed: "+err.Error())
		return
	}

	// 取消先前的任务
	CancelTask(clip.TaskID)
	// 重新创建任务
	err = HandleRetryEncode(req, clip)
	if err != nil {
		resp.AbortWithMsg(c, "Failed to retry encode: "+err.Error())
		return
	}

	resp.OK(c)
}

func HandleRetryEncode(req RetryEncodeRequest, clip db.VideoClipInfo) error {
	payload, err := sonic.Marshal(task.EncodeTaskPayload{
		EncodeParam: req.EncodeParam,
		Script:      req.Script,
		Clip:        clip,
		Retry:       true,
	})
	if err != nil {
		log.Logger.Error("Failed to marshal payload: " + err.Error())
		return err
	}

	encode := asynq.NewTask(task.VIDEO_ENCODE, payload)

	info, err := queue.Qc.Enqueue(encode, asynq.Queue(queue.ENCODE_QUEUE))
	if err != nil {
		log.Logger.Error("Failed to Re-enqueue task: " + err.Error())
		return err
	}

	// 清除 encode key
	err = db.UnsetVideoEncodeKey(db.VideoClipInfo{Key: req.VideoKey, ClipKey: clip.ClipKey})
	if err != nil {
		log.Logger.Error("Failed to unset encode key: " + err.Error())
		return err
	}

	// 更新 task id
	err = db.UpdateVideo(db.VideoClipInfo{Key: req.VideoKey, ClipKey: clip.ClipKey},
		db.VideoClipInfo{TaskID: info.ID})
	if err != nil {
		log.Logger.Error("Failed to Re-enqueue task: " + err.Error())
		return err
	}

	log.Logger.Info("Successfully Re-enqueued task: " + util.StructToString(clip))
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

type RetryMergeRequest struct {
	VideoKey string `form:"video_key" binding:"required"`
}

// RetryMerge 重试压制任务 (POST /retry/merge)
func RetryMerge(c *gin.Context) {
	// 绑定参数
	var req RetryMergeRequest
	if err := c.ShouldBind(&req); err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	// 检查任务是否存在
	if !db.CheckTaskExist(req.VideoKey) {
		resp.AbortWithMsg(c, "Task not found, please upload video first.")
		return
	}

	// 检查所有 clip 是否已经完成
	progress, err := db.GetVideoProgress(req.VideoKey)
	if err != nil {
		log.Logger.Error("Get video progress failed: " + err.Error())
		resp.AbortWithMsg(c, "Get video progress failed: "+err.Error())
		return
	}
	for _, status := range progress {
		if !status {
			resp.AbortWithMsg(c, "Some clips are not finished yet.")
			return
		}
	}

	// 重新创建任务
	err = HandleRetryMerge(req)
	if err != nil {
		resp.AbortWithMsg(c, "Failed to retry merge: "+err.Error())
		return
	}

	resp.OK(c)
}

func HandleRetryMerge(req RetryMergeRequest) error {
	// 开始合并任务
	clips, err := db.GetVideoClips(req.VideoKey)
	if err != nil {
		log.Logger.Error("Failed to get video clips: " + err.Error())
		return err
	}

	// 如果已经clear，不再合并
	if len(clips) == 0 {
		log.Logger.Info("No clips to merge.")
		return errors.New("no clips to merge")
	}

	payload, err := sonic.Marshal(task.MergeTaskPayload{
		Clips: clips,
	})
	if err != nil {
		log.Logger.Error("Failed to marshal payload: " + err.Error())
		return err
	}

	merge := asynq.NewTask(task.VIDEO_MERGE, payload)

	_, err = queue.Qc.Enqueue(merge, asynq.Queue(queue.MERGE_QUEUE))
	if err != nil {
		log.Logger.Error("Failed to Re-enqueue task: " + err.Error())
		return err
	}

	log.Logger.Info("Successfully Re-enqueued task: merge")
	return nil
}
