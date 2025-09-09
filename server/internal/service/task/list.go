package task

import (
	"sync"
	"time"

	"github.com/EutropicAI/FinalRip/common/db"
	"github.com/EutropicAI/FinalRip/common/task"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/oss"
	"github.com/EutropicAI/FinalRip/module/resp"
	"github.com/gin-gonic/gin"
)

type ListRequest struct {
	Completed *bool `form:"completed" binding:"required"`
	Pending   *bool `form:"pending" binding:"required"`
	Running   *bool `form:"running" binding:"required"`
}

type ListResponse []ListItem

type ListItem struct {
	CreateAt    int64  `json:"create_at"`
	EncodeKey   string `json:"encode_key"`
	EncodeParam string `json:"encode_param"`
	EncodeURL   string `json:"encode_url"`
	Key         string `json:"key"`
	Script      string `json:"script"`
	Status      string `json:"status"`
	URL         string `json:"url"`
}

// List 查看任务列表 (GET /list)
func List(c *gin.Context) {
	// 绑定参数
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	tasks, err := db.ListTask()
	if err != nil {
		log.Logger.Errorf("Failed to list tasks: %v", err)
		resp.AbortWithMsg(c, "Failed to list tasks")
		return
	}

	list := make(ListResponse, 0)
	for _, t := range tasks {
		status := task.TASK_STATUS_COMPLETED
		if t.EncodeParam == "" {
			status = task.TASK_STATUS_PENDING
		} else if t.EncodeKey == "" {
			status = task.TASK_STATUS_RUNNING
		}

		if (*req.Completed && status == task.TASK_STATUS_COMPLETED) ||
			(*req.Pending && status == task.TASK_STATUS_PENDING) ||
			(*req.Running && status == task.TASK_STATUS_RUNNING) {
			list = append(list, ListItem{
				CreateAt:    t.CreatedAt.Unix(),
				EncodeKey:   t.EncodeKey,
				EncodeParam: t.EncodeParam,
				EncodeURL:   "",
				Key:         t.Key,
				Script:      t.Script,
				Status:      status,
				URL:         "",
			})
		}
	}

	var wg sync.WaitGroup
	for i := range list {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// encode url
			if list[i].EncodeKey == "" {
				list[i].EncodeURL = ""
			} else {
				encodeUrl, err := oss.GetPresignedURL(list[i].EncodeKey, list[i].EncodeKey, 48*time.Hour)
				if err != nil {
					log.Logger.Errorf("oss.GetPresignedURL failed, err: %v", err)
					resp.AbortWithMsg(c, err.Error())
					return
				}
				list[i].EncodeURL = encodeUrl
			}
			// url
			{
				url, err := oss.GetPresignedURL(list[i].Key, list[i].Key, 48*time.Hour)
				if err != nil {
					log.Logger.Errorf("oss.GetPresignedURL failed, err: %v", err)
					resp.AbortWithMsg(c, err.Error())
					return
				}
				list[i].URL = url
			}
		}()
	}

	wg.Wait()

	resp.OKWithData(c, &list)
}
