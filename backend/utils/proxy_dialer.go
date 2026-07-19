package utils

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	xproxy "golang.org/x/net/proxy"
)

// 代理类型（小写归一化后匹配）
const (
	ProxyTypeHTTP    = "http"    // HTTP CONNECT
	ProxyTypeHTTPS   = "https"   // TLS 到代理后再 CONNECT
	ProxyTypeSOCKS5  = "socks5"  // SOCKS5（本地解析目标域名）
	ProxyTypeSOCKS5H = "socks5h" // SOCKS5（由代理解析域名，推荐）
	ProxyTypeSOCKS   = "socks"   // socks 别名 → socks5h
)

// 拨号默认超时（与常见 net.Dialer 一致）
const (
	maxProxyPort            = 65535
	minProxyPort            = 1
	defaultProxyDialTimeout = 30 * time.Second
)

var (
	ErrProxyDisabled      = errors.New("proxy: disabled")
	ErrProxyInvalidConfig = errors.New("proxy: invalid config")
	ErrProxyUnsupported   = errors.New("proxy: unsupported type")
)

// ContextDialer 带 context 的拨号器（SMTP / HTTP / 任意 TCP 均可复用）
type ContextDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// ProxyConfig 通用代理配置。
// Enabled=false 时 NewProxyDialer 返回直连 dialer，其它字段可忽略（开关优先，减少误配风险）。
type ProxyConfig struct {
	Enabled  bool
	Type     string // http / https / socks5 / socks5h / socks
	Host     string
	Port     int
	Username string
	Password string
	// Timeout 拨号超时（含连代理 + 握手）。<=0 使用默认 30s。
	Timeout time.Duration
	// TLSConfig 仅 https 代理使用；nil 时用 ServerName=Host 的默认配置。
	TLSConfig *tls.Config
}

var registerHTTPOnce sync.Once

// ensureHTTPProxySchemes 向 x/net/proxy 注册 http/https CONNECT（SOCKS5 已内置）。
func ensureHTTPProxySchemes() {
	registerHTTPOnce.Do(func() {
		xproxy.RegisterDialerType("http", newHTTPConnectDialerFromURL)
		xproxy.RegisterDialerType("https", newHTTPConnectDialerFromURL)
	})
}

// NormalizeProxyType 归一化代理类型；空串返回空串。
func NormalizeProxyType(raw string) string {
	t := strings.ToLower(strings.TrimSpace(raw))
	switch t {
	case ProxyTypeSOCKS:
		return ProxyTypeSOCKS5H
	case "socks5a": // 少数客户端别名
		return ProxyTypeSOCKS5H
	default:
		return t
	}
}

// Validate 校验配置。Enabled=false 时直接通过。
func (c ProxyConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	typ := NormalizeProxyType(c.Type)
	switch typ {
	case ProxyTypeHTTP, ProxyTypeHTTPS, ProxyTypeSOCKS5, ProxyTypeSOCKS5H:
	default:
		return fmt.Errorf("%w: type %q", ErrProxyUnsupported, c.Type)
	}

	host := strings.TrimSpace(c.Host)
	if host == "" {
		return fmt.Errorf("%w: host is empty", ErrProxyInvalidConfig)
	}
	// host 只允许主机名/IP，不要带 scheme 或路径
	if strings.ContainsAny(host, "/\\?#") || strings.Contains(host, "://") {
		return fmt.Errorf("%w: host must be hostname or IP only", ErrProxyInvalidConfig)
	}
	if c.Port < minProxyPort || c.Port > maxProxyPort {
		return fmt.Errorf("%w: port out of range (%d-%d)", ErrProxyInvalidConfig, minProxyPort, maxProxyPort)
	}
	return nil
}

// dialTimeout 返回拨号超时。
func (c ProxyConfig) dialTimeout() time.Duration {
	if c.Timeout <= 0 {
		return defaultProxyDialTimeout
	}
	return c.Timeout
}

// ProxyURL 构造代理 URL（含账密）。Enabled=false 返回 nil。
func (c ProxyConfig) ProxyURL() (*url.URL, error) {
	if !c.Enabled {
		return nil, ErrProxyDisabled
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	typ := NormalizeProxyType(c.Type)
	hostPort := net.JoinHostPort(strings.TrimSpace(c.Host), strconv.Itoa(c.Port))

	u := &url.URL{
		Scheme: typ,
		Host:   hostPort,
	}
	if c.Username != "" || c.Password != "" {
		u.User = url.UserPassword(c.Username, c.Password)
	}
	return u, nil
}

// ParseProxyURL 从 URL 解析代理配置（支持 http/https/socks5/socks5h，可含账密）。
// 例: socks5://user:pass@127.0.0.1:1080  /  http://127.0.0.1:7890
func ParseProxyURL(raw string) (ProxyConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ProxyConfig{}, fmt.Errorf("%w: empty url", ErrProxyInvalidConfig)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ProxyConfig{}, fmt.Errorf("%w: %v", ErrProxyInvalidConfig, err)
	}
	typ := NormalizeProxyType(u.Scheme)
	if typ == "" {
		return ProxyConfig{}, fmt.Errorf("%w: missing scheme", ErrProxyInvalidConfig)
	}
	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		switch typ {
		case ProxyTypeHTTPS:
			portStr = "443"
		case ProxyTypeHTTP:
			portStr = "8080"
		default:
			portStr = "1080"
		}
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return ProxyConfig{}, fmt.Errorf("%w: bad port", ErrProxyInvalidConfig)
	}
	cfg := ProxyConfig{
		Enabled: true,
		Type:    typ,
		Host:    host,
		Port:    port,
	}
	if u.User != nil {
		cfg.Username = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}
	if err := cfg.Validate(); err != nil {
		return ProxyConfig{}, err
	}
	return cfg, nil
}

