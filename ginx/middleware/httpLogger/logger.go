package httpLogger

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"sync/atomic"
	"time"
)

type HTTPLoggerBuilder struct {
	AllowReqBody  atomic.Bool
	AllowRespBody atomic.Bool

	LogFunc func(msg string, al *AccessLog)
}

func NewBuilder(logFunc func(msg string, al *AccessLog)) *HTTPLoggerBuilder {
	return &HTTPLoggerBuilder{
		LogFunc: logFunc,
	}
}

type AccessLog struct {
	StatusCode int
	URL        string
	Method     string
	ReqBody    string
	RespBody   string
	Cost       string
}

func (h *HTTPLoggerBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if h.LogFunc == nil {
			ctx.Next()
			return
		}

		start := time.Now()
		al := &AccessLog{}
		al.Method = ctx.Request.Method
		al.URL = func() string {
			url := ctx.Request.URL.String()
			if len(url) > 1024 {
				return url[0:1024] + "...(truncated)"
			}
			return url
		}()
		if h.AllowReqBody.Load() {
			body := ctx.Request.Body
			b, err := io.ReadAll(body)
			if err != nil {
				ctx.Next()
				return
			}
			if len(b) > 1024 {
				al.ReqBody = string(b[:1024])
			} else {
				al.ReqBody = string(b)
			}
			ctx.Request.Body = io.NopCloser(bytes.NewReader(b))
		}

		if h.AllowRespBody.Load() {
			ctx.Writer = &logWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
		}
		ctx.Next()
		al.Cost = time.Since(start).String()
		al.StatusCode = ctx.Writer.Status()
		h.LogFunc("server http request:", al)
	}
}

type logWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (l *logWriter) WriteString(body string) (int, error) {
	if len(body) > 1024 {
		l.al.RespBody = body[0:1024] + "...(truncated)"
	} else {
		l.al.RespBody = body
	}
	return l.ResponseWriter.WriteString(body)
}
func (l *logWriter) Write(p []byte) (int, error) {
	if len(p) > 1024 {
		l.al.RespBody = string(p[:1024]) + "...(truncated)"
	} else {
		l.al.RespBody = string(p)
	}
	return l.ResponseWriter.Write(p)
}
