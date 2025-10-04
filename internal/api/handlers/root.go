package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/network"
	"github.com/TheDarthMole/UPSWake/internal/ups"
	"github.com/labstack/echo/v4"
	"github.com/spf13/afero"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type RootHandler struct {
	cfg     *entity.Config
	rulesFS afero.Fs
}

type Response struct {
	Message string `json:"message"`
}

func NewRootHandler(cfg *entity.Config, rulesFS afero.Fs) *RootHandler {
	return &RootHandler{
		cfg:     cfg,
		rulesFS: rulesFS,
	}
}

func sanitizeString(input string) string {
	// Replace any non-printable characters with an empty string
	sanatised := strconv.QuoteToASCII(input)
	sanatised = strings.TrimSpace(sanatised)
	sanatised = strings.ReplaceAll(sanatised, "\n", "")
	sanatised = strings.ReplaceAll(sanatised, "\r", "")
	return sanatised
}

func (h *RootHandler) Register(g *echo.Group) {
	g.GET("/", h.Root)
	g.GET("/health", h.Health)
	g.GET("/swagger/*", echoSwagger.WrapHandler)
}

// Root godoc
//
//	@Summary		Root redirect to swagger
//	@Description	Redirect to swagger docs
//	@Tags			root
//	@Accept			plain
//	@Produce		html
//	@Router			/ [get]
func (*RootHandler) Root(c echo.Context) error {
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
	if err := h.cfg.Validate(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	if _, err := network.GetAllBroadcastAddresses(); err != nil {
		c.Logger().Errorf("Error getting broadcast addresses: %s", err)
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	// TODO: Speed this up by running in parallel
	for _, server := range h.cfg.NutServers {
		if _, err := ups.GetJSON(&server); err != nil {
			c.Logger().Errorf("Error getting NUT server status: %s", err)
			return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
	}

	c.Logger().Debugf("Health check OK")
	return c.JSON(http.StatusOK, Response{Message: "OK"})
}
