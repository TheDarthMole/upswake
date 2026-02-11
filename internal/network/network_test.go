package network

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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
			got := getIPBroadcast(tt.args.addr)
			if !got.Equal(tt.want) {
				t.Errorf("getIPBroadcast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAllBroadcastAddresses(t *testing.T) {
	got, err := GetAllBroadcastAddresses()
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func Test_filterAddressesFromInterfaces(t *testing.T) {
	type args struct {
		interfaces []net.Interface
	}
	tests := []struct {
		name  string
		args  args
		want  []net.Addr
		error error
	}{
		{
			name: "No Interfaces",
			args: args{
				interfaces: []net.Interface{},
			},
			want:  []net.Addr(nil),
			error: nil,
		},
		{
			name: "Loopback Interface",
			args: args{
				interfaces: []net.Interface{
					{
						Index:        1,
						MTU:          65536,
						Name:         "lo",
						HardwareAddr: nil,
						Flags:        net.FlagLoopback | net.FlagUp,
					},
				},
			},
			want:  []net.Addr(nil),
			error: nil,
		},
		// Unfortunately, we cannot easily mock net.Interfaces() to return custom interfaces.
		// Therefore, we will just test with the actual interfaces on the system running the tests.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterAddressesFromInterfaces(tt.args.interfaces)
			assert.ErrorIs(t, err, tt.error)
			assert.Equalf(t, tt.want, got, "filterAddressesFromInterfaces(%v)", tt.args.interfaces)
		})
	}
}
