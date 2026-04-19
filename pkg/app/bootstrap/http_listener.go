package bootstrap

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

func openHTTPListener(server *http.Server, host, port string) (net.Listener, error) {
	if server == nil {
		return nil, fmt.Errorf("http server is required")
	}
	return net.Listen("tcp", httpListenAddress(host, port))
}

func httpListenAddress(host, port string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "0.0.0.0"
	}
	port = strings.TrimSpace(port)
	if port == "" || strings.EqualFold(port, "auto") || port == "0" {
		port = "0"
	}
	return net.JoinHostPort(host, port)
}

func describeHTTPListener(listener net.Listener, host string) (addr string, port string, healthzURL string) {
	if listener == nil {
		return "", "", ""
	}

	addr = listener.Addr().String()
	_, listenPort, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, "", ""
	}

	displayHost := strings.TrimSpace(host)
	switch displayHost {
	case "", "0.0.0.0", "::":
		displayHost = "127.0.0.1"
	}

	return addr, listenPort, "http://" + net.JoinHostPort(displayHost, listenPort) + "/healthz"
}
