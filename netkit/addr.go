package netkit

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
)

// AddrCompletion convert given raw string to standardized address.
// If the ip/port part of raw missing, will use ip and port to complete it.
// ""                => ip:port
// 80, :80           => ip:80
// 1.1.1.1, 1.1.1.1: => 1.1.1.1:port
func AddrCompletion(raw string, ip net.IP, port uint16) (string, error) {
	if raw = strings.TrimSpace(raw); raw == "" {
		raw = net.JoinHostPort("", "")
	} else if rawip := net.ParseIP(raw); rawip != nil {
		raw = net.JoinHostPort(rawip.String(), "")
	} else if n, err := strconv.ParseUint(raw, 10, 64); err == nil {
		raw = net.JoinHostPort("", strconv.FormatUint(n, 10))
	}

	var rawhost, rawport, err = net.SplitHostPort(raw)
	if err != nil {
		return "", fmt.Errorf("invalid address, %w", err)
	}

	if rawhost == "" {
		if ip == nil {
			return "", fmt.Errorf("invalid default ip, should not be nil")
		}
		rawhost = ip.String()
	} else if rawip := net.ParseIP(rawhost); rawip == nil {
		return "", fmt.Errorf("invalid ip part")
	} else {
		rawhost = rawip.String()
	}

	if rawport == "" {
		if port == 0 {
			return "", fmt.Errorf("invalid default port, should not be 0")
		}
		rawport = strconv.FormatUint(uint64(port), 10)
	} else if n, err := strconv.ParseUint(rawport, 10, 64); err != nil {
		return "", fmt.Errorf("invalid port part")
	} else if n == 0 || n > math.MaxUint16 {
		return "", fmt.Errorf("invalid port part")
	} else {
		rawport = strconv.FormatUint(n, 10)
	}

	return net.JoinHostPort(rawhost, rawport), nil
}

// ResolveTCPAddr use AddrCompletion to resolve the raw as a tcp address
func ResolveTCPAddr(raw string, ip net.IP, port uint16) (*net.TCPAddr, error) {
	if addr, err := AddrCompletion(raw, ip, port); err != nil {
		return nil, err
	} else {
		return net.ResolveTCPAddr("tcp", addr)
	}
}

// ResolveUDPAddr use AddrCompletion to resolve the raw as a udp address
func ResolveUDPAddr(raw string, ip net.IP, port uint16) (*net.UDPAddr, error) {
	if addr, err := AddrCompletion(raw, ip, port); err != nil {
		return nil, err
	} else {
		return net.ResolveUDPAddr("udp", addr)
	}
}

// PrettyCSVAddr 会将给定的csv格式地址去重，删除默认端口，并整合为一个csv地址返回。
func PrettyCSVAddr(port uint16, csvs ...string) string {
	var addrs = TrimPort(port, UniqCSVAddrs(port, csvs...)...)
	sort.Strings(addrs)
	return strings.Join(addrs, ",")
}

// UniqCSVAddr 会将给定的csv格式地址去重，
// 同时会移除非法格式的地址，最终将所有有效地址整合为一个csv地址返回。
// 如果给定了port>0，则会对地址的端口部分执行自动补全。
func UniqCSVAddr(port uint16, csvs ...string) string {
	return strings.Join(UniqCSVAddrs(port, csvs...), ",")
}

// UniqCSVAddrs 会将给定的csv格式地址去重，
// 同时会移除非法格式的地址，最终返回地址列表。
// 如果给定了port>0，则会对地址的端口部分执行自动补全。
func UniqCSVAddrs(port uint16, csvs ...string) []string {
	var pstr string
	if port > 0 {
		pstr = strconv.FormatUint(uint64(port), 10)
	}

	var addrs, hit = make([]string, 0), make(map[string]bool)
	var parse = func(baddr []byte) {
		var addr = string(bytes.TrimSpace(baddr))
		if addr == "" {
			return
		}
		if port > 0 {
			if ip := net.ParseIP(addr); ip != nil {
				addr = net.JoinHostPort(ip.String(), pstr)
			} else if h, p, err := net.SplitHostPort(addr); err == nil {
				num, err := strconv.ParseUint(p, 10, 16)
				if err != nil {
					return
				}
				if num == 0 {
					addr = h + ":" + pstr
				}
				// do nothing
			} else if bytes.IndexByte(baddr, ':') < 0 {
				addr += ":" + pstr
			}
		} else {
			if ip := net.ParseIP(addr); ip != nil {
				// do nothing
			} else if h, p, err := net.SplitHostPort(addr); err == nil {
				num, err := strconv.ParseUint(p, 10, 16)
				if err != nil {
					return
				}
				if num == 0 {
					addr = h
				}
			}
		}
		if hit[addr] == false {
			addrs = append(addrs, addr)
			hit[addr] = true
		}
	}

	for i := range csvs {
		if csvs[i] == "" {
			continue
		}
		csv := []byte(csvs[i])
		for {
			to := bytes.IndexByte(csv, ',')
			if to < 0 {
				parse(csv[0:])
				break
			} else {
				parse(csv[0:to])
				csv = csv[to+1:]
			}
		}
	}

	return addrs
}

// TrimPort 将把给定的所有地址中端口号等于给定端口的地址，删除掉端口号，达到删除默认端口号的功能。
func TrimPort(port uint16, addrs ...string) []string {
	var res = make([]string, 0)
	for i := range addrs {
		h, p, err := net.SplitHostPort(addrs[i])
		if err == nil {
			num, err := strconv.ParseUint(p, 10, 16)
			if err == nil && uint16(num) == port {
				res = append(res, h)
				continue
			}
		}
		res = append(res, addrs[i])
	}
	return res
}
