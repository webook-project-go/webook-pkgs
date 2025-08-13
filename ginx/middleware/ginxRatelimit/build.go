package ginxRatelimit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/webook-project-go/webook-pkgs/ratelimit"
	"net/http"
)

type GinxRatelimit struct {
	limiter ratelimit.RateLimiter
	Prefix  string
}

func NewBuilder(limiter ratelimit.RateLimiter) *GinxRatelimit {
	return &GinxRatelimit{
		limiter: limiter,
		Prefix:  "ip-limiter",
	}
}

func (g *GinxRatelimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limit, err := g.limit(ctx)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limit {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func (g *GinxRatelimit) limit(ctx *gin.Context) (bool, error) {
	return g.limiter.Limit(ctx, fmt.Sprintf("%s:%s", g.Prefix, ctx.ClientIP()))
}
