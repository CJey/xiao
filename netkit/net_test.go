package netkit

import (
	"net"
	"testing"
)

func TestPrivateIP(t *testing.T) {
	a := []string{
		"1.1.1.1", "2.2.2.2", "3.3.3.3",
	}
	b := []string{
		"192.168.0.0", "10.123.0.0", "172.31.0.0", "100.80.80.80", "169.254.0.0",
	}
	for _, s := range a {
		if IsPrivateIPv4(net.ParseIP(s)) {
			t.Errorf("%s is not private ipv4!", s)
		}
		if !IsPublicIPv4(net.ParseIP(s)) {
			t.Errorf("%s is public ipv4!", s)
		}
	}
	for _, s := range b {
		if !IsPrivateIPv4(net.ParseIP(s)) {
			t.Errorf("%s is private ipv4!", s)
		}
		if IsPublicIPv4(net.ParseIP(s)) {
			t.Errorf("%s is not public ipv4!", s)
		}
	}
}
