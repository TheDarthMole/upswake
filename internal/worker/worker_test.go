package worker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkerPool(t *testing.T) {
	type args struct {
		ctx       context.Context
		config    *entity.Config
		tlsConfig *tls.Config
		logger    *slog.Logger
	}
	tests := []struct {
		args           args
		wantErr        error
		name           string
		wantNumWorkers int
	}{
		{
			name: "one server with one target",
			args: args{
				ctx: context.Background(),
				config: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name: "Test Server",
							Targets: []*entity.TargetServer{
								{Name: "Test Target"},
							},
						},
					},
				},
				logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{})),
			},
			wantNumWorkers: 1,
		},
		{
			name: "two servers with 6 total targets",
			args: args{
				ctx: context.Background(),
				config: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name: "Test Server 1",
							Targets: []*entity.TargetServer{
								{Name: "Test Target 1"},
								{Name: "Test Target 2"},
								{Name: "Test Target 3"},
								{Name: "Test Target 4"},
							},
						},
						{
							Name: "Test Server 2",
							Targets: []*entity.TargetServer{
								{Name: "Test Target 1"},
								{Name: "Test Target 2"},
							},
						},
					},
				},
				logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{})),
			},
			wantNumWorkers: 6,
		},
		{
			name: "no servers",
			args: args{
				ctx: context.Background(),
				config: &entity.Config{
					NutServers: []*entity.NutServer{},
				},
				logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{})),
			},
			wantNumWorkers: 0,
		},
		{
			name: "one server no targets",
			args: args{
				ctx: context.Background(),
				config: &entity.Config{
					NutServers: []*entity.NutServer{
						{Name: "Test server"},
					},
				},
				logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{})),
			},
			wantNumWorkers: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkerPool(tt.args.ctx, tt.args.config, tt.args.tlsConfig, tt.args.logger, "")
			assert.ErrorIs(t, err, tt.wantErr)
			if err != nil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Len(t, got.workers, tt.wantNumWorkers)
			assert.NotNil(t, got.wg)
		})
	}
}

func TestPool_Start(t *testing.T) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	errorStrings := []string{
		`"level","ERROR"`,
		`"level","error"`,
	}

	type fields struct {
		config *entity.Config
	}
	type attestations struct {
		wantLogOutputs    []string
		notWantLogOutputs []string
	}
	tests := []struct {
		name         string
		fields       fields
		attestations attestations
	}{
		{
			name: "one server with one target",
			fields: fields{
				config: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name: "Test Server",
							Targets: []*entity.TargetServer{
								{
									Name: "Test Target",
									MAC:  "00:11:22:33:44:55",
									Rules: []string{
										"test.rego",
									},
									Interval: 100 * time.Millisecond,
								},
							},
						},
					},
				},
			},
			attestations: attestations{
				wantLogOutputs: []string{
					`"worker_name":"Test Target"`,
					`"Gracefully stopping worker","type":"serveJob","worker_name":"Test Target"`,
				},
				notWantLogOutputs: errorStrings,
			},
		},
		{
			name: "two servers with 6 total targets",
			fields: fields{
				config: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name: "Test Server 1",
							Targets: []*entity.TargetServer{
								{Name: "Test Target 1", MAC: "00:11:22:33:44:55", Interval: 100 * time.Millisecond},
								{Name: "Test Target 2", MAC: "11:11:22:33:44:55", Interval: 100 * time.Millisecond},
								{Name: "Test Target 3", MAC: "22:11:22:33:44:55", Interval: 100 * time.Millisecond},
								{Name: "Test Target 4", MAC: "33:11:22:33:44:55", Interval: 100 * time.Millisecond},
							},
						},
						{
							Name: "Test Server 2",
							Targets: []*entity.TargetServer{
								{Name: "Test Target 5", MAC: "44:11:22:33:44:55", Interval: 100 * time.Millisecond},
								{Name: "Test Target 6", MAC: "55:11:22:33:44:55", Interval: 100 * time.Millisecond},
							},
						},
					},
				},
			},
			attestations: attestations{
				wantLogOutputs: []string{
					`"worker_name":"Test Target 1"`,
					`"Gracefully stopping worker","type":"serveJob","worker_name":"Test Target 1"`,
					`"worker_name":"Test Target 2"`,
					`"Gracefully stopping worker","type":"serveJob","worker_name":"Test Target 2"`,
					`"worker_name":"Test Target 3"`,
					`"Gracefully stopping worker","type":"serveJob","worker_name":"Test Target 3"`,
					`"worker_name":"Test Target 4"`,
					`"Gracefully stopping worker","type":"serveJob","worker_name":"Test Target 4"`,
					`"worker_name":"Test Target 5"`,
					`"Gracefully stopping worker","type":"serveJob","worker_name":"Test Target 5"`,
					`"worker_name":"Test Target 6"`,
					`"Gracefully stopping worker","type":"serveJob","worker_name":"Test Target 6"`,
				},
				notWantLogOutputs: errorStrings,
			},
		},
		{
			name: "no servers",
			fields: fields{
				config: &entity.Config{
					NutServers: []*entity.NutServer{},
				},
			},
			attestations: attestations{
				wantLogOutputs:    []string{},
				notWantLogOutputs: errorStrings,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpTest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				unmarshalled := map[string]string{}
				err = json.Unmarshal(body, &unmarshalled)
				assert.NoError(t, err)

				mac, ok := unmarshalled["mac"]
				assert.True(t, ok)
				assert.NotEmpty(t, mac)

				http.Error(w, "Found", http.StatusOK)
			}))
			t.Cleanup(httpTest.Close)

			buf := &strings.Builder{}

			logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			workerPool, err := NewWorkerPool(ctx, tt.fields.config, tlsConfig, logger, httpTest.URL)
			require.NoError(t, err)
			workerPool.Start()

			time.Sleep(2 * time.Second)
			cancel()
			workerPool.Wait()

			for _, output := range tt.attestations.wantLogOutputs {
				assert.Contains(t, buf.String(), output)
			}

			for _, output := range tt.attestations.notWantLogOutputs {
				assert.NotContains(t, buf.String(), output)
			}
		})
	}
}
