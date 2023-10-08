package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"upsWake/network"
	"upsWake/wol"
)

var (
	mac        string
	broadcasts []string
)

func init() {
	bc, err := network.GetAllBroadcastAddresses()
	if err != nil {
		panic(err)
	}
	log.Println(bc)
	log.Println("hehe")
	stringBroadcasts := network.IPsToStrings(bc)
	serveCmd.Flags().StringArrayVarP(&broadcasts, "broadcasts", "b", stringBroadcasts, "Broadcast addresses to send the WoL packet to")
	serveCmd.Flags().StringVarP(&mac, "mac", "m", "", "MAC address of the computer to wake")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the UPSWake server",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		ipBroadcasts, err := network.StringsToIPs(broadcasts)
		log.Println(ipBroadcasts)
		if err != nil {
			panic(err)
		}
		err = wol.Wake(mac, ipBroadcasts)
		if err != nil {
			log.Panic(err)
		}
	},
}
