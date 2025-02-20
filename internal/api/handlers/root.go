package handlers

import (
	"github.com/TheDarthMole/UPSWake/internal/util"
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

// Root godoc
//
//	@Summary		Root redirect to swagger
//	@Description	Redirect to swagger docs
//	@Tags			root
//	@Accept			plain
//	@Produce		html
//	@Router			/ [get]
func (h *RootHandler) Root(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
}

// Health godoc
//
//	@Summary		Health check
//	@Description	Health check
//	@Tags			root
//	@Accept			json
//	@Produce		json
//
//	@Success		200	{object}	Response	"OK"
//	@Failure		500	{object}	Response
//	@Router			/health [get]
func (h *RootHandler) Health(c echo.Context) error {
	if _, err := util.GetAllBroadcastAddresses(); err != nil {
		c.Logger().Errorf("Error getting broadcast addresses: %s", err)
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	c.Logger().Debugf("Health check OK")
	return c.JSON(http.StatusOK, Response{Message: "OK"})
}
