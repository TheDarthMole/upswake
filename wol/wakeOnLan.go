package wol

import (
	"fmt"
	"github.com/sabhiram/go-wol/wol"
	"log"
	"net"
)

const MAGIC_PACKET_SIZE = 102

func Wake(mac string, broadcasts []net.IP, port int) error {
	if broadcasts == nil || len(broadcasts) == 0 {
		return fmt.Errorf("no broadcast addresses specified")
	}
	bs, err := newMagicPacket(mac)
	if err != nil {
		return err
	}

	for _, broadcast := range broadcasts {
		if broadcast == nil {
			return fmt.Errorf("the broadcast address cannot be nil")
		}
		err = sendWoLPacket(broadcast, port, bs)
		if err != nil {
			return fmt.Errorf("failed to send WoL packet: %w", err)
		}

		log.Printf("sent WoL packet to %s", broadcast)
	}
	return nil
}

func newMagicPacket(mac string) ([]byte, error) {
	mp, err := wol.New(mac)
	if err != nil {
		return nil, fmt.Errorf("failed to create magic packet: %w", err)
	}

	bs, err := mp.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal magic packet: %w", err)
	}
	return bs, nil
}

func sendWoLPacket(ip net.IP, port int, bs []byte) error {
	if net.ParseIP(ip.String()) == nil {
		return fmt.Errorf("the broadcast address cannot be nil or empty")
	}

	conn, err := net.DialUDP(
		"udp",
		&net.UDPAddr{IP: nil},
		&net.UDPAddr{IP: ip, Port: port})

	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}

	defer conn.Close()
	log.Printf("sending magic packet to %s:%d\n", ip, port)

	write, err := conn.Write(bs)
	if err == nil && write != MAGIC_PACKET_SIZE {
		err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", write)
	}

	if err != nil {
		return err
	}
	return nil
}
