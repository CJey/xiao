package netkit

import (
	"net"
)

func IsPrivateIPv4(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		return false
	}
	return ip[0] == 127 || ip[0] == 10 ||
		(ip[0] == 100 && ip[1]&0xc0 == 64) ||
		(ip[0] == 172 && ip[1]&0xf0 == 16) ||
		(ip[0] == 192 && ip[1] == 168) ||
		(ip[0] == 169 && ip[1] == 254)
}

func IsPublicIPv4(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		return false
	}
	return !(ip[0] == 127 || ip[0] == 10 ||
		(ip[0] == 100 && ip[1]&0xc0 == 64) ||
		(ip[0] == 172 && ip[1]&0xf0 == 16) ||
		(ip[0] == 192 && ip[1] == 168) ||
		(ip[0] == 169 && ip[1] == 254))
}
