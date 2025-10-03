package handlers

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/util"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var mockValidBroadcastAddressesFunc = func() ([]net.IP, error) {
	return []net.IP{net.ParseIP("127.0.0.1")}, nil
}

const validMacBroadcast = `{"mac": "00:11:22:33:44:55", "broadcast": "127.0.0.255"}`

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

	assert.Len(t, e.Routes(), 2, "Expected 2 routes to be registered")
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
				body:                   validMacBroadcast,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + BroadcastWoLSentMessage + `"}`,
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "missing_mac",
			fields: fields{
				body:                   `{"broadcast": "127.0.0.255"}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid_mac",
			fields: fields{
				body:                   `{"mac": "invalid_mac", "broadcast": "127.0.0.255"}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing_broadcast",
			fields: fields{
				body:                   validMac,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + BroadcastWoLSentMessage + `"}`,
				statusCode: http.StatusCreated, // broadcast is optional, as it is enumerated from internal configurations
			},
		},
		{
			name: "empty_request",
			fields: fields{
				body:                   `{}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
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
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorBroadcastAddress.Error() + `"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "invalid_port",
			fields: fields{
				body:                   `{"mac": "00:11:22:33:44:55", "port": 70000}`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
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
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorBroadcastAddress.Error() + `"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "invalid_broadcast_address",
			fields: fields{
				body: validMac,
				mockBroadcastAddresses: func() ([]net.IP, error) {
					return []net.IP{nil}, nil // This will not be used, but we need to return something
				},
				mockNewTargetServer: entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorBroadcastAddress.Error() + `"}`,
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
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorCreatingTargetServer.Error() + `"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "bind_error",
			fields: fields{
				body:                   `this is not valid json`,
				mockBroadcastAddresses: mockValidBroadcastAddressesFunc,
				mockNewTargetServer:    entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorBindingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
	}
	e := echo.New()
	e.Validator = api.NewCustomValidator(t.Context())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/broadcastwake", strings.NewReader(tt.fields.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			// Mock the functions to return the desired values
			GetAllBroadcastAddresses = tt.fields.mockBroadcastAddresses
			NewTargetServer = tt.fields.mockNewTargetServer

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewServerHandler()

			if assert.NoError(t, h.BroadcastWakeServer(c)) {
				assert.JSONEq(t, tt.wantedResponse.body, rec.Body.String())
				assert.Equal(t, tt.wantedResponse.statusCode, rec.Code)
			}

			GetAllBroadcastAddresses = util.GetAllBroadcastAddresses // Reset to original function after test
			NewTargetServer = entity.NewTargetServer                 // Reset to original function after test
		})
	}
}

func TestServerHandler_WakeServer(t *testing.T) {
	type fields struct {
		body                string
		mockNewTargetServer func(name, mac, broadcast, interval string, port int, rules []string) (*entity.TargetServer, error)
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
			name: "bind_error",
			fields: fields{
				body:                `this is not valid json`,
				mockNewTargetServer: entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorBindingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "valid_request",
			fields: fields{
				body:                validMacBroadcast,
				mockNewTargetServer: entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + WoLSentMessage + `"}`,
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "missing_mac",
			fields: fields{
				body:                `{"broadcast": "127.0.0.255"}`,
				mockNewTargetServer: entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid_mac",
			fields: fields{
				body:                `{"mac": "invalid_mac", "broadcast": "127.0.0.255"}`,
				mockNewTargetServer: entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing_broadcast",
			fields: fields{
				body:                `{"mac": "00:11:22:33:44:55"}`,
				mockNewTargetServer: entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty_request",
			fields: fields{
				body:                `{}`,
				mockNewTargetServer: entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid_port",
			fields: fields{
				body:                `{"mac": "00:11:22:33:44:55", "broadcast": "127.0.0.255", "port": 70000}`,
				mockNewTargetServer: entity.NewTargetServer,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorValidatingRequest.Error() + `"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "mock_new_target_server_error",
			fields: fields{
				body: validMacBroadcast,
				mockNewTargetServer: func(name, mac, broadcast, interval string, port int, rules []string) (*entity.TargetServer, error) {
					return nil, errors.New("mock_new_target_server_error")
				},
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"` + ErrorCreatingTargetServer.Error() + `"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
	}
	e := echo.New()
	e.Validator = api.NewCustomValidator(t.Context())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/wake", strings.NewReader(tt.fields.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			// Mock the functions to return the desired values
			NewTargetServer = tt.fields.mockNewTargetServer

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewServerHandler()

			if assert.NoError(t, h.WakeServer(c)) {
				assert.JSONEq(t, tt.wantedResponse.body, rec.Body.String())
				assert.Equal(t, tt.wantedResponse.statusCode, rec.Code)
			}

			NewTargetServer = entity.NewTargetServer // Reset to original function after test
		})
	}
}
