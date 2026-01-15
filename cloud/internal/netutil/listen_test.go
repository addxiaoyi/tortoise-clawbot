package netutil

import (
	"net"
	"strings"
	"testing"
)

func TestHTTPBaseURL(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	u := HTTPBaseURL(ln)
	if !strings.HasPrefix(u, "http://127.0.0.1:") {
		t.Fatalf("unexpected base URL: %s", u)
	}
}

func TestListenFromSpec_autoRange(t *testing.T) {
	ln, err := ListenFromSpec("auto")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().(*net.TCPAddr)
	if addr.Port < defaultPortMin || addr.Port > defaultPortMax {
		t.Fatalf("port %d outside default range %d-%d", addr.Port, defaultPortMin, defaultPortMax)
	}
}
