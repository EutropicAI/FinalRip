package logger

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/util"
	"github.com/gin-gonic/gin"
)

// LogLayout 日志 layout
type LogLayout struct {
	RequestMethod string // 请求方法字段
	StatusCode    int    // 状态码
	Time          time.Time
	Metadata      map[string]interface{} // 存储自定义原数据
	Path          string                 // 访问路径
	Query         string                 // 携带 query
	Body          string                 // 携带 body 数据
	IP            string                 // ip 地址
	UserAgent     string                 // 代理
	Error         string                 // 错误
	Cost          time.Duration          // 花费时间
	Source        string                 // 来源
}

type Logger struct {
	// Filter 用户自定义过滤
	Filter func(c *gin.Context) bool
	// FilterKeyword 关键字过滤 (key)
	FilterKeyword func(layout *LogLayout) bool
	// AuthProcess 鉴权处理
	AuthProcess func(c *gin.Context, layout *LogLayout)
	// 日志处理
	Print func(LogLayout)
	// Source 服务唯一标识
	Source string
}

func (l Logger) SetLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		var body []byte
		if l.Filter != nil && !l.Filter(c) {
			body, _ = c.GetRawData()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		// 放行
		c.Next()

		cost := time.Since(start)
		layout := LogLayout{
			RequestMethod: c.Request.Method,
			StatusCode:    c.Writer.Status(),
			Time:          time.Now(),
			Path:          path,
			Query:         query,
			IP:            c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			Error:         strings.TrimRight(c.Errors.ByType(gin.ErrorTypePrivate).String(), "\n"),
			Cost:          cost,
			Source:        l.Source,
		}

		if l.Filter != nil && !l.Filter(c) {
			layout.Body = string(body)
		}
		if l.AuthProcess != nil {
			// 处理鉴权需要的信息
			l.AuthProcess(c, &layout)
		}
		if l.FilterKeyword != nil {
			// 自行判断 key/value 脱敏等
			l.FilterKeyword(&layout)
		}
		// 自行处理日志
		l.Print(layout)
	}
}

func DefaultLogger() gin.HandlerFunc {
	return Logger{
		Print: func(layout LogLayout) {
			StatusMessage := layout.RequestMethod + ": " + strconv.Itoa(layout.StatusCode)
			if layout.Error == "" {
				log.Logger.Info(util.HighlightString(util.GREEN, StatusMessage) + " - " + util.StructToString(layout))
			} else {
				log.Logger.Error(util.HighlightString(util.RED, StatusMessage) + " - " + util.StructToString(layout))
			}
		},
		Source: "FinalRip",
	}.SetLoggerMiddleware()
}
