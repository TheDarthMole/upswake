package wol

import (
	"net"
	"testing"
)

func TestWake(t *testing.T) {
	type args struct {
		mac        string
		broadcasts []net.IP
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid MAC",
			args: args{
				mac: "invalid",
				broadcasts: []net.IP{
					net.IPv4(192, 168, 1, 255),
					net.IPv4(10, 0, 0, 255)}},
			wantErr: true,
		},
		{
			name: "invalid broadcast",
			args: args{
				mac:        "00:00:00:00:00:00",
				broadcasts: []net.IP{},
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				mac: "01:02:03:04:05:06",
				broadcasts: []net.IP{
					net.IPv4(127, 0, 0, 255),
				},
			},
			wantErr: false,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Wake(tt.args.mac, tt.args.broadcasts); (err != nil) != tt.wantErr {
				t.Errorf("Wake() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
