package netkit

import (
	"net"
	"sync"
	"testing"
	"time"
)

// 探测结果因环境而异，只断言取值范围
func TestMyManIP(t *testing.T) {
	if ip := MyManIP(); !ip.IsPrivate() && !ip.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Errorf("MyManIP = %s", ip)
	}
}

func TestMyWanIP(t *testing.T) {
	if ip := MyWanIP(); !IsPublicIPv4(ip) && !ip.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Errorf("MyWanIP = %s", ip)
	}
}

func TestMyManIPCache(t *testing.T) {
	t.Cleanup(resetProbed)

	// 过期的回落值须重探
	probed.man, probed.manLast = net.IPv4(127, 0, 0, 1), time.Now().Add(-time.Hour)
	MyManIP()
	if time.Since(probed.manLast) > time.Minute {
		t.Error("expired fallback was not reprobed")
	}

	// 有效地址长期沿用
	var valid, last = net.ParseIP("10.1.2.3"), time.Now().Add(-time.Hour)
	probed.man, probed.manLast = valid, last
	if got := MyManIP(); !got.Equal(valid) || !probed.manLast.Equal(last) {
		t.Errorf("valid address was not kept, got %s", got)
	}
}

func TestProbeConcurrent(t *testing.T) {
	t.Cleanup(resetProbed)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			MyManIP()
			MyWanIP()
		})
	}
	wg.Wait()
}

func TestRouteSource(t *testing.T) {
	if ip := RouteSource("127.0.0.1"); !ip.IsLoopback() {
		t.Errorf("RouteSource(127.0.0.1) = %s", ip)
	}
}

func resetProbed() {
	probed.Lock()
	defer probed.Unlock()
	probed.man, probed.wan = nil, nil
	probed.manLast, probed.wanLast = time.Time{}, time.Time{}
}
