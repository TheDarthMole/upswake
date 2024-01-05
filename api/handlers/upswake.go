package handlers

import (
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/labstack/echo/v4"
	"net/http"
)

type UPSWakeHandler struct {
	cfg *config.Config
}

func NewUPSWakeHandler(cfg *config.Config) *UPSWakeHandler {
	return &UPSWakeHandler{
		cfg: cfg,
	}
}

func (h *UPSWakeHandler) Register(g *echo.Group) {
	g.GET("", h.ListNutServerMappings)
}

func (h *UPSWakeHandler) ListNutServerMappings(c echo.Context) error {
	return c.JSON(http.StatusOK, h.cfg.NutServerMappings)
}
