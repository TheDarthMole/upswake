package wol

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/sabhiram/go-wol/wol"
	"io"
	"log"
	"net"
)

const MagicPacketSize = 102

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

	return mp.Marshal()
}
