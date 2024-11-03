package handlers

import (
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/evaluator"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/hack-pad/hackpadfs"
	"github.com/labstack/echo/v4"
	"net/http"
)

type UPSWakeHandler struct {
	cfg     *entity.Config
	rulesFS hackpadfs.FS
}

type macAddress struct {
	Mac string `json:"mac" example:"00:11:22:33:44:55"`
}

type upsWakeResponse struct {
	Message string `json:"message" example:"Wake on LAN sent"`
	Woken   bool   `json:"woken" example:"true"`
}

func NewUPSWakeHandler(cfg *entity.Config, rulesFS hackpadfs.FS) *UPSWakeHandler {
	return &UPSWakeHandler{
		cfg:     cfg,
		rulesFS: rulesFS,
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
//	@Success		200	{object}	config.Config
//	@Router			/upswake [get]
func (h *UPSWakeHandler) ListNutServerMappings(c echo.Context) error {
	nutServers := h.cfg.NutServers
	// Don't leak passwords
	for i, nutServer := range nutServers {
		nutServer.Password = "********"
		nutServers[i] = nutServer
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
//	@Param			macAddress	body		macAddress	true	"MAC address"
//	@Success		200			{object}	Response "Wake on LAN sent"
//	@Success		304			{object}	Response "No rule evaluated to true"
//	@Failure		400			{object}	Response "Bad request"
//	@Failure		404			{object}	Response "MAC address not found in the config"
//	@Failure		500			{object}	Response "Internal server error"
//	@Router			/upswake	[post]
func (h *UPSWakeHandler) RunWakeEvaluation(c echo.Context) error {
	mac := &macAddress{}
	if err := c.Bind(mac); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	eval := evaluator.NewRegoEvaluator(h.cfg, mac.Mac, h.rulesFS)
	result, err := eval.EvaluateExpressions()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	if !result.Found {
		return c.JSON(http.StatusConflict, Response{Message: "MAC address not found in the config"})
	}

	if !result.Allowed {
		return c.JSON(http.StatusOK, upsWakeResponse{
			Message: "No rule evaluated to true",
			Woken:   false,
		})
	}

	ts, err := entity.NewTargetServer(
		"API Request",
		result.Target.MAC,
		result.Target.Broadcast,
		"15m",
		result.Target.Port,
		[]string{},
	)
	if err != nil {
		return err
	}

	wolClient := wol.NewWoLClient(ts)

	if err := wolClient.Wake(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, upsWakeResponse{
		Message: "Wake on Lan sent",
		Woken:   true,
	})
}
