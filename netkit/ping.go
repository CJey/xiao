package netkit

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"net"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// TCPing detect rtt time by establishing a tcp connection.
func TCPing(target string, timeout time.Duration) (time.Duration, error) {
	return TCPingV2("", target, timeout)
}

// TCPingV2 同TCPing，可指定源IP，为空时与TCPing一致。
// 源IP决定目标地址族，故不接受0.0.0.0、::这类未指定地址
func TCPingV2(source, target string, timeout time.Duration) (time.Duration, error) {
	var src, err = parseSource(source)
	if err != nil {
		return 0, err
	}
	var dialer = net.Dialer{Timeout: timeout}
	if src != nil {
		dialer.LocalAddr = &net.TCPAddr{IP: src}
	}

	// 每次尝试在connect之前登记起点，rtt取最终连上的那一次，不含域名解析与失败尝试的耗时
	var mu sync.Mutex
	var starts = make(map[string]time.Time)
	dialer.ControlContext = func(_ context.Context, _, address string, _ syscall.RawConn) error {
		mu.Lock()
		starts[address] = time.Now()
		mu.Unlock()
		return nil
	}

	var begin = time.Now()
	var c, e = dialer.Dial("tcp", target)
	var end = time.Now()
	if e != nil {
		return 0, e
	}
	var raddr = c.RemoteAddr().String()
	c.Close()

	mu.Lock()
	defer mu.Unlock()
	if start, ok := starts[raddr]; ok {
		begin = start
	}
	return end.Sub(begin), nil
}

// ICMPing 以icmp echo探测rtt。使用非特权的icmp数据报套接字，无需root或CAP_NET_RAW，
// 但进程的gid须落在sysctl net.ipv4.ping_group_range之内，ipv6同受此项约束
func ICMPing(target string, timeout time.Duration) (time.Duration, error) {
	return ICMPingV2("", target, timeout)
}

// ICMPingV2 同ICMPing，可指定源IP，为空时与ICMPing一致。
// 源IP决定目标地址族，故不接受0.0.0.0、::这类未指定地址
func ICMPingV2(source, target string, timeout time.Duration) (time.Duration, error) {
	var src, err = parseSource(source)
	if err != nil {
		return 0, err
	}

	var deadline = time.Now().Add(timeout)
	var ctx, cancel = context.WithDeadline(context.Background(), deadline)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, target)
	if err != nil {
		return 0, err
	}
	if len(addrs) == 0 {
		return 0, fmt.Errorf("no address of %s", target)
	}

	// 没给源IP时由解析结果的首个地址决定地址族，否则源IP和目的IP必须同族
	var addr = &addrs[0]
	if src != nil {
		var v4 = src.To4() != nil
		addr = nil
		for i := range addrs {
			if v4 == (addrs[i].IP.To4() != nil) {
				addr = &addrs[i]
				break
			}
		}
		if addr == nil {
			return 0, fmt.Errorf("no address of %s to match the source ip %s", target, source)
		}
	}

	var network, local string
	var reqType, repType icmp.Type
	if addr.IP.To4() != nil {
		reqType, repType = ipv4.ICMPTypeEcho, ipv4.ICMPTypeEchoReply
		network, local = "udp4", "0.0.0.0"
	} else {
		reqType, repType = ipv6.ICMPTypeEchoRequest, ipv6.ICMPTypeEchoReply
		network, local = "udp6", "::"
	}
	if src != nil {
		local = src.String()
	}

	conn, err := icmp.ListenPacket(network, local)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// echo id由内核改写，无从核对；payload带随机串，只认自己的应答
	var payload = binary.BigEndian.AppendUint64([]byte("xiao/netkit ping "), rand.Uint64())
	var req, _ = (&icmp.Message{
		Type: reqType,
		Body: &icmp.Echo{Seq: 1, Data: payload},
	}).Marshal(nil)

	if err = conn.SetReadDeadline(deadline); err != nil {
		return 0, err
	}

	var begin = time.Now()
	var dst = &net.UDPAddr{IP: addr.IP, Zone: addr.Zone}
	if _, err = conn.WriteTo(req, dst); err != nil {
		return 0, err
	}

	var buf = make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			return 0, err
		}
		var rtt = time.Since(begin)

		msg, err := icmp.ParseMessage(reqType.Protocol(), buf[:n])
		if err != nil || msg.Type != repType {
			continue
		}
		if rep, ok := msg.Body.(*icmp.Echo); ok && rep.Seq == 1 && bytes.Equal(rep.Data, payload) {
			return rtt, nil
		}
	}
}

func parseSource(source string) (net.IP, error) {
	if source == "" {
		return nil, nil
	}
	var src = net.ParseIP(source)
	if src == nil || src.IsUnspecified() {
		return nil, fmt.Errorf("invalid source ip, %s", source)
	}
	return src, nil
}
