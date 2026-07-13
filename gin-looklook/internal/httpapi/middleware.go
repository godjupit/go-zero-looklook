package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"gin-looklook/internal/platform"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
)

type successResponse struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
type errorResponse struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successResponse{Code: 200, Msg: "OK", Data: data})
}
func Fail(c *gin.Context, err error) {
	code, msg := platform.Public(err)
	slog.ErrorContext(c.Request.Context(), "request failed", "path", c.Request.URL.Path, "error", err)
	c.JSON(http.StatusBadRequest, errorResponse{Code: code, Msg: msg})
}
func Bind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		Fail(c, platform.E(platform.CodeParam, "参数错误, "+err.Error(), err))
		return false
	}
	return true
}

func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if len(raw) > 7 && raw[:7] == "Bearer " {
			raw = raw[7:]
		}
		token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Code: platform.CodeToken, Msg: "token失效，请重新登陆"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Code: platform.CodeToken, Msg: "token失效，请重新登陆"})
			return
		}
		id, ok := claims["jwtUserId"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Code: platform.CodeToken, Msg: "token失效，请重新登陆"})
			return
		}
		c.Set("userID", int64(id))
		c.Next()
	}
}
func UserID(c *gin.Context) int64 { v, _ := c.Get("userID"); id, _ := v.(int64); return id }

var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "gin_looklook_http_requests_total", Help: "HTTP requests"}, []string{"method", "path", "status"})
var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "gin_looklook_http_request_duration_seconds", Help: "HTTP latency", Buckets: prometheus.DefBuckets}, []string{"method", "path"})

func init() { prometheus.MustRegister(requestCount, requestDuration) }
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		requestCount.WithLabelValues(c.Request.Method, path, strconv.Itoa(c.Writer.Status())).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}
