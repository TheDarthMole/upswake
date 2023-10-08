package util

import "net"

func IsValidMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	if err == nil {
		return true
	}
	return false
}
