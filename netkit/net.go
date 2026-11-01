package netkit

import (
	"net"
	"net/netip"
	"slices"
)

// 私有地址：回环、RFC1918、CGNAT与链路本地
var privateIPv4 = []netip.Prefix{
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("169.254.0.0/16"),
}

// 既非私有也不可公网路由的地址：IANA特殊用途地址表中的其余条目，以及组播与保留段
var specialIPv4 = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
}

func IsPrivateIPv4(ip net.IP) bool {
	var addr, ok = addrOf4(ip)
	return ok && contains(privateIPv4, addr)
}

func IsPublicIPv4(ip net.IP) bool {
	var addr, ok = addrOf4(ip)
	return ok && !contains(privateIPv4, addr) && !contains(specialIPv4, addr)
}

func addrOf4(ip net.IP) (netip.Addr, bool) {
	if ip4 := ip.To4(); ip4 != nil {
		return netip.AddrFrom4([4]byte(ip4)), true
	}
	return netip.Addr{}, false
}

func contains(prefixes []netip.Prefix, addr netip.Addr) bool {
	return slices.ContainsFunc(prefixes, func(p netip.Prefix) bool { return p.Contains(addr) })
}
