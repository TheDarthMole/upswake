package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var errorNoBroadcasts = fmt.Errorf("no broadcast addresses provided; supply with --broadcasts or configure defaults")

type wakeCMD struct {
	logger *zap.SugaredLogger
}

func NewWakeCmd(logger *zap.SugaredLogger, broadcasts []net.IP) *cobra.Command {
	wc := &wakeCMD{logger: logger}
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
		return errorNoBroadcasts
	}

	var joined error
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
			joined = errors.Join(joined, fmt.Errorf("invalid target for %s: %w", broadcast, err))
			continue
		}
		wolClient := wol.NewWoLClient(ts)

		if err = wolClient.Wake(); err != nil {
			joined = errors.Join(joined, fmt.Errorf("failed to wake %s via %s: %w", mac, broadcast, err))
			continue
		}
		wake.logger.Infof("Sent WoL packet to %s to wake %s", broadcast, mac)
	}
	return joined
}
