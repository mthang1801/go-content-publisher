package bootstrap

import (
	"net"
	"net/http"
	"testing"
)

func TestHTTPListenAddressUsesConfiguredPort(t *testing.T) {
	if got := httpListenAddress("127.0.0.1", "8085"); got != "127.0.0.1:8085" {
		t.Fatalf("expected fixed port listen address, got %q", got)
	}
}

func TestHTTPListenAddressNormalizesAutoPort(t *testing.T) {
	cases := []string{"", "0", "auto", "AUTO"}
	for _, port := range cases {
		if got := httpListenAddress("127.0.0.1", port); got != "127.0.0.1:0" {
			t.Fatalf("expected auto port %q to normalize to :0, got %q", port, got)
		}
	}
}

func TestDescribeHTTPListenerUsesLocalHealthzForWildcardHost(t *testing.T) {
	listener := staticListener{addr: staticAddr("0.0.0.0:40151")}
	addr, port, healthzURL := describeHTTPListener(listener, "0.0.0.0")
	if addr == "" {
		t.Fatal("expected addr")
	}
	if port == "" {
		t.Fatal("expected port")
	}
	expected := "http://127.0.0.1:" + port + "/healthz"
	if healthzURL != expected {
		t.Fatalf("expected healthz %q, got %q", expected, healthzURL)
	}
}

func nilSafeHTTPServer() *http.Server {
	return &http.Server{}
}

type staticListener struct {
	addr net.Addr
}

func (s staticListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (s staticListener) Close() error              { return nil }
func (s staticListener) Addr() net.Addr            { return s.addr }

type staticAddr string

func (a staticAddr) Network() string { return "tcp" }
func (a staticAddr) String() string  { return string(a) }
