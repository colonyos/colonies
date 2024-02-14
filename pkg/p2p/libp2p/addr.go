package libp2p

import (
	"errors"
	"net"
)

// isPublicIP checks if an IP address is considered public.
func isPublicIP(ip net.IP) bool {
	// Loopback, link-local, multicast, and private IP ranges are considered non-public
	return !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsPrivate()
}

// getBestIPAddress iterates over available network interfaces and their IP addresses,
// returning the most suitable one based on predefined criteria.
func getBestIPAddress() (string, error) {
	var preferredIP string

	// Iterate over all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// Check if the interface is up and skip otherwise
		if iface.Flags&net.FlagUp == 0 {
			continue // Interface is down
		}

		// Skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue // Skip interfaces with address retrieval errors
		}

		for _, addr := range addrs {
			var ip net.IP

			// Check the type of the IP address
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Use the first public IP address found
			if isPublicIP(ip) {
				return ip.String(), nil
			}

			// If no public IP is found, keep the first private IP as a fallback
			if preferredIP == "" && ip.IsPrivate() {
				preferredIP = ip.String()
			}
		}
	}

	if preferredIP != "" {
		return preferredIP, nil
	}

	return "", errors.New("no suitable IP address found")
}
