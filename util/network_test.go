package util

//
//import (
//	"net"
//	"reflect"
//	"testing"
//)
//
//func Test_calculateIPv4Broadcast(t *testing.T) {
//	type args struct {
//		ipNet *net.IPNet
//	}
//	tests := []struct {
//		name string
//		args args
//		want net.IP
//	}{
//		{
//			name: "Class A Subnet",
//			args: args{ipNet: &net.IPNet{
//				IP:   net.IP{1, 52, 42, 4},
//				Mask: net.IPMask{255, 255, 255, 0},
//			}},
//			want: net.IP{1, 52, 42, 255},
//		},
//		{
//			name: "Class B Subnet",
//			args: args{ipNet: &net.IPNet{
//				IP:   net.IP{128, 37, 43, 6},
//				Mask: net.IPMask{255, 255, 255, 0},
//			}},
//			want: net.IP{128, 37, 43, 255},
//		},
//		{
//			name: "Class C Subnet",
//			args: args{ipNet: &net.IPNet{
//				IP:   net.IP{192, 2, 3, 25},
//				Mask: net.IPMask{255, 255, 255, 0},
//			}},
//			want: net.IP{192, 2, 3, 255},
//		},
//		{
//			name: "CIDR /32",
//			args: args{ipNet: &net.IPNet{
//				IP:   net.IP{192, 2, 3, 25},
//				Mask: net.IPMask{255, 255, 255, 255},
//			}},
//			want: net.IP{192, 2, 3, 25},
//		},
//		{
//			name: "CIDR /24",
//			args: args{ipNet: &net.IPNet{
//				IP:   net.IP{192, 2, 3, 25},
//				Mask: net.IPMask{255, 255, 255, 0},
//			}},
//			want: net.IP{192, 2, 3, 255},
//		},
//		{
//			name: "CIDR /16",
//			args: args{ipNet: &net.IPNet{
//				IP:   net.IP{192, 2, 3, 25},
//				Mask: net.IPMask{255, 0, 0, 0},
//			}},
//			want: net.IP{192, 255, 255, 255},
//		},
//		{
//			name: "CIDR /27",
//			args: args{ipNet: &net.IPNet{
//				IP:   net.IP{192, 2, 3, 25},
//				Mask: net.IPMask{255, 255, 255, 224},
//			}},
//			want: net.IP{192, 2, 3, 31},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := calculateIPv4Broadcast(tt.args.ipNet); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("calculateIPv4Broadcast() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
