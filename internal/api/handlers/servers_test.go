package handlers

import (
	"errors"
	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/util"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	mockValidBroadcastAddressesFunc = func() ([]net.IP, error) {
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}
)

type mockWakeOnLan struct {
	*entity.TargetServer
}

func (m *mockWakeOnLan) Wake() error {
	return errors.New("mock wake error")
}

func TestServerHandler_Register(t *testing.T) {
	e := echo.New()
	h := NewServerHandler()

	g := e.Group("")
	h.Register(g)

	expectedRoutes := []string{"/wake", "/broadcastwake"}
	for _, route := range e.Routes() {
		log.Println("Registered route:", route.Path)
		for i, expected := range expectedRoutes {
			if expected == route.Path {
				expectedRoutes = append(expectedRoutes[:i], expectedRoutes[i+1:]...)
				break
			}
		}
	}

	assert.Equal(t, 2, len(e.Routes()), "Expected 2 routes to be registered")
	assert.Equalf(t, []string{}, expectedRoutes, "The following expected routes are missing: %v", expectedRoutes)
}

func TestNewWakeServerRequest(t *testing.T) {
	newWakeServerRequest := NewWakeServerRequest()
	assert.Equalf(t, &WakeServerRequest{Port: 9}, newWakeServerRequest, "Expected default port to be set to 9")
}

func TestNewBroadcastWakeRequest(t *testing.T) {
	newBroadcastWakeRequest := NewBroadcastWakeRequest()
	assert.Equalf(t, &BroadcastWakeRequest{Port: 9}, newBroadcastWakeRequest, "Expected default port to be set to 9")
}

func TestServerHandler_BroadcastWakeServer(t *testing.T) {
	const validMac = `{"mac": "00:11:22:33:44:55"}`

	type fields struct {
		body                   string
		mockBroadcastAddresses func() ([]net.IP, error)
		mockNewTargetServer    func(name, mac, broadcast, interval string, port int, rules []string) (*entity.TargetServer, error)
		mockNewWoLClient       func(target *entity.TargetServer) *wol.WakeOnLan
	}
	type wantedResponse struct {
		body       string
		statusCode int
	}
	tests := []struct {
		name           string
		fields         fields
		wantedResponse wantedResponse
	}{
		{
			name: "valid_request",
			fields: fields{
				body:                   `{"mac": "00:11:22:33:44:55", "broadcast": "127.0.0.255"}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
				mockNewWoLClient:       wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Wake on LAN packets sent to all available broadcast addresses"}`,
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "missing_mac",
			fields: fields{
				body:                   `{"broadcast": "127.0.0.255"}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
				mockNewWoLClient:       wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Key: 'BroadcastWakeRequest.Mac' Error:Field validation for 'Mac' failed on the 'required' tag"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid_mac",
			fields: fields{
				body:                   `{"mac": "invalid_mac", "broadcast": "127.0.0.255"}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
				mockNewWoLClient:       wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Key: 'BroadcastWakeRequest.Mac' Error:Field validation for 'Mac' failed on the 'mac' tag"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing_broadcast",
			fields: fields{
				body:                   validMac,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
				mockNewWoLClient:       wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Wake on LAN packets sent to all available broadcast addresses"}`,
				statusCode: http.StatusCreated, // broadcast is optional, as it is enumerated from internal configurations
			},
		},
		{
			name: "empty_request",
			fields: fields{
				body:                   `{}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
				mockNewWoLClient:       wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Key: 'BroadcastWakeRequest.Mac' Error:Field validation for 'Mac' failed on the 'required' tag"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "no_broadcast_addresses",
			fields: fields{
				body: validMac,
				mockBroadcastAddresses: func() ([]net.IP, error) {
					return []net.IP{}, nil
				},
				mockNewTargetServer: entity.NewTargetServer,
				mockNewWoLClient:    wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"No broadcast addresses available"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "invalid_port",
			fields: fields{
				body:                   `{"mac": "00:11:22:33:44:55", "port": 70000}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
				mockNewWoLClient:       wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Key: 'BroadcastWakeRequest.Port' Error:Field validation for 'Port' failed on the 'lte' tag"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "mock_get_all_broadcast_addresses_error",
			fields: fields{
				body: validMac,
				mockBroadcastAddresses: func() ([]net.IP, error) {
					return []net.IP{}, errors.New("mock_get_all_broadcast_addresses_error")
				},
				mockNewTargetServer: entity.NewTargetServer,
				mockNewWoLClient:    wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"mock_get_all_broadcast_addresses_error"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "invalid_boradcast_address",
			fields: fields{
				body: validMac,
				mockBroadcastAddresses: func() ([]net.IP, error) {
					return []net.IP{nil}, nil // This will not be used, but we need to return something
				},
				mockNewTargetServer: entity.NewTargetServer,
				mockNewWoLClient:    wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Invalid broadcast address encountered"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "mock_new_target_server_error",
			fields: fields{
				body:                   validMac,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer: func(name, mac, broadcast, interval string, port int, rules []string) (*entity.TargetServer, error) {
					return nil, errors.New("mock_new_target_server_error")
				},
				mockNewWoLClient: wol.NewWoLClient,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"mock_new_target_server_error"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		//{ // TODO: Uncomment this test when the mockWakeOnLan is implemented correctly
		//	name: "mock_new_wol_client_error",
		//	fields: fields{
		//		body:                   validMac,
		//		mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
		//		mockNewTargetServer:    entity.NewTargetServer,
		//		mockNewWoLClient: func(target *entity.TargetServer) *WakeOnLanClient {
		//			return &mockWakeOnLan{}
		//		},
		//	},
		//},
	}
	e := echo.New()
	e.Validator = api.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/broadcastwake", strings.NewReader(tt.fields.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewServerHandler()

			// Mock the functions to return the desired values
			GetAllBroadcastAddresses = tt.fields.mockBroadcastAddresses
			NewTargetServer = tt.fields.mockNewTargetServer

			if assert.NoError(t, h.BroadcastWakeServer(c)) {
				assert.JSONEq(t, tt.wantedResponse.body, rec.Body.String())
				assert.Equal(t, tt.wantedResponse.statusCode, rec.Code)
			}

			GetAllBroadcastAddresses = util.GetAllBroadcastAddresses // Reset to original function after test
			NewTargetServer = entity.NewTargetServer                 // Reset to original function after test
		})
	}

}
