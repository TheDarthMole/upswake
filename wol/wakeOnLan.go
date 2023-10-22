package wol

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/sabhiram/go-wol/wol"
	"io"
	"net"
)

const MagicPacketSize = 102

type WoLTarget struct {
	Broadcast net.IP `validate:"required"`
	MAC       string `validate:"required,mac"`
	Port      int    `validate:"min=1,max=65535"`
}

func (tgt *WoLTarget) Wake() error {
	if err := validator.New().Struct(tgt); err != nil {
		return err
	}

	conn, err := net.DialUDP("udp",
		nil,
		&net.UDPAddr{
			IP:   tgt.Broadcast,
			Port: tgt.Port,
		})
	if err != nil {
		return err
	}
	defer conn.Close()
	return wakeInternal(conn, tgt.MAC)
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
