package wol

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/sabhiram/go-wol/wol"
)

const MagicPacketSize = 102

var (
	ErrFailedCreateMagicPacket = errors.New("failed to create magic packet")
	ErrFailedSendWoLPacket     = errors.New("failed to send WoL packet")
	ErrExpectedPacketSize      = errors.New(fmt.Sprintf("magic packet sent was expected to be of size %d", MagicPacketSize))
)

type WakeOnLan struct {
	*entity.TargetServer
}

func NewWoLClient(target *entity.TargetServer) *WakeOnLan {
	return &WakeOnLan{
		target,
	}
}

func (tgt *WakeOnLan) Wake() error {
	conn, err := net.DialUDP("udp",
		nil,
		&net.UDPAddr{
			IP:   net.ParseIP(tgt.Broadcast),
			Port: tgt.Port,
		})
	if err != nil {
		return err
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to close UDP connection: %s", err)
		}
	}(conn)
	return wakeInternal(conn, tgt.MAC)
}

func wakeInternal(dst io.ReadWriteCloser, mac string) error {
	mp, err := newMagicPacket(mac)
	if err != nil {
		return err
	}

	size, err := dst.Write(mp)

	if err != nil {
		return errors.Join(ErrFailedSendWoLPacket, err)
	}

	if size != MagicPacketSize {
		return errors.Join(ErrExpectedPacketSize, fmt.Errorf("magic packet sent was %d", size))
	}

	return nil
}

func newMagicPacket(mac string) ([]byte, error) {
	mp, err := wol.New(mac)
	if err != nil {
		return nil, errors.Join(ErrFailedCreateMagicPacket, err)
	}

	return mp.Marshal()
}
