package handlers

import (
	"github.com/TheDarthMole/UPSWake/internal/config"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/util"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
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
		Port: config.DefaultWoLPort,
	}
}

func NewBroadcastWakeRequest() *BroadcastWakeRequest {
	return &BroadcastWakeRequest{
		Port: config.DefaultWoLPort,
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
//	@Router			/servers/wake [post]
func (h *ServerHandler) WakeServer(c echo.Context) error {
	wsRequest := NewWakeServerRequest()
	if err := c.Bind(wsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := c.Validate(wsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
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
		log.Fatalf("failed to create target server %s", err)
	}
	wolClient := wol.NewWoLClient(ts)

	if err := wolClient.Wake(); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
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
//	@Router			/servers/broadcastwake [post]
func (h *ServerHandler) BroadcastWakeServer(c echo.Context) error {
	wsRequest := NewBroadcastWakeRequest()
	if err := c.Bind(wsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := c.Validate(wsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	broadcasts, err := util.GetAllBroadcastAddresses()
	if err != nil {
		return err
	}

	for _, broadcast := range broadcasts {
		ts, err := entity.NewTargetServer(
			"API Request",
			wsRequest.Mac,
			broadcast.String(),
			"15m",
			wsRequest.Port,
			[]string{},
		)

		if err != nil {
			log.Printf("failed to create new target server %s", err)
		}
		wolClient := wol.NewWoLClient(ts)

		if err = wolClient.Wake(); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusCreated, Response{Message: "Wake on LAN packets sent to all available broadcast addresses"})
}
