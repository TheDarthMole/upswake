package handlers

import (
	"github.com/TheDarthMole/UPSWake/api"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/TheDarthMole/UPSWake/wol"
	"github.com/labstack/echo/v4"
	"log"
	"net"
	"net/http"
)

type ServerHandler struct{}

type WakeServerRequest struct {
	Port      int    `json:"mac" validate:"omitempty,gte=1,lte=65535" default:"9"`
	Broadcast string `json:"broadcast" validate:"omitempty,ip"`
}

func NewServerHandler() *ServerHandler {
	return &ServerHandler{}
}

func (h *ServerHandler) Register(g *echo.Group) {
	g.GET("", HandlerNotImplemented)
	g.GET("/:mac", HandlerNotImplemented)
	g.POST("/:mac/wake", h.WakeServer)
}

func (h *ServerHandler) WakeServer(c echo.Context) error {
	mac := c.Param("mac")
	if _, err := net.ParseMAC(mac); err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorMessage{
			Error:   "Bad Request",
			Message: err.Error(),
		})
	}

	wsRequest := new(WakeServerRequest)
	if err := c.Bind(wsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorMessage{
			Error:   "Bad Request",
			Message: err.Error(),
		})
	}

	if err := c.Validate(wsRequest); err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, api.ErrorMessage{
			Error:   "Bad Request",
			Message: err.Error(),
		})
	}

	if wsRequest.Broadcast != "" {
		wolClient := wol.NewWoLClient(config.TargetServer{
			Mac:       mac,
			Broadcast: wsRequest.Broadcast,
			Port:      wsRequest.Port,
		})

		if err := wolClient.Wake(); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorMessage{
				Error:   "Internal Server Error",
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, "OK")
	}

	broadcasts, err := util.GetAllBroadcastAddresses()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.ErrorMessage{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
	}

	for _, broadcast := range broadcasts {
		wolClient := wol.NewWoLClient(config.TargetServer{
			Mac:       mac,
			Broadcast: broadcast.String(),
			Port:      9,
		})

		if err = wolClient.Wake(); err != nil {
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, api.ErrorMessage{
				Error:   "Internal Server Error",
				Message: err.Error(),
			})
		}
	}

	return c.String(http.StatusOK, "OK")
}
