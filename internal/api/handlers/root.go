package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	directups "github.com/TheDarthMole/UPSWake/internal/infrastructure/ups/direct_ups"
	"github.com/TheDarthMole/UPSWake/internal/network"
	"github.com/labstack/echo/v5"
	"github.com/spf13/afero"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
	"golang.org/x/sync/errgroup"
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
	sanitised := strconv.QuoteToASCII(input)
	sanitised = strings.TrimSpace(sanitised)
	sanitised = strings.ReplaceAll(sanitised, "\n", "")
	sanitised = strings.ReplaceAll(sanitised, "\r", "")
	return sanitised
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
func (*RootHandler) Root(c *echo.Context) error {
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
func (h *RootHandler) Health(c *echo.Context) error {
	if err := h.cfg.Validate(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	if _, err := network.GetAllBroadcastAddresses(); err != nil {
		c.Logger().Error("Error getting broadcast addresses", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	g := errgroup.Group{}

	upsRepo := directups.NewDirectRepository()
	for _, server := range h.cfg.NutServers {
		g.Go(func() error {
			if _, err := upsRepo.GetJSON(server); err != nil {
				c.Logger().Error("Error getting NUT server status", slog.Any("error", err))
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		c.Logger().Debug("Health check failed", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	c.Logger().Debug("Health check OK")
	return c.JSON(http.StatusOK, Response{Message: "OK"})
}
