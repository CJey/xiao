package netkit

import (
	"net"
	"testing"
)

func TestPrivateIP(t *testing.T) {
	var public = []string{
		"1.1.1.1", "2.2.2.2", "3.3.3.3", "8.8.8.8", "223.255.255.254",
	}
	var private = []string{
		"192.168.0.0", "10.123.0.0", "172.31.0.0", "100.80.80.80", "169.254.0.0", "127.0.0.1",
	}
	var neither = []string{
		"0.0.0.0", "192.0.0.1", "192.0.2.1", "192.88.99.1", "198.18.0.1", "198.51.100.1",
		"203.0.113.1", "224.0.0.1", "240.0.0.1", "255.255.255.255", "::1", "2001:db8::1",
	}
	for _, s := range public {
		if IsPrivateIPv4(net.ParseIP(s)) || !IsPublicIPv4(net.ParseIP(s)) {
			t.Errorf("%s should be public only", s)
		}
	}
	for _, s := range private {
		if !IsPrivateIPv4(net.ParseIP(s)) || IsPublicIPv4(net.ParseIP(s)) {
			t.Errorf("%s should be private only", s)
		}
	}
	for _, s := range neither {
		if IsPrivateIPv4(net.ParseIP(s)) || IsPublicIPv4(net.ParseIP(s)) {
			t.Errorf("%s should be neither private nor public", s)
		}
	}
}
