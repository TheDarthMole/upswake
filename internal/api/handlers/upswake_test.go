package handlers

import (
	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/hack-pad/hackpadfs"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUPSWakeHandler_RunWakeEvaluation(t *testing.T) {
	regoAlwaysTrue := newMemFS(t, map[string][]byte{
		"always_true.rego": []byte(`package test
default allow = true`),
	})
	validConfig := &entity.Config{
		NutServers: []entity.NutServer{
			{
				Name:     "test-nut-server",
				Host:     "127.0.0.1",
				Port:     3493,
				Username: "testuser",
				Password: "testpass",
				Targets: []entity.TargetServer{
					{
						Name:      "test-target",
						MAC:       "00:11:22:33:44:55",
						Broadcast: "127.0.0.255",
						Port:      9,
						Interval:  "15m",
						Rules:     []string{"always_true.rego"},
					},
				},
			},
		},
	}
	type fields struct {
		cfg     *entity.Config
		rulesFS hackpadfs.FS
		body    string
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
			name: "invalid_request_body",
			fields: fields{
				cfg:     validConfig,
				rulesFS: regoAlwaysTrue,
				body:    `invalid json`,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"failed to parse request body"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		//{
		//	name: "valid_request",
		//	fields: fields{
		//		cfg:     validConfig,
		//		rulesFS: regoAlwaysTrue,
		//	},
		//	wantedResponse: wantedResponse{
		//		body:       `{"body":"Wake on LAN sent","woken":true}`,
		//		statusCode: http.StatusOK,
		//	},
		//},
	}
	e := echo.New()
	e.Validator = api.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/upswake", strings.NewReader(tt.fields.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewUPSWakeHandler(tt.fields.cfg, tt.fields.rulesFS)

			if assert.NoError(t, h.RunWakeEvaluation(c)) {
				assert.JSONEq(t, tt.wantedResponse.body, rec.Body.String())
				assert.Equal(t, tt.wantedResponse.statusCode, rec.Code)
			}

		})
	}
}
