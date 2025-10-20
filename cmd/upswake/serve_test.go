package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestNewServeCommand(t *testing.T) {
	logger := zaptest.NewLogger(t)
	testSugar := logger.Sugar()

	got := NewServeCommand(t.Context(), testSugar)

	var gotFlagNames []string
	got.Flags().VisitAll(func(flag *pflag.Flag) {
		gotFlagNames = append(gotFlagNames, flag.Name)
	})

	var wantFlagNames []string
	got.Flags().VisitAll(func(flag *pflag.Flag) {
		wantFlagNames = append(wantFlagNames, flag.Name)
	})

	assert.NotEmpty(t, got.Use)
	assert.NotEmpty(t, got.Short)
	assert.NotEmpty(t, got.Long)
	assert.NotEmpty(t, got.Example)

	assert.ElementsMatch(t, gotFlagNames, wantFlagNames)
}

func Test_serveCmdRunE(t *testing.T) {
	type args struct {
		cmdFunc func(_ *zap.SugaredLogger) *cobra.Command
		args    []string
		fs      func(t2 *testing.T) afero.Fs
		regoFS  func(t2 *testing.T) afero.Fs
	}
	tests := []struct {
		name           string
		args           args
		wantErr        assert.ErrorAssertionFunc
		err            string
		wantOutputs    []string
		notWantOutputs []string
	}{
		{
			name: "empty config",
			args: args{
				cmdFunc: func(logger *zap.SugaredLogger) *cobra.Command {
					return NewServeCommand(t.Context(), logger)
				},
				args: []string{"serve", "--config", "upswake.yml", "--port", "8081"},
				fs: func(t *testing.T) afero.Fs {
					fs := afero.NewMemMapFs()
					err := afero.WriteFile(fs, "upswake.yml", []byte(""), 0o644)
					require.NoError(t, err)
					return fs
				},
				regoFS: func(_ *testing.T) afero.Fs {
					fs := afero.NewMemMapFs()
					return fs
				},
			},
			wantErr:        assert.Error,
			err:            ErrTimeout.Error(), // expect a timeout error, as the command will run indefinitely otherwise
			wantOutputs:    []string{"http server started on [::]:8081"},
			notWantOutputs: []string{`"level":"ERROR"`, `"level":"error"`, `"level":"Error"`},
		},
		{
			name: "valid config no rules",
			args: args{
				cmdFunc: func(logger *zap.SugaredLogger) *cobra.Command {
					return NewServeCommand(t.Context(), logger)
				},
				args: []string{"serve", "--config", "upswake.yml", "--port", "8082"},
				fs: func(t *testing.T) afero.Fs {
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
					return fs
				},
				regoFS: func(_ *testing.T) afero.Fs {
					fs := afero.NewMemMapFs()
					return fs
				},
			},
			wantErr: assert.Error,
			err:     ErrTimeout.Error(), // expect a timeout error, as the command will run indefinitely otherwise
			wantOutputs: []string{
				"http server started on [::]:8082",
				`"status":200`,
				`{"level":"info","ts":`,
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileSystem = tt.args.fs(t)
			regoFiles = tt.args.regoFS(t)

			gotOutput, err := executeCommandWithContext(t, tt.args.cmdFunc, 1*time.Second, tt.args.args...)
			t.Log(gotOutput)

			tt.wantErr(t, err, fmt.Sprintf("serveCmdRunE(%v)", tt.args.args))

			if tt.err != "" {
				assert.ErrorContains(t, err, tt.err)
			}
			for _, wantOutput := range tt.wantOutputs {
				assert.Contains(t, gotOutput, wantOutput)
			}
			for _, notWantOutput := range tt.notWantOutputs {
				assert.NotContains(t, gotOutput, notWantOutput)
			}
		})
	}
}
