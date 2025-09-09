package task

import (
	"time"

	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/oss"
	"github.com/EutropicAI/FinalRip/module/resp"
	"github.com/gin-gonic/gin"
)

type OSSPresignedRequest struct {
	VideoKey string `form:"video_key" binding:"required"`
}

type OSSPresignedResponse struct {
	Exist bool   `json:"exist"`
	URL   string `json:"url"`
}

// OSSPresigned 获取 OSS 上传 URL (GET /oss/presigned)
func OSSPresigned(c *gin.Context) {
	// 绑定参数
	var req OSSPresignedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.AbortWithMsg(c, err.Error())
		return
	}

	url, err := oss.GetUploadPresignedURL(req.VideoKey, 48*time.Hour)
	if err != nil {
		log.Logger.Error("get upload presigned url failed: " + err.Error())
		resp.AbortWithMsg(c, "get upload presigned url failed: "+err.Error())
		return
	}

	resp.OKWithData(c, &OSSPresignedResponse{
		Exist: oss.Exist(req.VideoKey),
		URL:   url,
	})
}
