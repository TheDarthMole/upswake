package handlers

import (
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/labstack/echo/v4"
	"net/http"
)

type RootHandler struct{}

type Response struct {
	Message string `json:"message"`
}

func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

func (h *RootHandler) Register(g *echo.Group) {
	g.GET("/", h.Root)
	g.GET("/health", h.Health)
}

func (h *RootHandler) Root(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{Message: "Hello, World!"})
}

func (h *RootHandler) Health(c echo.Context) error {
	if _, err := util.GetAllBroadcastAddresses(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, Response{Message: "OK"})
}

func HandlerNotImplemented(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, Response{Message: "Not Implemented"})
}
