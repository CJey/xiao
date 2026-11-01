package netkit

import (
	"errors"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestTCPing(t *testing.T) {
	var lsn, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lsn.Close()
	go func() {
		for {
			var c, err = lsn.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	var port = strconv.Itoa(lsn.Addr().(*net.TCPAddr).Port)

	for _, c := range []struct {
		source, target string
		ok             bool
	}{
		{"", "127.0.0.1:" + port, true},
		{"", "localhost:" + port, true},
		{"127.0.0.1", "127.0.0.1:" + port, true},
		{"0.0.0.0", "127.0.0.1:" + port, false},
		{"::1", "127.0.0.1:" + port, false},
	} {
		var rtt, err = TCPingV2(c.source, c.target, time.Second)
		if (err == nil) != c.ok || c.ok && (rtt <= 0 || rtt >= time.Second) {
			t.Errorf("TCPingV2(%q, %q) = %s, %v", c.source, c.target, rtt, err)
		}
	}
}

func TestICMPing(t *testing.T) {
	var rtt, err = ICMPing("127.0.0.1", time.Second)
	if errors.Is(err, os.ErrPermission) {
		t.Skipf("gid is out of net.ipv4.ping_group_range, %v", err)
	}
	if err != nil || rtt <= 0 {
		t.Fatalf("ICMPing = %s, %v", rtt, err)
	}

	for _, c := range []struct {
		source string
		ok     bool
	}{
		{"127.0.0.1", true},
		{"0.0.0.0", false},
		{"::1", false},
	} {
		if _, err := ICMPingV2(c.source, "127.0.0.1", time.Second); (err == nil) != c.ok {
			t.Errorf("ICMPingV2(%q) = %v", c.source, err)
		}
	}
}
