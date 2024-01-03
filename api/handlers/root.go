package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type RootHandler struct{}

func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

func (h *RootHandler) Register(g *echo.Group) {
	g.GET("/", h.Root)
}

func (h *RootHandler) Root(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func HandlerNotImplemented(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Not Implemented")
}
