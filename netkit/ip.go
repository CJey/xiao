package netkit

import (
	"bytes"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	_MyManIPLast time.Time
	_MyManIP     net.IP
	_MyWanIPLast time.Time
	_MyWanIP     net.IP
)

// MyManIP 返回本机上的管理网络IP地址。
// 原理上会尝试去查询访问10.0.0.1这个地址时所使用的来源IP，取此IP为管理网络地址。
// 如果获取失败，或者查询发现使用的IP地址并非私有IP地址，则返回127.0.0.1。
func MyManIP() (ip net.IP) {
	const (
		TARGET = "10.0.0.1"
	)

	if len(_MyManIP) > 0 && _MyManIP[0] != 127 {
		ip = make(net.IP, len(_MyManIP))
		copy(ip, _MyManIP)
		return
	}

	if len(_MyManIP) > 0 && time.Since(_MyManIPLast) < 3*time.Second {
		ip = make(net.IP, len(_MyManIP))
		copy(ip, _MyManIP)
		return
	}

	defer func() {
		if ip == nil || !ip.IsPrivate() {
			ip = net.IPv4(127, 0, 0, 1)
		}
		var clone = make(net.IP, len(ip))
		copy(clone, ip)
		_MyManIP = clone
		_MyManIPLast = time.Now()
	}()

	return RouteSource(TARGET)
}

// MyWanIP 返回本机上绑定的公网IP地址。
// 原理上会尝试去查询访问1.1.1.1这个地址时所使用的来源IP，取此IP为公网IP地址。
// 如果获取失败，或者查询发现使用的IP地址是私有IP地址，则返回127.0.0.1。
func MyWanIP() (ip net.IP) {
	const (
		TARGET = "1.1.1.1"
	)

	if _MyWanIP != nil && time.Now().Sub(_MyWanIPLast) < 3*time.Second {
		ip = make(net.IP, len(_MyWanIP))
		copy(ip, _MyWanIP)
		return
	}

	defer func() {
		if ip == nil || ip.IsPrivate() {
			ip = net.IPv4(127, 0, 0, 1)
		}
		var clone = make(net.IP, len(ip))
		copy(clone, ip)
		_MyWanIP = clone
		_MyWanIPLast = time.Now()
	}()

	return RouteSource(TARGET)
}

// RouteSource 使用命令行的方式，探测本机访问target时会使用的来源ip地址。
// 如果探测失败或者没能得到正确的ip，返回nil
func RouteSource(target string) (ip net.IP) {
	if runtime.GOOS == "darwin" {
		// route -n -v get 172.16.0.1
		// u: inet 172.16.0.1; u: link ; RTM_GET: Report Metrics: len 128, pid: 0, seq 1, errno 0, flags:<UP,GATEWAY,HOST,STATIC>
		// locks:  inits:
		// sockaddrs: <DST,IFP>
		//  172.16.0.1
		//    route to: 172.16.0.1
		// destination: 128.0.0.0
		//        mask: 128.0.0.0
		//     gateway: 10.94.12.9
		//   interface: utun10
		//       flags: <UP,GATEWAY,DONE,STATIC,PRCLONING>
		//  recvpipe  sendpipe  ssthresh  rtt,msec    rttvar  hopcount      mtu     expire
		//        0         0         0         0         0         0      1412         0
		//
		// locks:  inits:
		// sockaddrs: <DST,GATEWAY,NETMASK,IFP,IFA>
		//  128.0.0.0 10.94.12.9 128.0.0.0 utun10 10.94.12.10
		// we need 'IFA' of last line to find
		var cmd = exec.Command("route", "-nv", "get", target)
		if out, err := cmd.Output(); err == nil {
			var hit bool
			for _, line := range strings.Split(string(out), "\n") {
				var oss = strings.Fields(line)
				if hit {
					if src := net.ParseIP(oss[len(oss)-1]); src != nil {
						return src
					}
					break
				}
				hit = len(oss) == 2 && oss[0] == "sockaddrs:" &&
					oss[1] == "<DST,GATEWAY,NETMASK,IFP,IFA>"
			}
		}
	} else if runtime.GOOS == "linux" {
		// ip route get 172.16.0.1
		// output e.g.
		// 172.16.0.1 via 172.18.133.193 dev manbr  src 172.18.133.235
		//    cache
		// we need 'src 172.18.133.235', and 172.18.133.235 is the target to find
		var cmd = exec.Command("ip", "route", "get", target)
		if out, err := cmd.Output(); err == nil {
			// output parse, find src <target>
			var oss = bytes.Fields(out)
			for i, v := range oss {
				if string(v) != "src" {
					continue
				}
				if len(oss)-1 >= i+1 {
					if src := net.ParseIP(string(oss[i+1])); src != nil {
						return src
					}
				}
				break
			}
		}
	}
	return nil
}
