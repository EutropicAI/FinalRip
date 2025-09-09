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

type ProgressRequest struct {
	VideoKey string `form:"video_key" binding:"required"`
}

type ProgressResponse struct {
	CreateAt    int64                  `json:"create_at"`
	EncodeKey   string                 `json:"encode_key"`
	EncodeParam string                 `json:"encode_param"`
	EncodeSize  string                 `json:"encode_size"`
	EncodeURL   string                 `json:"encode_url"`
	Key         string                 `json:"key"`
	Progress    []db.VideoProgressITEM `json:"progress"`
	Script      string                 `json:"script"`
	Size        string                 `json:"size"`
	Status      string                 `json:"status"`
	URL         string                 `json:"url"`
}

// Progress 查看进度 (GET /progress)
func Progress(c *gin.Context) {
	// 绑定参数
	var req ProgressRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	progress, err := db.GetVideoProgress(req.VideoKey)
	if err != nil {
		log.Logger.Errorf("db.GetVideoProgress failed, err: %v", err)
		resp.AbortWithMsg(c, err.Error())
		return
	}

	var wg sync.WaitGroup
	for i := range progress {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// encode url
			if progress[i].EncodeKey == "" {
				progress[i].EncodeURL = ""
			} else {
				encodeUrl, err := oss.GetPresignedURL(progress[i].EncodeKey, progress[i].EncodeKey, 48*time.Hour)
				if err != nil {
					log.Logger.Errorf("oss.GetPresignedURL failed, err: %v", err)
					resp.AbortWithMsg(c, err.Error())
					return
				}
				progress[i].EncodeURL = encodeUrl
			}
			// clip url
			{
				clipUrl, err := oss.GetPresignedURL(progress[i].ClipKey, progress[i].ClipKey, 48*time.Hour)
				if err != nil {
					log.Logger.Errorf("oss.GetPresignedURL failed, err: %v", err)
					resp.AbortWithMsg(c, err.Error())
					return
				}
				progress[i].ClipURL = clipUrl
			}
		}()
	}

	t, err := db.GetTask(req.VideoKey)
	if err != nil {
		log.Logger.Errorf("db.GetCompletedEncodeKey failed, err: %v", err)
		resp.AbortWithMsg(c, err.Error())
		return
	}

	var url string
	var encodeUrl string
	wg.Add(2)

	go func() {
		defer wg.Done()
		url, err = oss.GetPresignedURL(t.Key, t.Key, 48*time.Hour)
		if err != nil {
			log.Logger.Errorf("oss.GetPresignedURL failed, err: %v", err)
			resp.AbortWithMsg(c, err.Error())
			return
		}
	}()
	go func() {
		defer wg.Done()
		if t.EncodeKey == "" {
			log.Logger.Warnf("encode task not completed, key: %s", req.VideoKey)
			encodeUrl = ""
		} else {
			encodeUrl, err = oss.GetPresignedURL(t.EncodeKey, t.EncodeKey, 48*time.Hour)
			if err != nil {
				log.Logger.Errorf("oss.GetPresignedURL failed, err: %v", err)
				resp.AbortWithMsg(c, err.Error())
				return
			}
		}
	}()

	var size string
	var encodeSize string
	wg.Add(2)

	go func() {
		defer wg.Done()
		size, err = oss.Size(t.Key)
		if err != nil {
			log.Logger.Errorf("get origin size failed, err: %v", err)
			resp.AbortWithMsg(c, err.Error())
			return
		}
	}()
	go func() {
		defer wg.Done()
		if t.EncodeKey != "" {
			encodeSize, err = oss.Size(t.EncodeKey)
			if err != nil {
				log.Logger.Errorf("get encode size failed, err: %v", err)
				resp.AbortWithMsg(c, err.Error())
				return
			}
		}
	}()

	status := task.TASK_STATUS_COMPLETED
	if t.EncodeParam == "" {
		status = task.TASK_STATUS_PENDING
	} else if t.EncodeKey == "" {
		status = task.TASK_STATUS_RUNNING
	}

	wg.Wait()

	resp.OKWithData(c, &ProgressResponse{
		Key:         t.Key,
		URL:         url,
		EncodeKey:   t.EncodeKey,
		EncodeParam: t.EncodeParam,
		EncodeSize:  encodeSize,
		EncodeURL:   encodeUrl,
		Progress:    progress,
		Script:      t.Script,
		Size:        size,
		Status:      status,
		CreateAt:    t.CreatedAt.Unix(),
	})
}
