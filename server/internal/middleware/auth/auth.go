package auth

import (
	"github.com/EutropicAI/FinalRip/module/config"
	"github.com/EutropicAI/FinalRip/module/resp"
	"github.com/gin-gonic/gin"
)

// RequireAuth 鉴权中间件
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 token
		token := c.Request.Header.Get("token")

		if token != config.ServerConfig.Token {
			resp.AbortWithMsg(c, "Token is invalid, please check it")
			return
		}

		// 放行
		c.Next()
	}
}
