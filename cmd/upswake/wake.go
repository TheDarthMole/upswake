package main

import (
	"github.com/TheDarthMole/UPSWake/internal/config"
	"github.com/TheDarthMole/UPSWake/internal/util"
	"github.com/TheDarthMole/UPSWake/internal/wol"
	"github.com/spf13/cobra"
	"log"
)

var (
	mac        string
	broadcasts []string
)

const WoLPort = 9

func init() {
	bc, err := util.GetAllBroadcastAddresses()
	if err != nil {
		panic(err)
	}
	stringBroadcasts := util.IPsToStrings(bc)
	wakeCmd.Flags().StringArrayVarP(&broadcasts, "broadcasts", "b", stringBroadcasts, "Broadcast addresses to send the WoL packets to, e.g. 192.168.1.255,172.16.0.255")
	wakeCmd.Flags().StringVarP(&mac, "mac", "m", "", "MAC address of the computer to wake")
	err = wakeCmd.MarkFlagRequired("mac")
	if err != nil {
		log.Panicf("not sure what happened here: %s", err)
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
			log.Fatal(err)
		}

		for _, broadcast := range ipBroadcasts {
			wolClient := wol.NewWoLClient(config.TargetServer{
				Name:      "CLI Request",
				Mac:       mac,
				Broadcast: broadcast.String(),
				Port:      WoLPort,
			})

			err := wolClient.Validate()
			if err != nil {
				log.Fatalf("failed to validate %s: %s", mac, err)
			}
			if err = wolClient.Wake(); err != nil {
				log.Fatalf("failed to wake %s: %s", mac, err)
			}
			log.Printf("Sent WoL packet to %s to wake %s", broadcast, mac)
		}

	},
}
