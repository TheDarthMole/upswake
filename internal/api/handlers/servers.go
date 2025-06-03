package handlers

import (
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/util"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/labstack/echo/v4"
	"net/http"
)

var (
	GetAllBroadcastAddresses = util.GetAllBroadcastAddresses
	NewTargetServer          = entity.NewTargetServer
	NewWoLClient             = wol.NewWoLClient
)

type ServerHandler struct{}

type WakeServerRequest struct {
	Port      int    `json:"port" validate:"gte=1,lte=65535" example:"9"`
	Broadcast string `json:"broadcast" validate:"required,ip" example:"192.168.1.13"`
	Mac       string `json:"mac" validate:"required,mac" example:"00:11:22:33:44:55"`
}

type BroadcastWakeRequest struct {
	Port int    `json:"port" validate:"gte=1,lte=65535" example:"9"`
	Mac  string `json:"mac" validate:"required,mac" example:"00:11:22:33:44:55"`
}

func NewWakeServerRequest() *WakeServerRequest {
	return &WakeServerRequest{
		Port: entity.DefaultWoLPort,
	}
}

func NewBroadcastWakeRequest() *BroadcastWakeRequest {
	return &BroadcastWakeRequest{
		Port: entity.DefaultWoLPort,
	}
}

func NewServerHandler() *ServerHandler {
	return &ServerHandler{}
}

func (h *ServerHandler) Register(g *echo.Group) {
	g.POST("/wake", h.WakeServer)
	g.POST("/broadcastwake", h.BroadcastWakeServer)
}

// WakeServer godoc
//
//	@Summary		Wake a server using a mac and a broadcast address
//	@Description	Wake a server using Wake on LAN using the mac and broadcast address provided
//	@Tags			servers
//	@Accept			json
//	@Produce		json
//	@Param			wakeServerRequest	body		WakeServerRequest	true	"Wake server request"
//	@Success		201					{object}	Response			"Wake on LAN packet sent"
//	@Failure		400					{object}	Response			"Input validation failed"
//	@Failure		500					{object}	Response			"Wake on LAN packet failed to send"
//	@Router			/api/servers/wake [post]
func (h *ServerHandler) WakeServer(c echo.Context) error {
	wsRequest := NewWakeServerRequest()
	if err := c.Bind(wsRequest); err != nil {
		c.Logger().Errorf("failed to bind wake server request %s", err)
		return c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
	}

	if err := c.Validate(wsRequest); err != nil {
		c.Logger().Errorf("failed to validate wake server request %s", err)
		return c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
	}

	ts, err := entity.NewTargetServer(
		"API Request",
		wsRequest.Mac,
		wsRequest.Broadcast,
		"15m",
		wsRequest.Port,
		[]string{},
	)
	if err != nil {
		c.Logger().Fatalf("failed to create target server %s", err)
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	wolClient := wol.NewWoLClient(ts)

	if err = wolClient.Wake(); err != nil {
		c.Logger().Fatalf("failed to send wake on lan %s", err)
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	c.Logger().Debugf("wake on lan packet sent to %s", wsRequest.Mac)
	return c.JSON(http.StatusCreated, Response{Message: "Wake on LAN packet sent"})
}

// BroadcastWakeServer godoc
//
//	@Summary		Wake a server using just a mac
//	@Description	Wake a server using Wake on LAN by using the mac and enumerating all available broadcast addresses
//	@Tags			servers
//	@Accept			json
//	@Produce		json
//	@Param			broadcastWakeRequest	body		BroadcastWakeRequest	true	"Broadcast wake request"
//	@Success		201						{object}	Response				"Wake on LAN packets successfully sent to all available broadcast addresses"
//	@Failure		400						{object}	Response				"Input validation failed"
//	@Failure		500						{object}	Response				"Wake on LAN packet failed to send"
//	@Router			/api/servers/broadcastwake [post]
func (h *ServerHandler) BroadcastWakeServer(c echo.Context) error {
	wsRequest := NewBroadcastWakeRequest()
	if err := c.Bind(wsRequest); err != nil {
		c.Logger().Errorf("failed to bind wake server request: %s", err)
		return c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
	}

	if err := c.Validate(wsRequest); err != nil {
		c.Logger().Errorf("failed to validate wake server request: %s", err)
		return c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
	}
	broadcasts, err := GetAllBroadcastAddresses()
	if err != nil {
		c.Logger().Errorf("failed to get broadcast addresses, %s", err)
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	if broadcasts == nil || len(broadcasts) == 0 {
		c.Logger().Errorf("no broadcast addresses available, got %v", broadcasts)
		return c.JSON(http.StatusInternalServerError, Response{Message: "No broadcast addresses available"})
	}

	for _, broadcast := range broadcasts {
		if broadcast == nil {
			c.Logger().Errorf("invalid broadcast address, got %v", broadcast)
			return c.JSON(http.StatusInternalServerError, Response{Message: "Invalid broadcast address encountered"})
		}

		c.Logger().Debugf("sending wake on lan to %s:%d", broadcast.String(), wsRequest.Port)
		ts, err := NewTargetServer(
			"API Request",
			wsRequest.Mac,
			broadcast.String(),
			"15m",
			wsRequest.Port,
			[]string{},
		)

		if err != nil {
			c.Logger().Infof("failed to create new target server %s", err)
			return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		wolClient := NewWoLClient(ts)

		if err = wolClient.Wake(); err != nil {
			c.Logger().Errorf("failed to send wake on lan %s", err)
			return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
	}
	return c.JSON(http.StatusCreated, Response{Message: "Wake on LAN packets sent to all available broadcast addresses"})
}
