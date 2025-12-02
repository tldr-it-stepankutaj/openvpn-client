package utils

import (
	"fmt"
	"net"
	"strings"
)

// CIDRToNetmask converts CIDR notation to IP and netmask
// Example: "192.168.1.0/24" -> "192.168.1.0 255.255.255.0"
func CIDRToNetmask(cidr string) (string, error) {
	// If no slash, treat as single IP
	if !strings.Contains(cidr, "/") {
		ip := net.ParseIP(cidr)
		if ip == nil {
			return "", fmt.Errorf("invalid IP: %s", cidr)
		}
		return cidr + " 255.255.255.255", nil
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	// Get network address and mask
	ip := ipNet.IP.String()
	mask := net.IP(ipNet.Mask).String()

	return ip + " " + mask, nil
}

// IsDefaultRoute checks if the CIDR represents a default route
func IsDefaultRoute(cidr string) bool {
	return cidr == "0.0.0.0/0" || cidr == "0/0" || cidr == "::/0"
}
