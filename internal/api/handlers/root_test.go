package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/labstack/echo/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingUPSRepo struct {
	err   error
	json  string
	calls atomic.Int32
}

func (r *countingUPSRepo) GetJSON(_ *entity.NutServer) (string, error) {
	r.calls.Add(1)
	return r.json, r.err
}

var cfg = &entity.Config{
	NutServers: []*entity.NutServer{
		{
			Name:     "testNUTServer",
			Host:     "127.0.0.1",
			Port:     1234,
			Username: "test-user",
			Password: "test-password",
			Targets: []*entity.TargetServer{
				{
					Name:      "testTarget",
					MAC:       "00:00:00:00:00:00",
					Broadcast: "192.168.1.255",
					Port:      9,
					Interval:  15 * time.Second,
					Rules:     nil,
				},
			},
		},
	},
}

func newMemFS(t *testing.T, data map[string][]byte) afero.Fs {
	t.Helper()
	memfs := afero.NewMemMapFs()

	for x := range data {
		err := afero.WriteFile(memfs, x, data[x], 0o644)
		require.NoErrorf(t, err, "error writing to in-memory filesystem: %s", err)
	}
	return memfs
}

func TestRootHandlerRoot(t *testing.T) {
	e := echo.New()
	e.Validator = api.NewCustomValidator(t.Context())
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	rulesFS := newMemFS(t, map[string][]byte{})
	h := NewRootHandler(cfg, rulesFS, &countingUPSRepo{})

	if assert.NoError(t, h.Root(c)) {
		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/swagger/index.html", rec.Header().Get(echo.HeaderLocation))
	}
}

