package handlers

import (
	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	type fields struct {
		body   io.Reader
		method string
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
				method: http.MethodPost,
				body:   strings.NewReader(`{"mac": "00:11:22:33:44:55", "broadcast": "127.0.0.255"}`),
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Wake on LAN packets sent to all available broadcast addresses"}`,
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "missing_mac",
			fields: fields{
				method: http.MethodPost,
				body:   strings.NewReader(`{"broadcast": "127.0.0.255"}`),
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Key: 'BroadcastWakeRequest.Mac' Error:Field validation for 'Mac' failed on the 'required' tag"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid_mac",
			fields: fields{
				method: http.MethodPost,
				body:   strings.NewReader(`{"mac": "invalid_mac", "broadcast": "127.0.0.255"}`),
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Key: 'BroadcastWakeRequest.Mac' Error:Field validation for 'Mac' failed on the 'mac' tag"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing_broadcast",
			fields: fields{
				method: http.MethodPost,
				body:   strings.NewReader(`{"mac": "00:11:22:33:44:55"}`),
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Wake on LAN packets sent to all available broadcast addresses"}`,
				statusCode: http.StatusCreated, // broadcast is optional, as it is enumerated from internal configurations
			},
		},
		{
			name: "empty_request",
			fields: fields{
				method: http.MethodPost,
				body:   strings.NewReader(`{}`),
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Key: 'BroadcastWakeRequest.Mac' Error:Field validation for 'Mac' failed on the 'required' tag"}`,
				statusCode: http.StatusBadRequest,
			},
		},
	}
	e := echo.New()
	e.Validator = api.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.fields.method, "/broadcastwake?mac=00:11:22:33:44:55", tt.fields.body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewServerHandler()

			if assert.NoError(t, h.BroadcastWakeServer(c)) {
				assert.JSONEq(t, tt.wantedResponse.body, rec.Body.String())
				assert.Equal(t, tt.wantedResponse.statusCode, rec.Code)
			}
		})
	}
}
