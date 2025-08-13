package prometheus

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type GinxPrometheus struct {
	vector *prometheus.SummaryVec
}

func NewBuilder(opt prometheus.SummaryOpts) *GinxPrometheus {
	vector := prometheus.NewSummaryVec(opt, []string{"method", "path", "status"})
	return &GinxPrometheus{vector: vector}
}

func (g *GinxPrometheus) Build() gin.HandlerFunc {
	prometheus.MustRegister(g.vector)
	return func(ctx *gin.Context) {
		now := time.Now()
		ctx.Next()
		g.vector.WithLabelValues(ctx.Request.Method, ctx.FullPath(), strconv.Itoa(ctx.Writer.Status())).
			Observe(float64(time.Since(now).Milliseconds()))
	}
}
