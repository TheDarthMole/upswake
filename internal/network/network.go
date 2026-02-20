package network

import (
	"errors"
	"net"
	"strings"
)

var ErrNoInterfaceFound = errors.New("no local interface found")

func getAllInterfaceAddresses() ([]net.Addr, error) {
	// Get a list of network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	return filterAddressesFromInterfaces(interfaces)
}

func filterAddressesFromInterfaces(interfaces []net.Interface) ([]net.Addr, error) {
	var validAddresses []net.Addr
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
				validAddresses = append(validAddresses, addr)
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
	if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
		return calculateIPv4Broadcast(ipNet)
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

func HasLocalInterface(ip string) error {
	ifaces, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, iface := range ifaces {
		parsedIP, _, err := net.ParseCIDR(iface.String())
		if err != nil {
			return err
		}
		if strings.Contains(ip, parsedIP.String()) {
			return nil
		}
	}

	return ErrNoInterfaceFound
}
