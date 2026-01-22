package main

import (
	"log/slog"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServeCommand(t *testing.T) {
	logger := newTestLogger()

	emptyFs := afero.NewMemMapFs()

	got := NewServeCommand(t.Context(), logger, emptyFs, emptyFs)

	var gotFlagNames []string
	got.Flags().VisitAll(func(flag *pflag.Flag) {
		gotFlagNames = append(gotFlagNames, flag.Name)
	})

	wantFlagNames := []string{
		"host",
		"port",
		"ssl",
		"certFile",
		"keyFile",
		"config",
	}

	assert.Equal(t, "serve", got.Use)
	assert.NotEmpty(t, got.Short)
	assert.NotEmpty(t, got.Long)
	assert.NotEmpty(t, got.Example)

	assert.ElementsMatch(t, gotFlagNames, wantFlagNames)
}

func Test_serveCmdRunE(t *testing.T) {
	type args struct {
		cmdFunc func(_ *slog.Logger) *cobra.Command
		args    []string
	}
	tests := []struct {
		name           string
		args           args
		err            error
		wantOutputs    []string
		notWantOutputs []string
	}{
		{
			name: "empty config",
			args: args{
				cmdFunc: func(logger *slog.Logger) *cobra.Command {
					fs := afero.NewMemMapFs()
					err := afero.WriteFile(fs, "upswake.yaml", []byte(""), 0o644)
					require.NoError(t, err)

					regoFs := afero.NewMemMapFs()

					return NewServeCommand(t.Context(), logger, fs, regoFs)
				},
				args: []string{"serve", "--config", "upswake.yaml", "--port", "8081"},
			},
			err:            ErrTimeout, // expect a timeout error, as the command will run indefinitely otherwise
			wantOutputs:    []string{`"msg":"http(s) server started","address":"[::]:8081`},
			notWantOutputs: []string{`"level":"ERROR"`, `"level":"error"`, `"level":"Error"`},
		},
		{
			name: "valid config no rules",
			args: args{
				cmdFunc: func(logger *slog.Logger) *cobra.Command {
					fs := afero.NewMemMapFs()
					cfgYaml := `
nut_servers:
  - name: test-nut-server
    host: 127.0.0.1
    port: 3493
    username: username
    password: password
    targets:
      - name: test-target-server
        mac: "00:00:00:00:00:00"
        broadcast: 127.0.0.255
        port: 9
        interval: 50ms
        rules: {}
`
					err := afero.WriteFile(fs, "upswake.yml", []byte(cfgYaml), 0o644)
					require.NoError(t, err)

					regoFs := afero.NewMemMapFs()

					return NewServeCommand(t.Context(), logger, fs, regoFs)
				},
				args: []string{"serve", "--config", "upswake.yml", "--port", "8082"},
			},
			err: ErrTimeout, // expect a timeout error, as the command will run indefinitely otherwise
			wantOutputs: []string{
				`"msg":"http(s) server started","address":"[::]:8082"`,
				`"status":200`,
				`"level":"INFO"`,
				`"msg":"REQUEST","remote_ip":"127.0.0.1","host":"127.0.0.1:8082","method":"POST","uri":"/api/upswake","user_agent":"Go-http-client/1.1","status":200}`,
			},
			notWantOutputs: []string{
				`"level":"ERROR"`,
				`"level":"error"`,
				`"level":"Error"`,
				`"status":5`, // http 5xx errors
				`"status":4`, // http 4xx errors
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotOutput, err := executeCommandWithContext(t, test.args.cmdFunc, 1*time.Second, test.args.args...)

			assert.ErrorIs(t, err, test.err)

			for _, wantOutput := range test.wantOutputs {
				assert.Contains(t, gotOutput, wantOutput)
			}
			for _, notWantOutput := range test.notWantOutputs {
				assert.NotContains(t, gotOutput, notWantOutput)
			}
		})
	}
}
