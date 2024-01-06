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
	mappings := h.cfg.NutServerMappings
	// Don't leak passwords
	for i, mapping := range mappings {
		mapping.NutServer.Credentials.Password = "********"
		mappings[i] = mapping
	}
	return c.JSON(http.StatusOK, mappings)
}
