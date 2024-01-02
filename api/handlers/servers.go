package handlers

import "github.com/labstack/echo/v4"

type ServerHandler struct{}

func NewServerHandler() *ServerHandler {
	return &ServerHandler{}
}

func (h *ServerHandler) Register(g *echo.Group) {
	g.GET("", h.ListServers)
	g.GET("/:mac", func(c echo.Context) error { return c.String(501, "Not Implemented") })
	g.POST("/:mac/wake", func(c echo.Context) error { return c.String(501, "Not Implemented") })
}

func (h *ServerHandler) ListServers(c echo.Context) error {

	return c.String(501, "Not Implemented")
}
