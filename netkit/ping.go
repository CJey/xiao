package netkit

import (
	"net"
	"time"
)

// TCPing detect rtt time by establishing a tcp connection.
func TCPing(target string, timeout time.Duration) (rtt time.Duration, err error) {
	var begin = time.Now()
	var c, e = net.DialTimeout("tcp", target, timeout)
	if e != nil {
		return 0, e
	}

	rtt = time.Since(begin)
	c.Close()
	return rtt, nil
}
