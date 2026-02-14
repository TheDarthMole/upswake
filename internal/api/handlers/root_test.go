package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/labstack/echo/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var cfg = &entity.Config{
	NutServers: []entity.NutServer{
		{
			Name:     "testNUTServer",
			Host:     "127.0.0.1",
			Port:     1234,
			Username: "test-user",
			Password: "test-password",
			Targets: []entity.TargetServer{
				{
					Name:      "testTarget",
					MAC:       "00:00:00:00:00:00",
					Broadcast: "192.168.1.255",
					Port:      9,
					Interval:  "15s",
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
		if err != nil {
			t.Fatalf("could not write file to memfs: %s", err)
		}
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
	h := NewRootHandler(cfg, rulesFS)

	if assert.NoError(t, h.Root(c)) {
		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/swagger/index.html", rec.Header().Get(echo.HeaderLocation))
	}
}

func TestRootHandler_Health(t *testing.T) {
	type fields struct {
		cfg     *entity.Config
		rulesFS afero.Fs
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
				cfg: &entity.Config{NutServers: []entity.NutServer{
					{},
				}},
				rulesFS: newMemFS(t, map[string][]byte{}),
			},
			wantedResponse: wantedResponse{
				body:       `{"message": "name is required"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "cannot-connect-to-server",
			fields: fields{
				cfg: &entity.Config{NutServers: []entity.NutServer{
					{
						Name:     "testNUTServer",
						Host:     "127.0.0.1",
						Port:     1234,
						Username: "test-user",
						Password: "test-password",
					},
				}},
				rulesFS: newMemFS(t, map[string][]byte{}),
			},
			wantedResponse: wantedResponse{
				body:       `{"message": "could not connect to NUT server: connection failed\ndial tcp 127.0.0.1:1234: connect: connection refused"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "empty-config-success",
			fields: fields{
				cfg:     &entity.Config{},
				rulesFS: newMemFS(t, map[string][]byte{}),
			},
			wantedResponse: wantedResponse{
				body:       `{"message": "OK"}`,
				statusCode: http.StatusOK,
			},
		},
		//	TODO: Add tests that check for valid NUT server responses, where the config isn't empty
		// 		    and test when the GetAllBroadcastAddresses fails
	}
	e := echo.New()
	e.Validator = api.NewCustomValidator(t.Context())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := NewRootHandler(tt.fields.cfg, tt.fields.rulesFS)

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
	h := NewRootHandler(cfg, rulesFS)

	g := e.Group("")
	h.Register(g)

	expectedRoutes := []string{"/", "/health", "/swagger/*"}
	lenExpectedRoutes := len(expectedRoutes)
	for _, route := range e.Router().Routes() {
		for i, expected := range expectedRoutes {
			if expected == route.Path {
				expectedRoutes = append(expectedRoutes[:i], expectedRoutes[i+1:]...)
				break
			}
		}
	}

	assert.Lenf(t, e.Router().Routes(), lenExpectedRoutes, "Expected %d routes to be registered", lenExpectedRoutes)
	assert.Equalf(t, []string{}, expectedRoutes, "The following expected routes are missing: %v", expectedRoutes)
}

func TestNewRootHandler(t *testing.T) {
	type args struct {
		cfg     *entity.Config
		rulesFS afero.Fs
	}
	memFS1 := newMemFS(t, map[string][]byte{})
	memFS2 := newMemFS(t, map[string][]byte{
		"test": []byte("test"),
	})
	tests := []struct {
		name string
		args args
		want *RootHandler
	}{
		{
			name: "test-1",
			args: args{
				cfg:     cfg,
				rulesFS: memFS1,
			},
			want: &RootHandler{
				cfg:     cfg,
				rulesFS: memFS1,
			},
		},
		{
			name: "test-2",
			args: args{
				cfg:     cfg,
				rulesFS: memFS2,
			},
			want: &RootHandler{
				cfg:     cfg,
				rulesFS: memFS2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewRootHandler(tt.args.cfg, tt.args.rulesFS), "NewRootHandler(%v, %v)", tt.args.cfg, tt.args.rulesFS)
		})
	}
}
