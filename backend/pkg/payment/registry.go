// Package payment 全局支付通道注册表。
//
// 各通道在 init() 中调用 RegisterChannel 注册，启动时由 pay_balance 等插件
// 的 blank import 触发。运行时通过 GetProvider(channelType) 获取实例。
package payment

import (
	"sort"
	"strings"
	"sync"
)

// PayTypeMeta 二级支付方式元数据（如 alipay/wxpay）
type PayTypeMeta struct {
	Value string `json:"value"` // 调用值，如 alipay
	Name  string `json:"name"`  // 显示名，如 支付宝
}

// SignTypeMeta 签名算法元数据
type SignTypeMeta struct {
	Value string `json:"value"` // 调用值，如 MD5 / RSA
	Name  string `json:"name"`  // 显示名
}

// DeviceMeta 设备类型元数据
type DeviceMeta struct {
	Value string `json:"value"` // 调用值，如 pc / mobile
	Name  string `json:"name"`  // 显示名
}

// ConfigFieldOption select 类型的选项
type ConfigFieldOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ConfigField 通道扩展配置字段 schema，供前端动态渲染表单
type ConfigField struct {
	Name        string              `json:"name"`
	Label       string              `json:"label"`
	Type        string              `json:"type"` // input / textarea / select
	Required    bool                `json:"required"`
	Secret      bool                `json:"secret"`
	Placeholder string              `json:"placeholder"`
	Options     []ConfigFieldOption `json:"options"` // 仅 select 类型有效
}

// ChannelVersionMeta 通道版本元数据，同一通道可有多个版本（V1 MD5 / V2 RSA）
type ChannelVersionMeta struct {
	Version      string         `json:"version"`
	Name         string         `json:"name"`
	SignTypes    []SignTypeMeta `json:"signTypes"`
	ConfigFields []ConfigField  `json:"configFields"`
}

// ChannelMeta 通道类型元数据
type ChannelMeta struct {
	Type              string               `json:"type"`
	Name              string               `json:"name"`
	Currency          string               `json:"currency"`
	PayTypes          []PayTypeMeta        `json:"payTypes"`
	Devices           []DeviceMeta         `json:"devices"`
	SupportCashbox    bool                 `json:"supportCashbox"`
	DefaultNotifyPath string               `json:"defaultNotifyPath"`
	Versions          []ChannelVersionMeta `json:"versions"`
	ConfigFields      []ConfigField        `json:"configFields"` // 通用/网关级动态配置字段
}

// GetVersionMeta 取通道指定版本的元数据，version 为空时取第一个
type channelEntry struct {
	Meta       ChannelMeta
	ProviderFn func() Provider
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]channelEntry)
)

// RegisterChannel 注册一个通道类型（含元数据 + provider 工厂）
// 通常在 payment/<channel>/ 包的 init() 中调用
func RegisterChannel(meta ChannelMeta, providerFn func() Provider) {
	if meta.Type == "" {
		panic("payment.RegisterChannel: meta.Type cannot be empty")
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[meta.Type] = channelEntry{Meta: meta, ProviderFn: providerFn}
}

// GetChannelMeta 取通道元数据，未注册返回 nil
func GetChannelMeta(channelType string) *ChannelMeta {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if e, ok := registry[channelType]; ok {
		m := e.Meta
		return &m
	}
	return nil
}

// GetProvider 取通道 provider，未注册返回 nil
func GetProvider(channelType string) Provider {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if e, ok := registry[channelType]; ok && e.ProviderFn != nil {
		return e.ProviderFn()
	}
	return nil
}

// ListChannelMetas 列出所有已注册通道元数据，按 Type 升序
func ListChannelMetas() []ChannelMeta {
	registryMu.RLock()
	defer registryMu.RUnlock()
	metas := make([]ChannelMeta, 0, len(registry))
	for _, e := range registry {
		metas = append(metas, e.Meta)
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].Type < metas[j].Type
	})
	return metas
}

// GetVersionMeta 取通道指定版本的元数据
func (m *ChannelMeta) GetVersionMeta(version string) *ChannelVersionMeta {
	v := strings.TrimSpace(version)
	if v == "" {
		if len(m.Versions) > 0 {
			return &m.Versions[0]
		}
		return nil
	}
	for i := range m.Versions {
		if m.Versions[i].Version == v {
			return &m.Versions[i]
		}
	}
	return nil
}
