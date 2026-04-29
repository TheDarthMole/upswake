package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/rules"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUPSRepo struct {
	err  error
	json string
}

func (m *mockUPSRepo) GetJSON(_ *entity.NutServer) (string, error) {
	return m.json, m.err
}

func TestUPSWakeHandler_RunWakeEvaluation(t *testing.T) {
	regoAlwaysTrue := newMemFS(t, map[string][]byte{
		"always_true.rego": []byte(`package upswake
default wake := true`),
	})
	regoAlwaysFalse := newMemFS(t, map[string][]byte{
		"always_true.rego": []byte(`package upswake
default wake := false`),
	})

	alwaysTrueRuleRepo, err := rules.NewPreparedRepository(regoAlwaysTrue)
	require.NoError(t, err)

	alwaysFalseRuleRepo, err := rules.NewPreparedRepository(regoAlwaysFalse)
	require.NoError(t, err)

	upsRepo := &mockUPSRepo{
		json: `[{"Name":"test-ups","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`,
		err:  nil,
	}

	validConfig := &entity.Config{
		NutServers: []*entity.NutServer{
			{
				Name:     "test-nut-server",
				Host:     "127.0.0.1",
				Port:     3493,
				Username: "upsmon",
				Password: "upsmon",
				Targets: []*entity.TargetServer{
					{
						Name:      "test-target",
						MAC:       "00:11:22:33:44:55",
						Broadcast: "127.0.0.255",
						Port:      9,
						Interval:  15 * time.Minute,
						Rules:     []string{"always_true.rego"},
					},
				},
			},
		},
	}
	invalidConfig := &entity.Config{
		NutServers: []*entity.NutServer{
			{
				Name:     "test-nut-server",
				Host:     "127.0.0.1",
				Port:     3493,
				Username: "upsmon",
				Password: "upsmon",
				Targets: []*entity.TargetServer{
					{
						Name:      "test-target",
						MAC:       "00:11:22:33:44:55",
						Broadcast: "777.666.555.444",
						Port:      9,
						Interval:  15 * time.Minute,
						Rules:     []string{"always_true.rego"},
					},
				},
			},
		},
	}

	type fields struct {
		cfg      *entity.Config
		upsRepo  repository.UPSRepository
		ruleRepo repository.RuleRepository
		body     string
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
				cfg:      validConfig,
				upsRepo:  upsRepo,
				ruleRepo: alwaysTrueRuleRepo,
				body:     `invalid json`,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"failed to parse request body","woken":false}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "valid_request",
			fields: fields{
				cfg:      validConfig,
				upsRepo:  upsRepo,
				ruleRepo: alwaysTrueRuleRepo,
				body:     `{"mac":"00:11:22:33:44:55"}`,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Wake on LAN sent","woken":true}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "rule_evaluates_to_false",
			fields: fields{
				cfg:      validConfig,
				upsRepo:  upsRepo,
				ruleRepo: alwaysFalseRuleRepo,
				body:     `{"mac":"00:11:22:33:44:55"}`,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"No rule evaluated to true","woken":false}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "mac_not_in_config",
			fields: fields{
				cfg:      validConfig,
				upsRepo:  upsRepo,
				ruleRepo: alwaysTrueRuleRepo,
				body:     `{"mac":"99:11:22:33:44:44"}`,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"MAC address not found in the config","woken":false}`,
				statusCode: http.StatusConflict,
			},
		},
		{
			name: "invalid_broadcast_address",
			fields: fields{
				cfg:      invalidConfig,
				upsRepo:  upsRepo,
				ruleRepo: alwaysTrueRuleRepo,
				body:     `{"mac":"00:11:22:33:44:55"}`,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"Failed to create target server: broadcast is invalid, must be an IP address","woken":false}`,
				statusCode: http.StatusInternalServerError,
			},
		},
	}
	e := echo.New()
	e.Validator = api.NewCustomValidator(t.Context())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/upswake", strings.NewReader(tt.fields.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewUPSWakeHandler(tt.fields.cfg, tt.fields.upsRepo, tt.fields.ruleRepo)

			if assert.NoError(t, h.RunWakeEvaluation(c)) {
				assert.JSONEq(t, tt.wantedResponse.body, rec.Body.String())
				assert.Equal(t, tt.wantedResponse.statusCode, rec.Code)
			}
		})
	}
}
