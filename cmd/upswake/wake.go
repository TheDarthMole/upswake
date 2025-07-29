package main

import (
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/util"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/spf13/cobra"
)

var (
	mac        string
	broadcasts []string
)

func init() {
	bc, err := util.GetAllBroadcastAddresses()
	if err != nil {
		sugar.Panic(err)
		return
	}
	stringBroadcasts := util.IPsToStrings(bc)
	wakeCmd.Flags().StringArrayVarP(&broadcasts, "broadcasts", "b", stringBroadcasts, "Broadcast addresses to send the WoL packets to")
	wakeCmd.Flags().StringVarP(&mac, "mac", "m", "", "MAC address of the computer to wake")
	err = wakeCmd.MarkFlagRequired("mac")
	if err != nil {
		sugar.Panicf("not sure what happened here: %s", err)
		return
	}
	rootCmd.AddCommand(wakeCmd)
}

var wakeCmd = &cobra.Command{
	Use:   "wake -b [mac address]",
	Short: "Manually wake a computer",
	Long:  `Manually wake a computer without using a UPS's status`,
	Run: func(cmd *cobra.Command, args []string) {
		ipBroadcasts, err := util.StringsToIPs(broadcasts)
		if err != nil {
			sugar.Fatal(err)
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
				sugar.Fatalf("failed to create new target server %s", err)
			}
			wolClient := wol.NewWoLClient(ts)

			if err = wolClient.Wake(); err != nil {
				sugar.Fatalf("failed to wake %s: %s", mac, err)
			}
			sugar.Infof("Sent WoL packet to %s to wake %s", broadcast, mac)
		}

	},
}
