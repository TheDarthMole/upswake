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
			log.Fatalf("failed to create raw client on %s: %q\n", iface.Name, err)
		}
		err = client.Wake(hwMac)
		if err != nil {
			log.Fatalf("failed to send magic packet on %s: %q\n", iface.Name, err)
		}
		err = client.Close()
		if err != nil {
			log.Fatalf("failed to close raw client on %s: %q", iface.Name, err)
		}
	}
	return nil
}
