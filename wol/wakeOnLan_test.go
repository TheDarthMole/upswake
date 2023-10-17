package wol

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"testing"
	"time"
)

type udpResponse struct {
	bs       []byte
	received chan int
}

func TestWake(t *testing.T) {
	type args struct {
		mac            string
		broadcasts     []net.IP
		expectedPacket []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		conns   []*net.UDPConn
		port    int
	}{
		{
			name: "invalid MAC",
			args: args{
				mac: "invalid",
				broadcasts: []net.IP{
					net.IPv4(127, 0, 0, 1)}},
			wantErr: true,
			conns:   []*net.UDPConn{},
			port:    0,
		},
		{
			name: "invalid broadcast",
			args: args{
				mac:        "00:00:00:00:00:00",
				broadcasts: []net.IP{},
			},
			wantErr: true,
			conns:   []*net.UDPConn{},
			port:    0,
		},
		{
			name: "valid",
			args: args{
				mac: "01:02:03:04:05:06",
				broadcasts: []net.IP{
					net.IPv4(127, 0, 0, 1),
				},
				expectedPacket: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			},
			wantErr: false,
			conns: []*net.UDPConn{
				createUDPListener("127.0.0.1", 6422),
			},
			port: 6422,
		},
		{
			name: "invalid udp address",
			args: args{
				mac: "01:02:03:04:05:06",
				broadcasts: []net.IP{
					nil,
				},
			},
			wantErr: true,
			conns:   []*net.UDPConn{},
			port:    0,
		},
	}
	responseMap := make(map[int][]udpResponse)

	for testNumber, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseMap[testNumber] = make([]udpResponse, len(tt.args.broadcasts))

			for i, broadcast := range tt.args.broadcasts {
				// we can't create a listener on a nil broadcast, default to localhost
				if broadcast == nil {
					broadcast = net.IPv4(127, 0, 0, 1)
				}
				responseMap[testNumber][i] = udpResponse{
					bs:       make([]byte, MAGIC_PACKET_SIZE),
					received: make(chan int),
				}

				if len(tt.conns) != 0 {
					defer tt.conns[i].Close()
					go listenOnConn(tt.conns[i], &responseMap[testNumber][i])
					time.Sleep(1 * time.Second)
				}
			}

			if err := Wake(tt.args.mac, tt.args.broadcasts, tt.port); (err != nil) != tt.wantErr {
				t.Errorf("Wake() error = %v, wantErr %v", err, tt.wantErr)
			}

			for _, response := range responseMap[testNumber] {

				// if we don't expect a packet, don't wait for one
				select {
				case <-response.received:
					if !reflect.DeepEqual(response.bs, tt.args.expectedPacket) {
						t.Errorf("Wake() got = %v, want %v", response.bs, tt.args.expectedPacket)
					}
				case <-time.After(25 * time.Millisecond):
					if !tt.wantErr {
						t.Errorf("failed to recieve UDP packet on %s", tt.conns[0].LocalAddr().String())
					}
				}
			}

		})
	}
}

func Test_newMagicPacket(t *testing.T) {
	type args struct {
		mac string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "invalid MAC",
			args: args{
				mac: "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid MAC too long",
			args: args{
				mac: "01:02:03:04:05:06:07",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid MAC too short",
			args: args{
				mac: "01:02:03:04:05",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid MAC wrong format",
			args: args{
				mac: "01:02:03:04:gg",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid MAC",
			args: args{
				mac: "01:02:03:04:05:06",
			},
			want: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newMagicPacket(tt.args.mac)
			if (err != nil) != tt.wantErr {
				t.Errorf("newMagicPacket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMagicPacket() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sendWoLPacket(t *testing.T) {
	wolSizedPacket := make([]byte, MAGIC_PACKET_SIZE)
	type args struct {
		ip net.IP
		bs []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil IP",
			args: args{
				ip: nil,
				bs: wolSizedPacket,
			},
			wantErr: true,
		},
		{
			name: "empty IP",
			args: args{
				ip: net.IP{},
				bs: wolSizedPacket,
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				ip: net.IPv4(127, 0, 0, 255),
				bs: wolSizedPacket,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendWoLPacket(tt.args.ip, 9, tt.args.bs); (err != nil) != tt.wantErr {
				t.Errorf("sendWoLPacket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func createUDPListener(address string, port int) *net.UDPConn {
	udpaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Fatalf("failed to resolve UDP address: %v", err)
	}
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		log.Fatalf("failed to listen on UDP: %v", err)
	}
	return conn
}

func listenOnConn(conn *net.UDPConn, response *udpResponse) {
	_, _, err := conn.ReadFromUDP(response.bs)
	if err != nil {
		log.Fatalf("failed to read UDP: %v", err)
	}
	// signal that we received a packet
	response.received <- 1
}
