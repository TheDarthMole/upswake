package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/spf13/cobra"
)

var ErrNoBroadcasts = errors.New("no broadcast addresses provided; supply with --broadcasts or configure defaults")

type wakeCMD struct {
	logger *slog.Logger
}

func NewWakeCmd(logger *slog.Logger, broadcasts []net.IP) *cobra.Command {
	childLogger := logger.With(
		slog.String("cmd", "wake"),
	)

	wc := &wakeCMD{logger: childLogger}
	wakeCmd := &cobra.Command{
		Use:   "wake",
		Short: "Manually wake a computer",
		Long:  `Manually wake a computer without using a UPS's status`,
		Example: `  upswake wake -m 00:11:22:33:44:55
  upswake wake -m 00:11:22:33:44:55 -b 192.168.1.255,192.168.2.255`,
		RunE: wc.wakeCmdRunE,
	}

	wakeCmd.Flags().IPSliceP("broadcasts", "b", broadcasts, "Broadcast addresses to send the WoL packets to")
	wakeCmd.Flags().StringP("mac", "m", "", "(required) MAC address of the computer to wake")
	_ = wakeCmd.MarkFlagRequired("mac")

	return wakeCmd
}

func (wake *wakeCMD) wakeCmdRunE(cmd *cobra.Command, _ []string) error {
	mac, err := cmd.Flags().GetString("mac")
	if err != nil {
		return err
	}

	broadcasts, err := cmd.Flags().GetIPSlice("broadcasts")
	if err != nil {
		return err
	}

	if len(broadcasts) == 0 {
		return ErrNoBroadcasts
	}
	var joinedErr error
	for _, broadcast := range broadcasts {
		ts, err := entity.NewTargetServer(
			"CLI Request",
			mac,
			broadcast.String(),
			"1s",
			entity.DefaultWoLPort,
			[]string{},
		)
		if err != nil {
			wake.logger.Warn("Failed to create target server",
				slog.String("broadcast", broadcast.String()),
				slog.Any("error", err))
			joinedErr = errors.Join(joinedErr, fmt.Errorf("invalid target for %s: %w", broadcast, err))
			continue
		}
		wolClient := wol.NewWoLClient(ts)

		if err = wolClient.Wake(); err != nil {
			wake.logger.Warn("failed to send WoL packet",
				slog.String("broadcast", broadcast.String()),
				slog.String("mac", mac),
				slog.Any("error", err))
			joinedErr = errors.Join(joinedErr, fmt.Errorf("failed to wake %s via %s: %w", mac, broadcast, err))
			continue
		}
		wake.logger.Info("Sent WoL packet",
			slog.String("broadcast", broadcast.String()),
			slog.String("mac", mac))
	}
	return joinedErr
}
