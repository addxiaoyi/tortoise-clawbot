package netutil

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	defaultPortMin = 8080
	defaultPortMax = 8199
)

// ListenFromSpec binds TCP:
//   - "auto" or empty: 127.0.0.1:8080–8199 中第一个可用端口
//   - "0.0.0.0:auto" / ":auto": 在所有接口上同上端口范围
//   - 其它: 直接传给 net.Listen（如 ":8080", "127.0.0.1:3000"）
func ListenFromSpec(spec string) (net.Listener, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" || strings.EqualFold(spec, "auto") {
		return listenScan("127.0.0.1", defaultPortMin, defaultPortMax)
	}
	if ls := strings.ToLower(spec); strings.HasSuffix(ls, ":auto") {
		host := strings.TrimSpace(spec[:len(spec)-5])
		if host == "" {
			host = "127.0.0.1"
		}
		return listenScan(host, defaultPortMin, defaultPortMax)
	}
	return net.Listen("tcp", spec)
}

func listenScan(host string, portMin, portMax int) (net.Listener, error) {
	for p := portMin; p <= portMax; p++ {
		addr := net.JoinHostPort(host, strconv.Itoa(p))
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			return ln, nil
		}
	}
	return nil, fmt.Errorf("no free TCP port in range %d-%d on host %q", portMin, portMax, host)
}

// HTTPBaseURL returns a canonical http://host:port for the listener (便于本地开发填 SuperTokens apiDomain)。
func HTTPBaseURL(ln net.Listener) string {
	addr := ln.Addr().(*net.TCPAddr)
	host := addr.IP.String()
	if host == "::" {
		host = "127.0.0.1"
	}
	// 监听 0.0.0.0 时对外提示用本机回环，避免前端写 0.0.0.0
	if addr.IP != nil && addr.IP.IsUnspecified() {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%d", host, addr.Port)
}
