package utils

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func stopListener(ln net.Listener, done <-chan struct{}) {
	_ = ln.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

// relayTCP 双向转发；任一侧结束即关闭两端，避免测试清理死锁。
func relayTCP(a, b net.Conn) {
	var once sync.Once
	closeBoth := func() {
		once.Do(func() {
			_ = a.Close()
			_ = b.Close()
		})
	}
	go func() {
		defer closeBoth()
		_, _ = io.Copy(a, b)
	}()
	_, _ = io.Copy(b, a)
	closeBoth()
}

func startEchoBannerServer(t *testing.T, banner string) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen target: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
				_, _ = conn.Write([]byte(banner))
			}(c)
		}
	}()
	return ln.Addr().String(), func() { stopListener(ln, done) }
}

func startHTTPConnectProxy(t *testing.T, wantUser, wantPass string) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen proxy: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
				br := bufio.NewReader(conn)
				line, err := br.ReadString('\n')
				if err != nil {
					return
				}
				parts := strings.Fields(line)
				if len(parts) < 2 || !strings.EqualFold(parts[0], "CONNECT") {
					_, _ = io.WriteString(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
					return
				}
				target := parts[1]
				authed := wantUser == "" && wantPass == ""
				for {
					h, err := br.ReadString('\n')
					if err != nil {
						return
					}
					if h == "\r\n" || h == "\n" {
						break
					}
					lower := strings.ToLower(h)
					if strings.HasPrefix(lower, "proxy-authorization:") {
						val := strings.TrimSpace(h[len("Proxy-Authorization:"):])
						const p = "basic "
						if len(val) >= len(p) && strings.EqualFold(val[:len(p)], p) {
							raw, decErr := base64.StdEncoding.DecodeString(strings.TrimSpace(val[len(p):]))
							if decErr == nil {
								up := strings.SplitN(string(raw), ":", 2)
								u, pwd := up[0], ""
								if len(up) > 1 {
									pwd = up[1]
								}
								if u == wantUser && pwd == wantPass {
									authed = true
								}
							}
						}
					}
				}
				if !authed {
					_, _ = io.WriteString(conn, "HTTP/1.1 407 Proxy Authentication Required\r\n\r\n")
					return
				}
				remote, err := net.DialTimeout("tcp", target, 3*time.Second)
				if err != nil {
					_, _ = io.WriteString(conn, "HTTP/1.1 502 Bad Gateway\r\n\r\n")
					return
				}
				_, _ = io.WriteString(conn, "HTTP/1.1 200 Connection Established\r\n\r\n")
				_ = conn.SetDeadline(time.Time{})
				// br 可能缓冲了客户端后续字节，转发时优先从 br 读
				go func() {
					defer remote.Close()
					defer conn.Close()
					_, _ = io.Copy(remote, br)
				}()
				_, _ = io.Copy(conn, remote)
				_ = remote.Close()
			}(c)
		}
	}()
	return ln.Addr().String(), func() { stopListener(ln, done) }
}

func startSOCKS5Proxy(t *testing.T) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen socks5: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
				buf := make([]byte, 512)
				n, err := conn.Read(buf)
				if err != nil || n < 3 || buf[0] != 0x05 {
					return
				}
				_, _ = conn.Write([]byte{0x05, 0x00})
				n, err = conn.Read(buf)
				if err != nil || n < 7 || buf[0] != 0x05 || buf[1] != 0x01 {
					return
				}
				var host string
				var port uint16
				switch buf[3] {
				case 0x01:
					if n < 10 {
						return
					}
					host = net.IP(buf[4:8]).String()
					port = binary.BigEndian.Uint16(buf[8:10])
				case 0x03:
					l := int(buf[4])
					if n < 5+l+2 {
						return
					}
					host = string(buf[5 : 5+l])
					port = binary.BigEndian.Uint16(buf[5+l : 7+l])
				default:
					_, _ = conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
					return
				}
				remote, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))), 3*time.Second)
				if err != nil {
					_, _ = conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
					return
				}
				_, _ = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
				_ = conn.SetDeadline(time.Time{})
				relayTCP(conn, remote)
			}(c)
		}
	}()
	return ln.Addr().String(), func() { stopListener(ln, done) }
}

