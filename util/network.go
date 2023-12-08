package util

import (
	"fmt"
	"net"
)

func getAllInterfaceAddresses() ([]net.Addr, error) {
	var validAddresses []net.Addr
	// Get a list of network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error getting network interfaces:", err)
		return nil, err
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("Error getting addresses for interface %s: %v\n", iface.Name, err)
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil && !v.IP.IsLoopback() {
					// This is an IPv4 address that isn't a loopback
					validAddresses = append(validAddresses, addr)
				}
			}
		}
		if validAddresses == nil {
			continue
		}
	}
	return validAddresses, nil
}

func GetAllBroadcastAddresses() ([]net.IP, error) {
	var broadcastAddresses []net.IP
	interfaceAddresses, err := getAllInterfaceAddresses()
	if err != nil {
		return nil, err
	}

	for _, address := range interfaceAddresses {
		broadcast := getIPBroadcast(address)
		if broadcast != nil {
			broadcastAddresses = append(broadcastAddresses, broadcast)
		}
	}

	return broadcastAddresses, nil
}

func getIPBroadcast(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPNet:
		if v.IP.To4() != nil {
			// This is an IPv4 address
			broadcast := calculateIPv4Broadcast(v)
			//fmt.Printf("IPv4 %s, Broadcast Address: %s\n", v.IP.To4(), broadcast)
			return broadcast
		}
	}
	return nil
}

func calculateIPv4Broadcast(ipNet *net.IPNet) net.IP {
	ip := ipNet.IP.To4()
	mask := ipNet.Mask
	broadcast := net.IP(make([]byte, 4))
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}
	return broadcast
}

func IPsToStrings(input []net.IP) []string {
	if len(input) == 0 {
		return nil
	}
	ips := make([]string, len(input))
	for i, ip := range input {
		ips[i] = ip.String()
	}
	return ips
}

func StringsToIPs(ips []string) ([]net.IP, error) {
	parsedIps := make([]net.IP, len(ips))
	for i, ip := range ips {
		parsedIp := net.ParseIP(ip)
		if parsedIp == nil {
			return nil, fmt.Errorf("invalid ip address: %s", ip)
		}
		parsedIps[i] = parsedIp
	}
	return parsedIps, nil
}
