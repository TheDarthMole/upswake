package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/TheDarthMole/UPSWake/internal/evaluator"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/labstack/echo/v5"
)

type UPSWakeHandler struct {
	cfg      *entity.Config
	upsRepo  repository.UPSRepository
	ruleRepo repository.RuleRepository
}

type WakeEvaluationRequest struct {
	Mac string `json:"mac" example:"00:11:22:33:44:55"`
}

type UpsWakeResponse struct {
	Message string `json:"message" example:"Wake on LAN sent"`
	Woken   bool   `json:"woken" example:"true"`
}

// NewUPSWakeHandler creates a UPSWakeHandler configured with the supplied server configuration and repositories.
// The returned handler holds cfg, upsRepo and ruleRepo for use by its HTTP endpoints.
func NewUPSWakeHandler(cfg *entity.Config, upsRepo repository.UPSRepository, ruleRepo repository.RuleRepository) *UPSWakeHandler {
	return &UPSWakeHandler{
		cfg:      cfg,
		upsRepo:  upsRepo,
		ruleRepo: ruleRepo,
	}
}

func (h *UPSWakeHandler) Register(g *echo.Group) {
	g.GET("", h.ListNutServerMappings)
	g.POST("", h.RunWakeEvaluation)
}

// ListNutServerMappings godoc
//
//	@Summary		List NUT server mappings
//	@Description	List NUT server mappings using the config stored in the server
//	@Tags			UPSWake
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]viper.NutServer
//	@Router			/api/upswake [get]
func (h *UPSWakeHandler) ListNutServerMappings(c *echo.Context) error {
	nutServers := make([]*viper.NutServer, len(h.cfg.NutServers))

	for i, nutServer := range h.cfg.NutServers {
		nutServers[i] = viper.ToFileNutServer(nutServer)
		// Don't leak passwords
		nutServers[i].Password = "********"
	}

	return c.JSON(http.StatusOK, nutServers)
}

// RunWakeEvaluation godoc
//
//	@Summary		Run wake evaluation
//	@Description	Run wake evaluation using the config and rules stored in the server
//	@Tags			UPSWake
//	@Accept			json
//	@Produce		json
//	@Param			request			body		WakeEvaluationRequest	true	"the mac address of the target to wake"
//	@Success		200				{object}	UpsWakeResponse			"Wake on LAN sent"
//	@Success		304				{object}	UpsWakeResponse			"No rule evaluated to true"
//	@Failure		400				{object}	UpsWakeResponse			"Bad request"
//	@Failure		404				{object}	UpsWakeResponse			"MAC address not found in the config"
//	@Failure		500				{object}	UpsWakeResponse			"Internal server error"
//	@Router			/api/upswake	[post]
func (h *UPSWakeHandler) RunWakeEvaluation(c *echo.Context) error {
	request := &WakeEvaluationRequest{}
	if err := c.Bind(request); err != nil {
		c.Logger().Error("failed to bind mac address", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, UpsWakeResponse{
			Message: ErrorBindingRequest.Error(),
			Woken:   false,
		})
	}
	mac, err := entity.NewMacAddress(request.Mac)
	if err != nil {
		c.Logger().Error("failed to validate mac address", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, UpsWakeResponse{
			Message: err.Error(),
			Woken:   false,
		})
	}

	eval := evaluator.NewRegoEvaluator(h.cfg, mac, h.upsRepo, h.ruleRepo)
	result, err := eval.EvaluateExpressions()
	if err != nil {
		c.Logger().Error("Failed to evaluate expressions", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, UpsWakeResponse{
			Message: err.Error(),
			Woken:   false,
		})
	}

	if !result.Found {
		c.Logger().Error("mac address not found in the config", slog.String("mac", mac.String()))
		return c.JSON(http.StatusConflict, UpsWakeResponse{
			Message: "MAC address not found in the config",
			Woken:   false,
		})
	}

	if !result.Allowed {
		c.Logger().Debug("no rule evaluated to true", slog.String("mac", mac.String()))
		return c.JSON(http.StatusOK, UpsWakeResponse{
			Message: "No rule evaluated to true",
			Woken:   false,
		})
	}

	ts, err := entity.NewTargetServer(
		"API Request",
		result.Target.MAC.String(),
		result.Target.Broadcast,
		15*time.Minute,
		result.Target.Port,
		[]string{},
	)
	if err != nil {
		c.Logger().Error("Failed to create target server", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, UpsWakeResponse{
			Message: fmt.Sprintf("Failed to create target server: %s", err),
			Woken:   false,
		})
	}

	wolClient := wol.NewWoLClient(ts)

	if err = wolClient.Wake(); err != nil {
		c.Logger().Error("Failed to send wake on lan", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, UpsWakeResponse{
			Message: fmt.Sprintf("Failed to send wake on LAN: %s", err),
			Woken:   false,
		})
	}

	c.Logger().Debug("Wake on LAN sent", slog.String("mac", mac.String()))
	return c.JSON(http.StatusOK, UpsWakeResponse{
		Message: "Wake on LAN sent",
		Woken:   true,
	})
}
