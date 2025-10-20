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
				wake := wake{logger: testSugar}
				wakeCmd := &cobra.Command{
					Use:   "wake -b [mac address]",
					Short: "Manually wake a computer",
					Long:  "Manually wake a computer without using a UPS's status",
					RunE:  wake.wakeCmdRunE,
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
				wake := wake{logger: testSugar}
				wakeCmd := &cobra.Command{
					Use:   "wake -b [mac address]",
					Short: "Manually wake a computer",
					Long:  "Manually wake a computer without using a UPS's status",
					RunE:  wake.wakeCmdRunE,
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
				wake := wake{logger: testSugar}
				wakeCmd := &cobra.Command{
					Use:   "wake -b [mac address]",
					Short: "Manually wake a computer",
					Long:  "Manually wake a computer without using a UPS's status",
					RunE:  wake.wakeCmdRunE,
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
			got.Flags().VisitAll(func(flag *pflag.Flag) {
				wantFlagNames = append(wantFlagNames, flag.Name)
			})

			wantBroadcasts, err := want.Flags().GetIPSlice("broadcasts")
			assert.NoError(t, err)
			gotBroadcasts, err := got.Flags().GetIPSlice("broadcasts")
			assert.NoError(t, err)

			assert.Equal(t, want.Use, got.Use)
			assert.Equal(t, want.Short, got.Short)
			assert.Equal(t, want.Long, got.Long)
			assert.Equal(t, wantBroadcasts, gotBroadcasts)
			assert.ElementsMatch(t, gotFlagNames, wantFlagNames)
		})
	}
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executeCommandWithContext(t, tt.args.cmdFunc, 1*time.Second, tt.args.args...)
			fmt.Println(output)

			tt.wantErr(t, err, fmt.Sprintf("wakeCmdRunE(%v)", tt.args.args))
			assert.Contains(t, output, tt.output)
		})
	}
}
