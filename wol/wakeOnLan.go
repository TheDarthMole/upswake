package wol

import (
	wol "github.com/mdlayher/wol"
	"log"
	"net"
	"upsWake/network"
)

func Wake(mac string) error {
	interfaces, err := network.GetNetworkInterfaces()
	if err != nil {
		return err
	}
	for _, iface := range interfaces {
		hwMac, err := net.ParseMAC(mac)
		if err != nil {
			log.Fatalln("failed to parse MAC:", err)
		}
		client, err := wol.NewRawClient(&iface)
		if err != nil {
			log.Fatalln("failed to create raw client:", err)
		}
		err = client.Wake(hwMac)
		if err != nil {
			log.Fatalln("failed to send magic packet:", err)
		}
		err = client.Close()
		if err != nil {
			log.Fatalln("failed to close raw client:", err)
		}
	}
	return nil
}
