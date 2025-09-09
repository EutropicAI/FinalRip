package task

import (
	"github.com/EutropicAI/FinalRip/common/db"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/oss"
	"github.com/EutropicAI/FinalRip/module/resp"
	"github.com/gin-gonic/gin"
)

type NewRequest struct {
	VideoKey string `form:"video_key" binding:"required"`
}

// New 创建任务 (POST /new)
func New(c *gin.Context) {
	// 绑定参数
	var req NewRequest
	if err := c.ShouldBind(&req); err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	// 检查任务是否存在
	if db.CheckTaskExist(req.VideoKey) {
		resp.AbortWithMsg(c, "Task already exists, please wait for it to complete or delete it.")
		return
	}

	// 检查 OSS 文件是否存在
	if !oss.Exist(req.VideoKey) {
		log.Logger.Error("OSS video file does not exist: " + req.VideoKey)
		resp.AbortWithMsg(c, "OSS video file does not exist.")
		return
	}

	err := db.InsertTask(req.VideoKey)
	if err != nil {
		log.Logger.Error("Failed to insert task: " + err.Error())
		resp.AbortWithMsg(c, err.Error())
		return
	}

	resp.OK(c)
}
