package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestPProfRegisterDefaultPrefix(t *testing.T) {
	pprofPaths := []struct {
		path string
	}{
		{"/"},
		{"/allocs"},
		{"/block"},
		{"/cmdline"},
		{"/goroutine"},
		{"/heap"},
		{"/mutex"},
		{"/profile?seconds=1"},
		{"/symbol"},
		{"/symbol"},
		{"/threadcreate"},
		{"/trace"},
	}
	for _, tt := range pprofPaths {
		t.Run(tt.path, func(t *testing.T) {
			e := echo.New()

			profilerHandler := NewProfilerHandler()
			profilerHandler.Register(e.Group(""))
			req, _ := http.NewRequest(http.MethodGet, tt.path, http.NoBody)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestPProfRegisterCustomPrefix(t *testing.T) {
	pprofPaths := []struct {
		path string
	}{
		{"/"},
		{"/allocs"},
		{"/block"},
		{"/cmdline"},
		{"/goroutine"},
		{"/heap"},
		{"/mutex"},
		{"/profile?seconds=1"},
		{"/symbol"},
		{"/symbol"},
		{"/threadcreate"},
		{"/trace"},
	}
	for _, tt := range pprofPaths {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			e := echo.New()

			prefix := "/test/profiler"

			profilerHandler := NewProfilerHandler()
			profilerHandler.Register(e.Group(prefix))
			req, _ := http.NewRequest(http.MethodGet, prefix+tt.path, http.NoBody)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