func hostPort(addr string) (host string, port int) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0
	}
	n, _ := strconv.Atoi(p)
	return h, n
}

func TestProxyConfigValidate(t *testing.T) {
	t.Parallel()
	if err := (ProxyConfig{Enabled: false}).Validate(); err != nil {
		t.Fatalf("disabled should pass: %v", err)
	}
	bad := ProxyConfig{Enabled: true, Type: "http", Host: "127.0.0.1", Port: 0}
	if err := bad.Validate(); err == nil {
		t.Fatal("port 0 should fail")
	}
	bad.Port = 70000
	if err := bad.Validate(); err == nil {
		t.Fatal("port 70000 should fail")
	}
	bad.Port = 8080
	bad.Host = "evil.com/path"
	if err := bad.Validate(); err == nil {
		t.Fatal("host with slash should fail")
	}
	bad.Host = "127.0.0.1"
	bad.Type = "ftp"
	if err := bad.Validate(); err == nil {
		t.Fatal("ftp should fail")
	}
	bad.Type = "socks"
	if err := bad.Validate(); err != nil {
		t.Fatalf("socks alias should pass: %v", err)
	}
}

func TestNormalizeProxyType(t *testing.T) {
	t.Parallel()
	if got := NormalizeProxyType("SOCKS"); got != ProxyTypeSOCKS5H {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeProxyType(" Socks5 "); got != ProxyTypeSOCKS5 {
		t.Fatalf("got %q", got)
	}
}

func TestParseProxyURL(t *testing.T) {
	t.Parallel()
	cfg, err := ParseProxyURL("socks5://u:p@127.0.0.1:1080")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || cfg.Type != ProxyTypeSOCKS5 || cfg.Host != "127.0.0.1" || cfg.Port != 1080 {
		t.Fatalf("unexpected %#v", cfg)
	}
	if cfg.Username != "u" || cfg.Password != "p" {
		t.Fatalf("auth %#v", cfg)
	}
	cfg, err = ParseProxyURL("http://127.0.0.1:7890")
	if err != nil || cfg.Port != 7890 {
		t.Fatalf("http parse: %#v %v", cfg, err)
	}
	_, err = ParseProxyURL("not-a-url")
	if err == nil {
		t.Fatal("missing scheme should fail")
	}
}

func TestNewProxyDialerDisabledIsDirect(t *testing.T) {
	target, stop := startEchoBannerServer(t, "HELLO\n")
	defer stop()

	d, err := NewProxyDialer(ProxyConfig{Enabled: false, Timeout: 3 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	c, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	buf := make([]byte, 16)
	n, err := c.Read(buf)
	if err != nil || string(buf[:n]) != "HELLO\n" {
		t.Fatalf("got %q err=%v", buf[:n], err)
	}
}

func TestHTTPConnectProxyRoundTrip(t *testing.T) {
	banner := "220 smtp.example ESMTP ready\r\n"
	target, stopT := startEchoBannerServer(t, banner)
	defer stopT()
	proxyAddr, stopP := startHTTPConnectProxy(t, "", "")
	defer stopP()
	host, port := hostPort(proxyAddr)

	cfg := ProxyConfig{Enabled: true, Type: ProxyTypeHTTP, Host: host, Port: port, Timeout: 5 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := DialViaProxy(ctx, cfg, "tcp", target)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	buf := make([]byte, 64)
	n, err := c.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got != banner {
		t.Fatalf("banner mismatch: got %q want %q", got, banner)
	}
}

func TestHTTPConnectProxyWithAuth(t *testing.T) {
	banner := "OK\n"
	target, stopT := startEchoBannerServer(t, banner)
	defer stopT()
	proxyAddr, stopP := startHTTPConnectProxy(t, "alice", "secret")
	defer stopP()
	host, port := hostPort(proxyAddr)

	cfg := ProxyConfig{Enabled: true, Type: ProxyTypeHTTP, Host: host, Port: port, Timeout: 3 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := DialViaProxy(ctx, cfg, "tcp", target); err == nil {
		t.Fatal("expected auth failure")
	}

	cfg.Username, cfg.Password = "alice", "secret"
	c, err := DialViaProxy(ctx, cfg, "tcp", target)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	buf := make([]byte, 8)
	n, _ := c.Read(buf)
	if string(buf[:n]) != banner {
		t.Fatalf("got %q", buf[:n])
	}
}

func TestSOCKS5ProxyRoundTrip(t *testing.T) {
	banner := "SOCKS-OK\n"
	target, stopT := startEchoBannerServer(t, banner)
	defer stopT()
	proxyAddr, stopP := startSOCKS5Proxy(t)
	defer stopP()
	host, port := hostPort(proxyAddr)

	cfg := ProxyConfig{Enabled: true, Type: ProxyTypeSOCKS5, Host: host, Port: port, Timeout: 5 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := DialViaProxy(ctx, cfg, "tcp", target)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	buf := make([]byte, 32)
	n, err := c.Read(buf)
	if err != nil || string(buf[:n]) != banner {
		t.Fatalf("got %q err=%v", buf[:n], err)
	}
}

func generateTestTLSCertificate(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

func TestDialTLSViaProxy(t *testing.T) {
	cert := generateTestTLSCertificate(t)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_, _ = c.Write([]byte("TLS-HI\n"))
	}()

	proxyAddr, stopP := startHTTPConnectProxy(t, "", "")
	defer stopP()
	ph, pp := hostPort(proxyAddr)

	cfg := ProxyConfig{Enabled: true, Type: ProxyTypeHTTP, Host: ph, Port: pp, Timeout: 5 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tlsCfg := &tls.Config{InsecureSkipVerify: true, ServerName: "localhost"}
	c, err := DialTLSViaProxy(ctx, cfg, "tcp", ln.Addr().String(), tlsCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	buf := make([]byte, 16)
	n, err := c.Read(buf)
	if err != nil || string(buf[:n]) != "TLS-HI\n" {
		t.Fatalf("got %q err=%v", buf[:n], err)
	}
}

func TestProxyURLAndMask(t *testing.T) {
	t.Parallel()
	cfg := ProxyConfig{Enabled: true, Type: "http", Host: "127.0.0.1", Port: 7890, Username: "u", Password: "secret"}
	u, err := cfg.ProxyURL()
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != "http" || u.Hostname() != "127.0.0.1" {
		t.Fatalf("%v", u)
	}
	masked := MaskProxyConfig(cfg)
	if strings.Contains(masked, "secret") {
		t.Fatalf("password leaked: %s", masked)
	}
	if MaskProxyConfig(ProxyConfig{}) != "proxy=off" {
		t.Fatal("disabled mask")
	}
}

func TestNewProxyDialerRejectsBadEnabledConfig(t *testing.T) {
	t.Parallel()
	_, err := NewProxyDialer(ProxyConfig{Enabled: true, Type: "http", Host: "", Port: 1})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrProxyInvalidConfig) && !strings.Contains(err.Error(), "host") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPrefixConnRead(t *testing.T) {
	t.Parallel()
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	go func() {
		_, _ = b.Write([]byte("WORLD"))
		_ = b.Close()
	}()
	c := &prefixConn{Conn: a, prefix: []byte("HELLO")}
	buf := make([]byte, 32)
	n1, _ := c.Read(buf[:3])
	n2, _ := c.Read(buf[3:])
	got := string(buf[:n1+n2])
	if !strings.HasPrefix(got, "HELLO") {
		t.Fatalf("got %q", got)
	}
}
