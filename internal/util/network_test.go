package util

import (
	"net"
	"reflect"
	"testing"
)

func Test_getIPBroadcast(t *testing.T) {
	type args struct {
		addr net.Addr
	}
	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{
			name: "IPv6 Address",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv6loopback,
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: nil,
		},
		{
			name: "Test 1",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 168, 1, 1),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: net.IPv4(192, 168, 1, 255),
		},

		{
			name: "Test 2",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(10, 0, 0, 1),
					Mask: net.IPv4Mask(255, 0, 0, 0),
				},
			},
			want: net.IPv4(10, 255, 255, 255),
		},

		{
			name: "Test 3",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(172, 16, 0, 1),
					Mask: net.IPv4Mask(255, 255, 0, 0),
				},
			},
			want: net.IPv4(172, 16, 255, 255),
		},

		{
			name: "Test 4",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 168, 0, 1),
					Mask: net.IPv4Mask(255, 255, 255, 128),
				},
			},
			want: net.IPv4(192, 168, 0, 127),
		},

		{
			name: "Test 5",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(10, 1, 2, 3),
					Mask: net.IPv4Mask(255, 255, 255, 255),
				},
			},
			want: net.IPv4(10, 1, 2, 3),
		},
		{
			name: "Test 6",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(23, 4, 6, 8),
					Mask: net.IPv4Mask(255, 240, 0, 0),
				},
			},
			want: net.IPv4(23, 15, 255, 255),
		},
		{
			name: "Test 7",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(172, 31, 15, 1),
					Mask: net.IPv4Mask(255, 255, 255, 240),
				},
			},
			want: net.IPv4(172, 31, 15, 15),
		},

		{
			name: "Test 8",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 168, 5, 10),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: net.IPv4(192, 168, 5, 255),
		},

		{
			name: "Test 9",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(10, 0, 0, 5),
					Mask: net.IPv4Mask(255, 255, 255, 128),
				},
			},
			want: net.IPv4(10, 0, 0, 127),
		},

		{
			name: "Test 10",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 168, 2, 1),
					Mask: net.IPv4Mask(255, 255, 255, 192),
				},
			},
			want: net.IPv4(192, 168, 2, 63),
		},

		{
			name: "Test 11",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(172, 20, 10, 5),
					Mask: net.IPv4Mask(255, 255, 255, 254),
				},
			},
			want: net.IPv4(172, 20, 10, 5),
		},

		{
			name: "Test 12",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 168, 3, 1),
					Mask: net.IPv4Mask(255, 255, 255, 240),
				},
			},
			want: net.IPv4(192, 168, 3, 15),
		},

		{
			name: "Test 13",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(10, 2, 3, 4),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: net.IPv4(10, 2, 3, 255),
		},

		{
			name: "Test 14",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(172, 16, 8, 4),
					Mask: net.IPv4Mask(255, 255, 255, 128),
				},
			},
			want: net.IPv4(172, 16, 8, 127),
		},

		{
			name: "Test 15",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 168, 10, 20),
					Mask: net.IPv4Mask(255, 255, 0, 0),
				},
			},
			want: net.IPv4(192, 168, 255, 255),
		},
		{
			name: "Test 16",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 168, 1, 1),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: net.IPv4(192, 168, 1, 255),
		},
		{
			name: "Test 17",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 168, 1, 5),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: net.IPv4(192, 168, 1, 255),
		},
		{
			name: "Test 18",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(192, 5, 1, 1),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: net.IPv4(192, 5, 1, 255),
		},
		{
			name: "Test 19",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(127, 0, 0, 50),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: net.IPv4(127, 0, 0, 255),
		},
		{
			name: "Test 20",
			args: args{
				addr: &net.IPNet{
					IP:   net.IPv4(52, 11, 43, 5),
					Mask: net.IPv4Mask(255, 254, 0, 0),
				},
			},
			want: net.IPv4(52, 11, 255, 255),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIPBroadcast(tt.args.addr); !got.Equal(tt.want) {
				t.Errorf("getIPBroadcast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPsToStrings(t *testing.T) {
	type args struct {
		input []net.IP
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test 1",
			args: args{
				input: []net.IP{
					net.IPv4(192, 168, 1, 1),
					net.IPv4(192, 168, 1, 2),
					net.IPv4(192, 168, 1, 3),
					net.IPv4(192, 168, 1, 4),
				},
			},
			want: []string{
				"192.168.1.1",
				"192.168.1.2",
				"192.168.1.3",
				"192.168.1.4",
			},
		},
		{
			name: "Test 2",
			args: args{
				input: []net.IP{},
			},
			want: []string{},
		},
		{
			name: "Test 3",
			args: args{
				input: []net.IP{
					net.IPv4(203, 42, 12, 56),
					net.IPv4(172, 31, 254, 1),
					net.IPv4(10, 20, 30, 40),
					net.IPv4(128, 0, 0, 1),
					net.IPv4(255, 255, 255, 255),
					net.IPv4(145, 78, 90, 200),
					net.IPv4(192, 168, 0, 100),
					net.IPv4(173, 194, 55, 22),
					net.IPv4(44, 67, 89, 123),
					net.IPv4(58, 104, 76, 92),
				},
			},
			want: []string{
				"203.42.12.56",
				"172.31.254.1",
				"10.20.30.40",
				"128.0.0.1",
				"255.255.255.255",
				"145.78.90.200",
				"192.168.0.100",
				"173.194.55.22",
				"44.67.89.123",
				"58.104.76.92",
			},
		},
		{
			name: "Test 4",
			args: args{
				input: []net.IP{
					net.IPv6loopback,
				},
			},
			want: []string{
				"::1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPsToStrings(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IPsToStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringsToIPs(t *testing.T) {
	type args struct {
		ips []string
	}
	tests := []struct {
		name    string
		args    args
		want    []net.IP
		wantErr bool
	}{
		{
			name: "Invalid Address",
			args: args{
				ips: []string{"invalid ip address"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test 1 (IPv4)",
			args: args{
				ips: []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4"},
			},
			want:    []net.IP{net.IPv4(192, 168, 1, 1), net.IPv4(192, 168, 1, 2), net.IPv4(192, 168, 1, 3), net.IPv4(192, 168, 1, 4)},
			wantErr: false,
		},

		{
			name: "Test 2 (IPv4)",
			args: args{
				ips: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"},
			},
			want:    []net.IP{net.IPv4(10, 0, 0, 1), net.IPv4(10, 0, 0, 2), net.IPv4(10, 0, 0, 3), net.IPv4(10, 0, 0, 4)},
			wantErr: false,
		},

		{
			name: "Test 3 (IPv4)",
			args: args{
				ips: []string{"172.16.0.1", "172.16.0.2", "172.16.0.3", "172.16.0.4"},
			},
			want:    []net.IP{net.IPv4(172, 16, 0, 1), net.IPv4(172, 16, 0, 2), net.IPv4(172, 16, 0, 3), net.IPv4(172, 16, 0, 4)},
			wantErr: false,
		},

		{
			name: "Test 4 (IPv6)",
			args: args{
				ips: []string{"2001:db8::1", "2001:db8::2", "2001:db8::3", "2001:db8::4"},
			},
			want:    []net.IP{net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::2"), net.ParseIP("2001:db8::3"), net.ParseIP("2001:db8::4")},
			wantErr: false,
		},

		{
			name: "Test 5 (IPv6)",
			args: args{
				ips: []string{"fe80::1", "fe80::2", "fe80::3", "fe80::4"},
			},
			want:    []net.IP{net.ParseIP("fe80::1"), net.ParseIP("fe80::2"), net.ParseIP("fe80::3"), net.ParseIP("fe80::4")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringsToIPs(tt.args.ips)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringsToIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringsToIPs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