func TestRootHandler_Health(t *testing.T) {
	type fields struct {
		cfg     *entity.Config
		rulesFS afero.Fs
		upsRepo repository.UPSRepository
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
			name: "test-invalid-config",
			fields: fields{
				cfg: &entity.Config{NutServers: []*entity.NutServer{
					{},
				}},
				rulesFS: newMemFS(t, map[string][]byte{}),
				upsRepo: &countingUPSRepo{
					err:  nil,
					json: `[{"Name":"ups1"}]`,
				},
			},
			wantedResponse: wantedResponse{
				body:       `{"message": "name is required"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "cannot-connect-to-server",
			fields: fields{
				cfg: &entity.Config{NutServers: []*entity.NutServer{
					{
						Name:     "testNUTServer",
						Host:     "127.0.0.1",
						Port:     1234,
						Username: "test-user",
						Password: "test-password",
					},
				}},
				rulesFS: newMemFS(t, map[string][]byte{}),
				upsRepo: &countingUPSRepo{
					err:  errors.New("could not connect to NUT server: dial tcp 127.0.0.1:1234: connect: connection refused"),
					json: "",
				},
			},
			wantedResponse: wantedResponse{
				body:       `{"message": "could not connect to NUT server: dial tcp 127.0.0.1:1234: connect: connection refused"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "empty-config-success",
			fields: fields{
				cfg:     &entity.Config{},
				rulesFS: newMemFS(t, map[string][]byte{}),
				upsRepo: &countingUPSRepo{
					json: `[{"Name":"ups1"}]`,
				},
			},
			wantedResponse: wantedResponse{
				body:       `{"message": "OK"}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "valid-healthy-config",
			fields: fields{
				cfg: &entity.Config{NutServers: []*entity.NutServer{
					{
						Name:     "testNUTServer",
						Host:     "127.0.0.1",
						Port:     entity.DefaultNUTServerPort,
						Username: "test-user",
						Password: "test-password",
						Targets: []*entity.TargetServer{
							{
								Name:      "testTarget",
								MAC:       "00:00:00:00:00:00",
								Broadcast: "127.0.0.255",
								Port:      9,
								Interval:  15 * time.Second,
								Rules:     nil,
							},
						},
					},
				}},
				rulesFS: newMemFS(t, map[string][]byte{}),
				upsRepo: &countingUPSRepo{
					json: `[{"Name":"ups1"}]`,
				},
			},
			wantedResponse: wantedResponse{
				body:       `{"message": "OK"}`,
				statusCode: http.StatusOK,
			},
		},
		//	TODO: Add tests that check for where the config isn't empty
		// 		    and test when the GetAllBroadcastAddresses fails
	}
	e := echo.New()
	e.Validator = api.NewCustomValidator(t.Context())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewRootHandler(tt.fields.cfg, tt.fields.rulesFS, tt.fields.upsRepo)

			if assert.NoError(t, h.Health(c)) {
				assert.Equal(t, tt.wantedResponse.statusCode, rec.Code)
				assert.JSONEq(t, tt.wantedResponse.body, rec.Body.String())
			}
		})
	}
}

func TestRootHandler_Register(t *testing.T) {
	e := echo.New()
	e.Validator = api.NewCustomValidator(t.Context())
	rulesFS := newMemFS(t, map[string][]byte{})
	h := NewRootHandler(cfg, rulesFS, &countingUPSRepo{})

	g := e.Group("")
	h.Register(g)

	expectedRoutes := echo.Routes{
		{
			Name:   "GET:/",
			Path:   "/",
			Method: "GET",
		},
		{
			Name:   "GET:/health",
			Path:   "/health",
			Method: "GET",
		},
		{
			Name:       "GET:/swagger/*",
			Path:       "/swagger/*",
			Method:     "GET",
			Parameters: []string{"*"},
		},
	}

	assert.Equal(t, expectedRoutes, e.Router().Routes())
}

func TestNewRootHandler(t *testing.T) {
	type args struct {
		cfg     *entity.Config
		rulesFS afero.Fs
		upsRepo repository.UPSRepository
	}
	emptyFS := newMemFS(t, map[string][]byte{})
	ruleOneFS := newMemFS(t, map[string][]byte{
		"rule1.rego": []byte(`package upswake
default wake := true`),
	})
	tests := []struct {
		args args
		want *RootHandler
		name string
	}{
		{
			name: "empty rules filesystem",
			args: args{
				cfg:     cfg,
				rulesFS: emptyFS,
				upsRepo: &countingUPSRepo{
					json: `[{"Name":"ups1"}]`,
				},
			},
			want: &RootHandler{
				cfg:     cfg,
				rulesFS: emptyFS,
				upsRepo: &countingUPSRepo{
					json: `[{"Name":"ups1"}]`,
				},
			},
		},
		{
			name: "one rule in filesystem",
			args: args{
				cfg: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name:     "testNUTServer",
							Host:     "127.0.0.1",
							Port:     1234,
							Username: "test-user",
							Password: "test-password",
							Targets: []*entity.TargetServer{
								{
									Name:      "testTarget",
									MAC:       "00:00:00:00:00:00",
									Broadcast: "192.168.1.255",
									Port:      9,
									Interval:  15 * time.Second,
									Rules:     []string{"rule1.rego"},
								},
							},
						},
					},
				},
				rulesFS: ruleOneFS,
			},
			want: &RootHandler{
				cfg: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name:     "testNUTServer",
							Host:     "127.0.0.1",
							Port:     1234,
							Username: "test-user",
							Password: "test-password",
							Targets: []*entity.TargetServer{
								{
									Name:      "testTarget",
									MAC:       "00:00:00:00:00:00",
									Broadcast: "192.168.1.255",
									Port:      9,
									Interval:  15 * time.Second,
									Rules:     []string{"rule1.rego"},
								},
							},
						},
					},
				},
				rulesFS: ruleOneFS,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewRootHandler(tt.args.cfg, tt.args.rulesFS, tt.args.upsRepo), "NewRootHandler(%v, %v)", tt.args.cfg, tt.args.rulesFS)
		})
	}
}
