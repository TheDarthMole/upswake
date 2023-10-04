package network

import (
	"net"
)

func GetNetworkInterfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

/**
type InterfaceIP struct {
	Interface net.Interface
	IPs       []net.Addr
}

func getAllInterfaceAddresses() ([]InterfaceIP, error) {
	var interfaceBroadcasts []InterfaceIP
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

		interfaceIp := InterfaceIP{
			Interface: iface,
			IPs:       addrs,
		}
		interfaceBroadcasts = append(interfaceBroadcasts, interfaceIp)
	}
	return interfaceBroadcasts, nil
}

// GetAllBroadcastAddresses TODO:This could be simplified by not getting the broadcast address for each interface
// but instead feed the interface and mac to the wol NewRawClient function
func GetAllBroadcastAddresses() ([]net.IP, error) {
	var broadcastAddresses []net.IP
	interfaceBroadcasts, err := getAllInterfaceAddresses()
	if err != nil {
		return nil, err
	}

	for _, interfaceBroadcast := range interfaceBroadcasts {
		for _, addr := range interfaceBroadcast.IPs {
			broadcast := getIPBroadcast(addr)
			if broadcast != nil {
				broadcastAddresses = append(broadcastAddresses, broadcast)
			}
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
*/
