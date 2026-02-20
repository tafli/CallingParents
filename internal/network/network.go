package network

import (
	"fmt"
	"net"
)

// LanURL returns the best-guess HTTP URL for reaching this server from the
// local network. It picks the first non-loopback IPv4 address found on any
// network interface and combines it with the given listen address.
//
// listenAddr is in the format accepted by net.Listen, e.g. ":8080" or "0.0.0.0:8080".
// If no suitable LAN IP is found, it falls back to "localhost".
func LanURL(listenAddr string) string {
	ip := lanIP()
	_, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		port = "8080"
	}
	return fmt.Sprintf("http://%s:%s", ip, port)
}

func lanIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}

	for _, iface := range ifaces {
		// Skip loopback and down interfaces.
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil {
				continue // skip IPv6
			}
			return ip.String()
		}
	}

	return "localhost"
}
