package wol

import (
	"fmt"
	"github.com/sabhiram/go-wol/wol"
	"io"
	"net"
)

const MagicPacketSize = 102

func Wake(dst *net.UDPConn, mac string) error {
	return wakeInternal(dst, mac)
}

func wakeInternal(dst io.ReadWriteCloser, mac string) error {
	mp, err := newMagicPacket(mac)
	if err != nil {
		return err
	}

	size, err := dst.Write(mp)
	if err == nil && size != MagicPacketSize {
		err = fmt.Errorf("magic packet sent was %d bytes (expected %d bytes sent)", size, MagicPacketSize)
	}

	if err != nil {
		return fmt.Errorf("failed to send WoL packet: %w", err)
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
