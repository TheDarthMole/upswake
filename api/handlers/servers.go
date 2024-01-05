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
	Port      int    `json:"port" validate:"required,gte=1,lte=65535"`
	Broadcast string `json:"broadcast" validate:"required,ip"`
	Mac       string `json:"mac" validate:"required,mac"`
}

type BroadcastWakeRequest struct {
	Port int    `json:"port" validate:"required,gte=1,lte=65535"`
	Mac  string `json:"mac" validate:"required,mac"`
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
		return err
	}
	return c.JSON(http.StatusCreated, Response{Message: "Wake on LAN packet sent"})
}

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
