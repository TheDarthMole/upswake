package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerHandler_Register(t *testing.T) {
	e := echo.New()
	h := NewServerHandler()

	g := e.Group("")
	h.Register(g)

	expectedRoutes := []string{"/wake", "/broadcastwake"}
	for _, route := range e.Routes() {
		for i, expected := range expectedRoutes {
			if expected == route.Path {
				expectedRoutes = append(expectedRoutes[:i], expectedRoutes[i+1:]...)
				break
			}
		}
	}

	assert.Equal(t, 2, len(e.Routes()), "Expected 2 routes to be registered")
	assert.Equalf(t, []string{}, expectedRoutes, "The following expected routes are missing: %v", expectedRoutes)
}
