package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/rules"
	mockrepository "github.com/TheDarthMole/UPSWake/internal/mocks"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mockUPSRepo struct {
	err  error
	json string
}

func (m *mockUPSRepo) GetJSON(_ *entity.NutServer) (string, error) {
	return m.json, m.err
}

func TestUPSWakeHandler_RunWakeEvaluation(t *testing.T) {
	validMac, err := entity.NewMacAddress("00:11:22:33:44:55")
	require.NoError(t, err)

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

	const validJSON = `[{"Name":"test-ups","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`

	upsRepo := mockrepository.NewMockUPSRepository(gomock.NewController(t))
	upsRepo.EXPECT().GetJSON(gomock.Any()).Return(validJSON, nil).AnyTimes()

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
						MAC:       validMac,
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
						MAC:       validMac,
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
		{
			name: "failing_ups_repo",
			fields: fields{
				cfg: validConfig,
				upsRepo: &mockUPSRepo{
					json: "",
					err:  errors.New("failing rule"),
				},
				body: `{"mac":"00:11:22:33:44:55"}`,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"failing rule","woken":false}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "invalid_mac_address",
			fields: fields{
				cfg:     validConfig,
				upsRepo: upsRepo,
				body:    `{"mac":"invalid mac address"}`,
			},
			wantedResponse: wantedResponse{
				body:       `{"message":"MAC address is invalid","woken":false}`,
				statusCode: http.StatusBadRequest,
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

func TestUPSWakeHandler_Register(t *testing.T) {
	config := &entity.Config{}

	e := echo.New()

	mock := gomock.NewController(t)

	upsRepo := mockrepository.NewMockUPSRepository(mock)
	ruleRepo := mockrepository.NewMockRuleRepository(mock)

	h := NewUPSWakeHandler(config, upsRepo, ruleRepo)
	h.Register(e.Group("/"))

	expectedRoutes := echo.Routes{
		{
			Name:   "GET:/",
			Path:   "/",
			Method: "GET",
		},
		{
			Name:   "POST:/",
			Path:   "/",
			Method: "POST",
		},
	}

	assert.Equal(t, expectedRoutes, e.Router().Routes())
}

func TestUPSWakeHandler_ListNutServerMappings(t *testing.T) {
	validMac, err := entity.NewMacAddress("00:11:22:33:44:55")
	require.NoError(t, err)
	type fields struct {
		cfg *entity.Config
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
			name: "default config",
			fields: fields{
				cfg: entity.CreateDefaultConfig(),
			},
			wantedResponse: wantedResponse{
				body:       `[{"name":"NUT Server 1","host":"192.168.1.13","username":"username","password":"********","targets":[{"name":"NAS 1","mac":"00:00:00:00:00:00","broadcast":"192.168.1.255","rules":["80percentOn.rego"],"interval":"15m0s","port":9}],"port":3493}]`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "nut server with no targets",
			fields: fields{
				cfg: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name:     "NUT Server 1",
							Host:     "127.0.0.1",
							Username: "test",
							Password: "",
							Targets:  nil,
							Port:     1337,
						},
					},
				},
			},
			wantedResponse: wantedResponse{
				body:       `[{"name":"NUT Server 1","host":"127.0.0.1","username":"test","password":"********","targets":[],"port":1337}]`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "nut server with one target",
			fields: fields{
				cfg: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name:     "NUT Server 1",
							Host:     "127.0.0.1",
							Username: "test",
							Password: "",
							Targets: []*entity.TargetServer{
								{
									Name:      "NAS 1",
									MAC:       validMac,
									Broadcast: "127.0.0.255",
									Rules:     []string{"test.rego"},
									Interval:  15 * time.Minute,
									Port:      1337,
								},
							},
							Port: 1337,
						},
					},
				},
			},
			wantedResponse: wantedResponse{
				body:       `[{"name":"NUT Server 1","host":"127.0.0.1","username":"test","password":"********","targets":[{"name":"NAS 1","mac":"00:11:22:33:44:55","broadcast":"127.0.0.255","rules":["test.rego"],"interval":"15m0s","port":1337}],"port":1337}]`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "nut server with two targets",
			fields: fields{
				cfg: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name:     "NUT Server 1",
							Host:     "127.0.0.1",
							Username: "test",
							Password: "",
							Targets: []*entity.TargetServer{
								{
									Name:      "NAS 1",
									MAC:       validMac,
									Broadcast: "127.0.0.255",
									Rules:     []string{"test.rego"},
									Interval:  15 * time.Minute,
									Port:      1337,
								},
								{
									Name:      "NAS 2",
									MAC:       validMac,
									Broadcast: "127.0.1.255",
									Rules:     []string{"test1.rego"},
									Interval:  10 * time.Minute,
									Port:      1338,
								},
							},
							Port: 1337,
						},
					},
				},
			},
			wantedResponse: wantedResponse{
				body:       `[{"name":"NUT Server 1","host":"127.0.0.1","username":"test","password":"********","targets":[{"name":"NAS 1","mac":"00:11:22:33:44:55","broadcast":"127.0.0.255","rules":["test.rego"],"interval":"15m0s","port":1337},{"name":"NAS 2","mac":"00:11:22:33:44:55","broadcast":"127.0.1.255","rules":["test1.rego"],"interval":"10m0s","port":1338}],"port":1337}]`,
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := gomock.NewController(t)

			upsRepo := mockrepository.NewMockUPSRepository(mock)
			ruleRepo := mockrepository.NewMockRuleRepository(mock)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewUPSWakeHandler(tt.fields.cfg, upsRepo, ruleRepo)

			if assert.NoError(t, h.ListNutServerMappings(c)) {
				assert.JSONEq(t, tt.wantedResponse.body, rec.Body.String())
				assert.Equal(t, tt.wantedResponse.statusCode, rec.Code)
			}
		})
	}
}
