package netkit

import (
	"net"
	"slices"
	"sync"
	"time"
)

// 管理网地址一经探得即长期沿用，其余结果(含回落值127.0.0.1)只缓存一个周期，过期重探
const probeTTL = 3 * time.Second

var probed struct {
	sync.Mutex
	man, wan         net.IP
	manLast, wanLast time.Time
}

// MyManIP 返回本机上的管理网络IP地址，即访问10.0.0.1时选用的源地址。
// 探测失败或该地址并非RFC1918私有地址时返回127.0.0.1
func MyManIP() net.IP {
	probed.Lock()
	defer probed.Unlock()
	if probed.man == nil || (probed.man.IsLoopback() && time.Since(probed.manLast) >= probeTTL) {
		var ip = RouteSource("10.0.0.1")
		if ip == nil || !ip.IsPrivate() {
			ip = net.IPv4(127, 0, 0, 1)
		}
		probed.man, probed.manLast = ip, time.Now()
	}
	return slices.Clone(probed.man)
}

// MyWanIP 返回本机上绑定的公网IP地址，即访问1.1.1.1时选用的源地址。
// 探测失败或该地址并非公网地址时返回127.0.0.1
func MyWanIP() net.IP {
	probed.Lock()
	defer probed.Unlock()
	if probed.wan == nil || time.Since(probed.wanLast) >= probeTTL {
		var ip = RouteSource("1.1.1.1")
		if ip == nil || !IsPublicIPv4(ip) {
			ip = net.IPv4(127, 0, 0, 1)
		}
		probed.wan, probed.wanLast = ip, time.Now()
	}
	return slices.Clone(probed.wan)
}

// RouteSource 探测本机访问target时选用的源地址，探测失败返回nil。
// 对UDP套接字执行connect只触发内核选路，并不发包
func RouteSource(target string) net.IP {
	var c, err = net.Dial("udp", net.JoinHostPort(target, "9"))
	if err != nil {
		return nil
	}
	defer c.Close()
	// 与net.ParseIP的结果保持同一形式
	return c.LocalAddr().(*net.UDPAddr).IP.To16()
}
