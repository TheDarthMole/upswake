package main

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewRootCommand(t *testing.T) {
	t.Run("root command", func(t *testing.T) {
		testRootCmd := NewRootCommand()
		assert.Equal(t, "upswake", testRootCmd.Use, "root command should be 'upswake'")
		assert.Equal(t, "UPSWake sends Wake on LAN packets based on a UPS's status", testRootCmd.Short, "root command short description mismatch")
		assert.Contains(t, testRootCmd.Long, "UPSWake sends Wake on LAN packets to target servers", "root command long description mismatch")
	})
}

func Test_Execute(t *testing.T) {
	type args struct {
		args       []string
		filesystem func() afero.Fs
		regoFiles  func() afero.Fs
	}
	tests := []struct {
		name          string
		args          args
		exitCode      int
		timeout       time.Duration
		wantOutput    []string
		notWantOutput []string
	}{
		{
			name: "root command help",
			args: args{
				args:       []string{"upswake", "--help"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		// {
		//	name: "root command version",
		//	args: args{
		//		args: []string{"upswake", "--version"},
		//	},
		//	exitCode: 0,
		// },
		{
			name: "json command help",
			args: args{
				args:       []string{"upswake", "json", "--help"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "json command with long args",
			args: args{
				args:       []string{"upswake", "json", "--host", "127.0.0.1", "--port", "3493", "--username", "admin", "--password", "password"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "json command with short args",
			args: args{
				args:       []string{"upswake", "json", "-H", "127.0.0.1", "-P", "3493", "-u", "admin", "-p", "password"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "wake command help",
			args: args{
				args:       []string{"upswake", "wake", "--help"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "wake command with mac short flag",
			args: args{
				args:       []string{"upswake", "wake", "-m", "00:00:00:00:00:00"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "wake command with mac and broadcast short flags",
			args: args{
				args:       []string{"upswake", "wake", "-m", "00:00:00:00:00:00", "-b", "127.0.0.255"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "wake command with mac long flag",
			args: args{
				args:       []string{"upswake", "wake", "--mac", "00:00:00:00:00:00"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "wake command with mac and broadcast long flags",
			args: args{
				args:       []string{"upswake", "wake", "--mac", "00:00:00:00:00:00", "--broadcasts", "127.0.0.255"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "wake command with mac and multiple broadcasts long flags",
			args: args{
				args:       []string{"upswake", "wake", "--mac", "00:00:00:00:00:00", "--broadcasts", "127.0.0.255,127.0.0.1"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "serve command help",
			args: args{
				args:       []string{"upswake", "serve", "--help"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   0,
			timeout:    5 * time.Second,
			wantOutput: []string{},
		},
		{
			name: "serve command no arguments no config",
			args: args{
				args:       []string{"upswake", "serve"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode: 1,
			timeout:  5 * time.Second,
			wantOutput: []string{
				"Error: error loading config: open config.yaml: file does not exist",
			},
			notWantOutput: []string{},
		},
		{
			name: "serve command no arguments basic config",
			args: args{
				args: []string{"upswake", "serve"},
				filesystem: func() afero.Fs {
					fs := afero.NewMemMapFs()
					const configYaml = `
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
        interval: 300ms
        rules: {}`
					err := afero.WriteFile(fs, "config.yaml", []byte(configYaml), 0o644)
					require.NoError(t, err)
					return fs
				},
				regoFiles: afero.NewMemMapFs,
			},
			exitCode: 0,
			timeout:  5 * time.Second,
			wantOutput: []string{
				"http server started on",
				"Gracefully stopping worker",
			},
			notWantOutput: []string{"ERROR", "error"},
		},
		{
			name: "non-existent command help",
			args: args{
				args:       []string{"upswake", "non-existent", "--help"},
				filesystem: afero.NewMemMapFs,
				regoFiles:  afero.NewMemMapFs,
			},
			exitCode:   1,
			timeout:    5 * time.Second,
			wantOutput: []string{`Error: unknown command "non-existent" for "upswake"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutput := getStdoutStderr(t, func() {
				c := make(chan int, 1)

				ctx, cancel := context.WithCancel(context.Background())

				var wg sync.WaitGroup
				wg.Add(1)

				origArgs := os.Args
				defer func() { os.Args = origArgs }()

				go func() {
					defer wg.Done()
					os.Args = tt.args.args

					c <- Execute(ctx, tt.args.filesystem(), tt.args.regoFiles())
				}()

				select {
				case exitCode := <-c:
					// use err and reply
					assert.Equal(t, tt.exitCode, exitCode)
					cancel()
				case <-time.After(tt.timeout):
					cancel()
					wg.Wait()
					// set the error to be a timeout error
				}
			})

			for _, output := range tt.wantOutput {
				assert.Contains(t, gotOutput, output)
			}
			for _, output := range tt.notWantOutput {
				assert.NotContains(t, gotOutput, output)
			}
		})
	}
}
