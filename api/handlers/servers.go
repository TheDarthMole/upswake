package handlers

import (
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/TheDarthMole/UPSWake/wol"
	"github.com/labstack/echo/v4"
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
	g.GET("/:mac", HandlerNotImplemented)
}

// WakeServer godoc
//
//	@Summary		Wake a server using a mac and a broadcast address
//	@Description	Wake a server using Wake on LAN
//	@Tags			servers
//	@Accept			json
//	@Produce		json
//	@Param			wakeServerRequest	body		WakeServerRequest	true	"Wake server request"
//	@Success		201					{object}	Response
//	@Failure		400					{object}	Response
//	@Failure		500					{object}	Response
//	@Router			/servers/wake [post]
func (h *ServerHandler) WakeServer(c echo.Context) error {
	wsRequest := NewWakeServerRequest()
	if err := c.Bind(wsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := c.Validate(wsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	wolClient := wol.NewWoLClient(config.TargetServer{
		Name:      "API Request",
		Mac:       wsRequest.Mac,
		Broadcast: wsRequest.Broadcast,
		Port:      wsRequest.Port,
	})

	if err := wolClient.Wake(); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, Response{Message: "Wake on LAN packet sent"})
}

// BroadcastWakeServer godoc
//
//	@Summary		Wake a server using just a mac (broadcast is enumerated)
//	@Description	Wake a server using Wake on LAN
//	@Tags			servers
//	@Accept			json
//	@Produce		json
//	@Param			broadcastWakeRequest	body		BroadcastWakeRequest	true	"Broadcast wake request"
//	@Success		201						{object}	Response				"Wake on LAN packets successfully sent to all available broadcast addresses"
//	@Failure		400						{object}	Response
//	@Failure		500						{object}	Response
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
		wolClient := wol.NewWoLClient(config.TargetServer{
			Name:      "API Request",
			Mac:       wsRequest.Mac,
			Broadcast: broadcast.String(),
			Port:      wsRequest.Port,
		})

		if err = wolClient.Wake(); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusCreated, Response{Message: "Wake on LAN packets sent to all available broadcast addresses"})
}
