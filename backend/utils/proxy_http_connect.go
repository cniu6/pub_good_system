package utils

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	xproxy "golang.org/x/net/proxy"
)

// httpConnectDialer 通过 HTTP/HTTPS 代理的 CONNECT 方法建立隧道。
// 实现 xproxy.Dialer / ContextDialer，可被 RegisterDialerType 注册。
type httpConnectDialer struct {
	proxyURL  *url.URL
	tlsConfig *tls.Config // 非 nil 表示先 TLS 连代理（https proxy）
	forward   xproxy.Dialer
	username  string
	password  string
}

func newHTTPConnectDialerFromURL(u *url.URL, forward xproxy.Dialer) (xproxy.Dialer, error) {
	return newHTTPConnectDialer(u, forward, nil)
}

func newHTTPConnectDialer(u *url.URL, forward xproxy.Dialer, tlsCfg *tls.Config) (xproxy.Dialer, error) {
	if u == nil {
		return nil, fmt.Errorf("%w: nil proxy url", ErrProxyInvalidConfig)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != ProxyTypeHTTP && scheme != ProxyTypeHTTPS {
		return nil, fmt.Errorf("%w: expect http/https, got %q", ErrProxyUnsupported, u.Scheme)
	}
	if forward == nil {
		forward = &net.Dialer{Timeout: defaultProxyDialTimeout}
	}

	d := &httpConnectDialer{
		proxyURL: u,
		forward:  forward,
	}
	if u.User != nil {
		d.username = u.User.Username()
		d.password, _ = u.User.Password()
	}
	if scheme == ProxyTypeHTTPS {
		host := u.Hostname()
		if tlsCfg != nil {
			d.tlsConfig = tlsCfg.Clone()
			if d.tlsConfig.ServerName == "" {
				d.tlsConfig.ServerName = host
			}
		} else {
			d.tlsConfig = &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		}
	}
	return d, nil
}

func (d *httpConnectDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *httpConnectDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("proxy: http connect only supports tcp, got %q", network)
	}
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, fmt.Errorf("%w: empty target address", ErrProxyInvalidConfig)
	}
	if _, _, err := net.SplitHostPort(address); err != nil {
		return nil, fmt.Errorf("%w: target must be host:port: %v", ErrProxyInvalidConfig, err)
	}

	proxyAddr := d.proxyURL.Host
	if !strings.Contains(proxyAddr, ":") {
		if d.tlsConfig != nil {
			proxyAddr = net.JoinHostPort(proxyAddr, "443")
		} else {
			proxyAddr = net.JoinHostPort(proxyAddr, "8080")
		}
	}

	var conn net.Conn
	var err error
	if cd, ok := d.forward.(xproxy.ContextDialer); ok {
		conn, err = cd.DialContext(ctx, "tcp", proxyAddr)
	} else {
		conn, err = d.forward.Dial("tcp", proxyAddr)
	}
	if err != nil {
		return nil, fmt.Errorf("proxy: dial proxy %s: %w", proxyAddr, err)
	}

	if err := ctx.Err(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	if d.tlsConfig != nil {
		tlsConn := tls.Client(conn, d.tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("proxy: tls to proxy: %w", err)
		}
		conn = tlsConn
	}

	tunneled, err := d.doConnect(ctx, conn, address)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return tunneled, nil
}

// doConnect 发送 CONNECT 并校验 200；若 bufio 多读了字节则用 prefixConn 回灌。
func (d *httpConnectDialer) doConnect(ctx context.Context, conn net.Conn, address string) (net.Conn, error) {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(defaultProxyDialTimeout)
	}
	_ = conn.SetDeadline(deadline)
	clearDeadline := true
	defer func() {
		if clearDeadline {
			_ = conn.SetDeadline(time.Time{})
		}
	}()

	var reqBuilder strings.Builder
	reqBuilder.Grow(256 + len(address))
	reqBuilder.WriteString("CONNECT ")
	reqBuilder.WriteString(address)
	reqBuilder.WriteString(" HTTP/1.1\r\nHost: ")
	reqBuilder.WriteString(address)
	reqBuilder.WriteString("\r\nProxy-Connection: Keep-Alive\r\n")
	if d.username != "" || d.password != "" {
		token := base64.StdEncoding.EncodeToString([]byte(d.username + ":" + d.password))
		reqBuilder.WriteString("Proxy-Authorization: Basic ")
		reqBuilder.WriteString(token)
		reqBuilder.WriteString("\r\n")
	}
	reqBuilder.WriteString("\r\n")

	if _, err := io.WriteString(conn, reqBuilder.String()); err != nil {
		return nil, fmt.Errorf("proxy: write CONNECT: %w", err)
	}

	// 按标准 HTTP 解析 CONNECT 响应（与常见代理客户端一致）
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodConnect})
	if err != nil {
		return nil, fmt.Errorf("proxy: read CONNECT response: %w", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proxy: CONNECT to %s failed: %s", address, resp.Status)
	}

	_ = conn.SetDeadline(time.Time{})
	clearDeadline = false

	if buffered := br.Buffered(); buffered > 0 {
		peeked, peekErr := br.Peek(buffered)
		if peekErr != nil {
			return nil, fmt.Errorf("proxy: peek buffered: %w", peekErr)
		}
		return &prefixConn{Conn: conn, prefix: append([]byte(nil), peeked...)}, nil
	}
	return conn, nil
}

// prefixConn 把 bufio 多读的字节塞回后续 Read，避免吞掉 SMTP/TLS banner。
type prefixConn struct {
	net.Conn
	prefix []byte
}

func (c *prefixConn) Read(b []byte) (int, error) {
	if len(c.prefix) > 0 {
		n := copy(b, c.prefix)
		c.prefix = c.prefix[n:]
		if len(c.prefix) == 0 {
			c.prefix = nil
		}
		return n, nil
	}
	return c.Conn.Read(b)
}
