package handlers

import (
	"net/http"
	"net/http/pprof"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/labstack/echo/v5"
)

type ProfilerHandler struct {
	profiler *entity.Profiler
}

func NewProfilerHandler(profiler *entity.Profiler) *ProfilerHandler {
	return &ProfilerHandler{
		profiler: profiler,
	}
}

// Register middleware for net/http/pprof
// Inspired by https://github.com/labstack/echo-contrib/blob/master/pprof/pprof_test.go
func (*ProfilerHandler) Register(group *echo.Group) {
	group.GET("/", handler(pprof.Index))
	group.GET("/allocs", handler(pprof.Handler("allocs").ServeHTTP))
	group.GET("/block", handler(pprof.Handler("block").ServeHTTP))
	group.GET("/cmdline", handler(pprof.Cmdline))
	group.GET("/goroutine", handler(pprof.Handler("goroutine").ServeHTTP))
	group.GET("/heap", handler(pprof.Handler("heap").ServeHTTP))
	group.GET("/mutex", handler(pprof.Handler("mutex").ServeHTTP))
	group.GET("/profile", handler(pprof.Profile))
	group.POST("/symbol", handler(pprof.Symbol))
	group.GET("/symbol", handler(pprof.Symbol))
	group.GET("/threadcreate", handler(pprof.Handler("threadcreate").ServeHTTP))
	group.GET("/trace", handler(pprof.Trace))
}

func handler(h http.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
