package wol

import (
	"fmt"
	"github.com/sabhiram/go-wol/wol"
	"log"
	"net"
)

func Wake(mac string, broadcasts []net.IP) error {
	if _, err := net.ParseMAC(mac); err != nil {
		return fmt.Errorf("invalid MAC address: %s", mac)
	}

	mp, err := wol.New(mac)
	if err != nil {
		return fmt.Errorf("failed to create magic packet: %w", err)
	}

	bs, err := mp.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal magic packet: %w", err)
	}

	for _, broadcast := range broadcasts {

		conn, err := net.DialUDP(
			"udp",
			&net.UDPAddr{IP: nil},
			&net.UDPAddr{IP: broadcast, Port: 9})

		if err != nil {
			log.Printf("failed to dial UDP: %s", err)
			continue
		}
		defer conn.Close()
		write, err := conn.Write(bs)
		if err == nil && write != 102 {
			err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", write)
		}
		if err != nil {
			log.Printf("failed to write UDP: %s", err)
			continue
		}
		log.Printf("sent WoL packet to %s", broadcast)
	}
	return nil
}
