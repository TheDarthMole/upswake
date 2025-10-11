package main

import (
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/network"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/spf13/cobra"
)

var (
	mac        string
	broadcasts []string
)

func init() {

}

func NewWakeCmd(broadcasts []string) *cobra.Command {
	wakeCmd := &cobra.Command{
		Use:   "wake -b [mac address]",
		Short: "Manually wake a computer",
		Long:  `Manually wake a computer without using a UPS's status`,
		RunE:  wakeCmdRunE,
	}

	wakeCmd.Flags().StringArrayVarP(&broadcasts, "broadcasts", "b", broadcasts, "Broadcast addresses to send the WoL packets to")
	wakeCmd.Flags().StringVarP(&mac, "mac", "m", "", "MAC address of the computer to wake")
	_ = wakeCmd.MarkFlagRequired("mac")

	return wakeCmd
}

func wakeCmdRunE(_ *cobra.Command, _ []string) error {
	ipBroadcasts, err := network.StringsToIPs(broadcasts)
	if err != nil {
		return err
	}

	for _, broadcast := range ipBroadcasts {
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
			return fmt.Errorf("failed to wake %s: %s", mac, err)
		}
		sugar.Infof("Sent WoL packet to %s to wake %s", broadcast, mac)
	}
	return nil
}
