package main

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestNewWakeCmd(t *testing.T) {
	logger := zaptest.NewLogger(t)
	testSugar := logger.Sugar()

	type args struct {
		broadcasts []net.IP
	}
	tests := []struct {
		name string
		args args
		want func() *cobra.Command
	}{
		{
			name: "empty broadcasts",
			args: args{broadcasts: []net.IP{}},
			want: func() *cobra.Command {
				wake := wakeCMD{logger: testSugar}
				wakeCmd := &cobra.Command{
					RunE: wake.wakeCmdRunE,
				}
				wakeCmd.Flags().IPSliceP("broadcasts", "b", []net.IP{}, "Broadcast addresses to send the WoL packets to")
				wakeCmd.Flags().StringP("mac", "m", "", "MAC address of the computer to wake")
				_ = wakeCmd.MarkFlagRequired("mac")

				return wakeCmd
			},
		},
		{
			name: "one broadcasts",
			args: args{broadcasts: []net.IP{{127, 0, 0, 255}}},
			want: func() *cobra.Command {
				wake := wakeCMD{logger: testSugar}
				wakeCmd := &cobra.Command{
					RunE: wake.wakeCmdRunE,
				}
				wakeCmd.Flags().IPSliceP("broadcasts", "b", []net.IP{{127, 0, 0, 255}}, "Broadcast addresses to send the WoL packets to")
				wakeCmd.Flags().StringP("mac", "m", "", "MAC address of the computer to wake")
				_ = wakeCmd.MarkFlagRequired("mac")

				return wakeCmd
			},
		},
		{
			name: "multiple broadcasts",
			args: args{broadcasts: []net.IP{{127, 0, 0, 255}, {192, 168, 1, 255}, {10, 0, 0, 255}}},
			want: func() *cobra.Command {
				wake := wakeCMD{logger: testSugar}
				wakeCmd := &cobra.Command{
					RunE: wake.wakeCmdRunE,
				}
				wakeCmd.Flags().IPSliceP("broadcasts", "b", []net.IP{{127, 0, 0, 255}, {192, 168, 1, 255}, {10, 0, 0, 255}}, "Broadcast addresses to send the WoL packets to")
				wakeCmd.Flags().StringP("mac", "m", "", "MAC address of the computer to wake")
				_ = wakeCmd.MarkFlagRequired("mac")

				return wakeCmd
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.want()
			got := NewWakeCmd(testSugar, tt.args.broadcasts)

			var gotFlagNames []string
			got.Flags().VisitAll(func(flag *pflag.Flag) {
				gotFlagNames = append(gotFlagNames, flag.Name)
			})

			var wantFlagNames []string
			want.Flags().VisitAll(func(flag *pflag.Flag) {
				wantFlagNames = append(wantFlagNames, flag.Name)
			})

			wantBroadcasts, err := want.Flags().GetIPSlice("broadcasts")
			assert.NoError(t, err)
			gotBroadcasts, err := got.Flags().GetIPSlice("broadcasts")
			assert.NoError(t, err)

			assert.Equal(t, wantBroadcasts, gotBroadcasts)
			assert.ElementsMatch(t, gotFlagNames, wantFlagNames)
		})
	}

	t.Run("viper config", func(t *testing.T) {
		broadcasts := []net.IP{{192, 168, 1, 255}}
		wakeCmd := NewWakeCmd(testSugar, broadcasts)

		assert.Equal(t, "wake", wakeCmd.Use)
		assert.NotEmpty(t, wakeCmd.Short)
		assert.NotEmpty(t, wakeCmd.Long)
		assert.NotEmpty(t, wakeCmd.Example)
	})
}

func Test_wakeCmdRunE(t *testing.T) {
	type args struct {
		cmdFunc func(_ *zap.SugaredLogger) *cobra.Command
		args    []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
		output  string
	}{
		{
			name: "valid",
			args: args{
				cmdFunc: func(logger *zap.SugaredLogger) *cobra.Command {
					return NewWakeCmd(logger, []net.IP{{127, 0, 0, 255}})
				},
				args: []string{"wake", "--mac", "00:00:00:00:00:00"},
			},
			wantErr: assert.NoError,
			output:  `Sent WoL packet to 127.0.0.255 to wake 00:00:00:00:00:00`,
		},
		{
			name: "no broadcasts",
			args: args{
				cmdFunc: func(logger *zap.SugaredLogger) *cobra.Command {
					return NewWakeCmd(logger, []net.IP{})
				},
				args: []string{"wake", "--mac", "00:00:00:00:00:00"},
			},
			wantErr: assert.Error,
			output:  errorNoBroadcasts.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executeCommandWithContext(t, tt.args.cmdFunc, 1*time.Second, tt.args.args...)
			t.Log(output)

			tt.wantErr(t, err, fmt.Sprintf("wakeCmdRunE(%v)", tt.args.args))
			assert.Contains(t, output, tt.output)
		})
	}
}
