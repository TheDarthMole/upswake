package handlers

import "github.com/labstack/echo/v4"

type RootHandler struct{}

func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

func (h *RootHandler) Register(g *echo.Group) {
	g.GET("/", h.Root)
}

func (h *RootHandler) Root(c echo.Context) error {
	return c.String(200, "Hello, World!")
}
