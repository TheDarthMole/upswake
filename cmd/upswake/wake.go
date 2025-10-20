package main

import (
	"fmt"
	"net"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type wakeCMD struct {
	logger *zap.SugaredLogger
}

func NewWakeCmd(logger *zap.SugaredLogger, broadcasts []net.IP) *cobra.Command {
	wake := &wakeCMD{logger: logger}
	wakeCmd := &cobra.Command{
		Use:   "wake -b [mac address]",
		Short: "Manually wake a computer",
		Long:  `Manually wake a computer without using a UPS's status`,
		RunE:  wake.wakeCmdRunE,
	}

	wakeCmd.Flags().IPSliceP("broadcasts", "b", broadcasts, "Broadcast addresses to send the WoL packets to")
	wakeCmd.Flags().StringP("mac", "m", "", "MAC address of the computer to wake")
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
			return err
		}
		wolClient := wol.NewWoLClient(ts)

		if err = wolClient.Wake(); err != nil {
			return fmt.Errorf("failed to wake %s: %w", mac, err)
		}
		wake.logger.Infof("Sent WoL packet to %s to wake %s", broadcast, mac)
	}
	return nil
}
