package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"upsWake/network"
	"upsWake/wol"
)

func init() {
	bc, err := network.GetAllBroadcastAddresses()
	if err != nil {
		panic(err)
	}
	stringBroadcasts := network.IPsToStrings(bc)
	wakeCmd.Flags().StringArrayVarP(&broadcasts, "broadcasts", "b", stringBroadcasts, "Broadcast addresses to send the WoL packets to, e.g. 192.168.1.255,172.16.0.255")
	wakeCmd.Flags().StringVarP(&mac, "mac", "m", "", "MAC address of the computer to wake")
	err = wakeCmd.MarkFlagRequired("mac")
	if err != nil {
		log.Println("baffled")
		return
	}
	rootCmd.AddCommand(wakeCmd)
}

var wakeCmd = &cobra.Command{
	Use:   "wake -b [mac address]",
	Short: "Run the UPSWake server",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		ipBroadcasts, err := network.StringsToIPs(broadcasts)
		if err != nil {
			panic(err)
		}
		err = wol.Wake(mac, ipBroadcasts)
		if err != nil {
			log.Panic(err)
		}
	},
}
