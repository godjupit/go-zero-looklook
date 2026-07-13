package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gin-looklook/internal/app"
	"gin-looklook/internal/config"

	"github.com/gin-gonic/gin"
)

func TestRouterContainsAllCompatibleRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(&app.App{Config: config.Config{JWTSecret: "test"}})
	if got := len(router.Routes()); got != 22 { // original 17 + 3 seckill + health + metrics
		t.Fatalf("route count=%d, want 22", got)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/order/v1/homestayOrder/userHomestayOrderList", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), `"code":100003`) {
		t.Fatalf("unexpected auth response: status=%d body=%s", w.Code, w.Body.String())
	}
}