// netContextDialer 包装 *net.Dialer 以满足 ContextDialer。
type netContextDialer struct {
	d *net.Dialer
}

func (n netContextDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return n.d.DialContext(ctx, network, address)
}

// xproxyContextAdapter 将 x/net/proxy.Dialer 适配为 ContextDialer。
type xproxyContextAdapter struct {
	d       xproxy.Dialer
	timeout time.Duration
}

func (a xproxyContextAdapter) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	// 优先走 ContextDialer
	if cd, ok := a.d.(xproxy.ContextDialer); ok {
		return cd.DialContext(ctx, network, address)
	}
	// 无 context 接口时用超时 + 取消监听兜底，避免永久阻塞
	type result struct {
		c   net.Conn
		err error
	}
	ch := make(chan result, 1)
	go func() {
		c, err := a.d.Dial(network, address)
		ch <- result{c, err}
	}()
	select {
	case <-ctx.Done():
		go func() {
			r := <-ch
			if r.c != nil {
				_ = r.c.Close()
			}
		}()
		return nil, ctx.Err()
	case r := <-ch:
		return r.c, r.err
	}
}

// NewDirectDialer 返回直连 dialer（开关关闭时使用）。
func NewDirectDialer(timeout time.Duration) ContextDialer {
	if timeout <= 0 {
		timeout = defaultProxyDialTimeout
	}
	return netContextDialer{d: &net.Dialer{Timeout: timeout}}
}

// NewProxyDialer 根据配置创建拨号器。
// Enabled=false → 直连（不校验 host/port，避免开关关闭时因脏配置报错）。
func NewProxyDialer(cfg ProxyConfig) (ContextDialer, error) {
	timeout := cfg.dialTimeout()
	if !cfg.Enabled {
		return NewDirectDialer(timeout), nil
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ensureHTTPProxySchemes()

	forward := &net.Dialer{Timeout: timeout}
	u, err := cfg.ProxyURL()
	if err != nil {
		return nil, err
	}

	// https 代理：把 TLS 配置注入到我们的 CONNECT dialer（RegisterDialerType 回调读不到 TLSConfig）
	// 因此 https 走专用构造，不经过 FromURL。
	typ := NormalizeProxyType(cfg.Type)
	if typ == ProxyTypeHTTPS {
		d, err := newHTTPConnectDialer(u, forward, cfg.TLSConfig)
		if err != nil {
			return nil, err
		}
		return xproxyContextAdapter{d: d, timeout: timeout}, nil
	}

	d, err := xproxy.FromURL(u, forward)
	if err != nil {
		return nil, fmt.Errorf("proxy: create dialer: %w", err)
	}
	return xproxyContextAdapter{d: d, timeout: timeout}, nil
}

// DialViaProxy 便捷拨号：按配置建立到 address 的 TCP 连接。
func DialViaProxy(ctx context.Context, cfg ProxyConfig, network, address string) (net.Conn, error) {
	d, err := NewProxyDialer(cfg)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	// 若调用方未设 deadline，用配置超时兜底
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.dialTimeout())
		defer cancel()
	}
	return d.DialContext(ctx, network, address)
}

// DialTLSViaProxy 先经代理拨通 TCP，再做 TLS 握手（SMTP 隐式 SSL / HTTPS 目标等）。
func DialTLSViaProxy(ctx context.Context, cfg ProxyConfig, network, address string, tlsCfg *tls.Config) (*tls.Conn, error) {
	raw, err := DialViaProxy(ctx, cfg, network, address)
	if err != nil {
		return nil, err
	}
	if tlsCfg == nil {
		host, _, _ := net.SplitHostPort(address)
		tlsCfg = &tls.Config{ServerName: host}
	}
	tlsConn := tls.Client(raw, tlsCfg)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = raw.Close()
		return nil, err
	}
	return tlsConn, nil
}

// MaskProxyConfig 日志用脱敏摘要（不输出密码明文）。
func MaskProxyConfig(cfg ProxyConfig) string {
	if !cfg.Enabled {
		return "proxy=off"
	}
	typ := NormalizeProxyType(cfg.Type)
	auth := "noauth"
	if cfg.Username != "" {
		auth = "user=" + cfg.Username
	}
	return fmt.Sprintf("proxy=%s://%s auth=%s", typ, net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)), auth)
}
