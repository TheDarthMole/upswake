package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheDarthMole/UPSWake/internal/api/handlers"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewHealthCheckCommand(t *testing.T) {
	logger := zap.NewNop().Sugar()
	got := NewHealthCheckCommand(logger)

	assert.NotNil(t, got)
	assert.Equal(t, "healthcheck", got.Use)
	assert.NotEmpty(t, got.Short)
	assert.NotEmpty(t, got.Long)

	assert.Equal(t, "localhost", got.Flags().Lookup("host").DefValue, "default host should be 'localhost'")
	assert.Equal(t, "8080", got.Flags().Lookup("port").DefValue, "default port should be '8080'")
	assert.Equal(t, "false", got.Flags().Lookup("ssl").DefValue, "default ssl should be 'false'")
}

func Test_healthCheck_HealthCheckRunE(t *testing.T) {
	type fields struct {
		logger      *zap.SugaredLogger
		handlerFunc func(http.ResponseWriter, *http.Request)
	}
	type args struct {
		cmd *cobra.Command
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		closeServer bool
		err         error
	}{
		{
			name: "successful healthcheck",
			fields: fields{
				logger: zap.NewNop().Sugar(),
				handlerFunc: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					apiResponse := handlers.Response{
						Message: "OK",
					}
					response, err := json.Marshal(apiResponse)
					assert.NoError(t, err)
					_, err = w.Write(response)
					assert.NoError(t, err)
				},
			},
			args: args{
				cmd: NewHealthCheckCommand(zap.NewNop().Sugar()),
			},
			err: nil,
		},
		{
			name: "internal error healthcheck",
			fields: fields{
				logger: zap.NewNop().Sugar(),
				handlerFunc: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					apiResponse := handlers.Response{
						Message: "Not OK",
					}
					response, err := json.Marshal(apiResponse)
					assert.NoError(t, err)
					_, err = w.Write(response)
					assert.NoError(t, err)
				},
			},
			args: args{
				cmd: NewHealthCheckCommand(zap.NewNop().Sugar()),
			},
			err: ErrHealthCheckFailed,
		},
		{
			name: "no response healthcheck",
			fields: fields{
				logger:      zap.NewNop().Sugar(),
				handlerFunc: func(_ http.ResponseWriter, _ *http.Request) {},
			},
			args: args{
				cmd: NewHealthCheckCommand(zap.NewNop().Sugar()),
			},
			closeServer: true,
			err:         ErrMakingRequest,
		},
	}
	for _, tt := range tests {
		t.Run("no ssl "+tt.name, func(t *testing.T) {
			h := &healthCheck{
				logger: tt.fields.logger,
			}

			mockServer := httptest.NewServer(http.HandlerFunc(tt.fields.handlerFunc))
			defer mockServer.Close()

			if tt.closeServer {
				mockServer.Close()
			}

			host, port, err := net.SplitHostPort(mockServer.Listener.Addr().String())
			require.NoError(t, err)

			cliArgs := []string{"upswake", "serve", "healthcheck", "--host", host, "--port", port}
			require.NoError(t, tt.args.cmd.ParseFlags(cliArgs))

			got := h.HealthCheckRunE(tt.args.cmd, cliArgs)

			assert.ErrorIs(t, got, tt.err)
		})

		t.Run("ssl "+tt.name, func(t *testing.T) {
			h := &healthCheck{
				logger: tt.fields.logger,
			}

			mockServer := httptest.NewTLSServer(http.HandlerFunc(tt.fields.handlerFunc))
			defer mockServer.Close()
			mockServer.Config.TLSConfig.InsecureSkipVerify = true // TODO: could probably pass in some trusted certificates instead of this

			if tt.closeServer {
				mockServer.Close()
			}

			host, port, err := net.SplitHostPort(mockServer.Listener.Addr().String())
			require.NoError(t, err)

			cliArgs := []string{"upswake", "serve", "healthcheck", "--host", host, "--port", port, "--ssl"}
			require.NoError(t, tt.args.cmd.ParseFlags(cliArgs))

			got := h.HealthCheckRunE(tt.args.cmd, cliArgs)

			assert.ErrorIs(t, got, tt.err)
		})
	}
}
