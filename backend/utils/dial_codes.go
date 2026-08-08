package utils

import (
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"
)

// DialCountry 国际区号条目（ISO 3166-1 alpha-2）
type DialCountry struct {
	Code     string `json:"code"`      // CN / US
	DialCode string `json:"dial_code"` // 86 / 1（不含 +）
	NameZH   string `json:"name_zh"`
	NameEN   string `json:"name_en"`
}
//@name 国际区号

// 常用国家/地区区号（覆盖主要市场；前端同构一份）
var dialCountries = []DialCountry{
	{Code: "CN", DialCode: "86", NameZH: "中国大陆", NameEN: "China Mainland"},
	{Code: "HK", DialCode: "852", NameZH: "中国香港", NameEN: "Hong Kong"},
	{Code: "MO", DialCode: "853", NameZH: "中国澳门", NameEN: "Macao"},
	{Code: "TW", DialCode: "886", NameZH: "中国台湾", NameEN: "Taiwan"},
	{Code: "US", DialCode: "1", NameZH: "美国", NameEN: "United States"},
	{Code: "CA", DialCode: "1", NameZH: "加拿大", NameEN: "Canada"},
	{Code: "GB", DialCode: "44", NameZH: "英国", NameEN: "United Kingdom"},
	{Code: "AU", DialCode: "61", NameZH: "澳大利亚", NameEN: "Australia"},
	{Code: "NZ", DialCode: "64", NameZH: "新西兰", NameEN: "New Zealand"},
	{Code: "JP", DialCode: "81", NameZH: "日本", NameEN: "Japan"},
	{Code: "KR", DialCode: "82", NameZH: "韩国", NameEN: "South Korea"},
	{Code: "SG", DialCode: "65", NameZH: "新加坡", NameEN: "Singapore"},
	{Code: "MY", DialCode: "60", NameZH: "马来西亚", NameEN: "Malaysia"},
	{Code: "TH", DialCode: "66", NameZH: "泰国", NameEN: "Thailand"},
	{Code: "VN", DialCode: "84", NameZH: "越南", NameEN: "Vietnam"},
	{Code: "PH", DialCode: "63", NameZH: "菲律宾", NameEN: "Philippines"},
	{Code: "ID", DialCode: "62", NameZH: "印度尼西亚", NameEN: "Indonesia"},
	{Code: "IN", DialCode: "91", NameZH: "印度", NameEN: "India"},
	{Code: "AE", DialCode: "971", NameZH: "阿联酋", NameEN: "United Arab Emirates"},
	{Code: "SA", DialCode: "966", NameZH: "沙特阿拉伯", NameEN: "Saudi Arabia"},
	{Code: "TR", DialCode: "90", NameZH: "土耳其", NameEN: "Turkey"},
	{Code: "RU", DialCode: "7", NameZH: "俄罗斯", NameEN: "Russia"},
	{Code: "DE", DialCode: "49", NameZH: "德国", NameEN: "Germany"},
	{Code: "FR", DialCode: "33", NameZH: "法国", NameEN: "France"},
	{Code: "IT", DialCode: "39", NameZH: "意大利", NameEN: "Italy"},
	{Code: "ES", DialCode: "34", NameZH: "西班牙", NameEN: "Spain"},
	{Code: "PT", DialCode: "351", NameZH: "葡萄牙", NameEN: "Portugal"},
	{Code: "NL", DialCode: "31", NameZH: "荷兰", NameEN: "Netherlands"},
	{Code: "BE", DialCode: "32", NameZH: "比利时", NameEN: "Belgium"},
	{Code: "CH", DialCode: "41", NameZH: "瑞士", NameEN: "Switzerland"},
	{Code: "AT", DialCode: "43", NameZH: "奥地利", NameEN: "Austria"},
	{Code: "SE", DialCode: "46", NameZH: "瑞典", NameEN: "Sweden"},
	{Code: "NO", DialCode: "47", NameZH: "挪威", NameEN: "Norway"},
	{Code: "DK", DialCode: "45", NameZH: "丹麦", NameEN: "Denmark"},
	{Code: "FI", DialCode: "358", NameZH: "芬兰", NameEN: "Finland"},
	{Code: "IE", DialCode: "353", NameZH: "爱尔兰", NameEN: "Ireland"},
	{Code: "PL", DialCode: "48", NameZH: "波兰", NameEN: "Poland"},
	{Code: "CZ", DialCode: "420", NameZH: "捷克", NameEN: "Czechia"},
	{Code: "BR", DialCode: "55", NameZH: "巴西", NameEN: "Brazil"},
	{Code: "MX", DialCode: "52", NameZH: "墨西哥", NameEN: "Mexico"},
	{Code: "AR", DialCode: "54", NameZH: "阿根廷", NameEN: "Argentina"},
	{Code: "CL", DialCode: "56", NameZH: "智利", NameEN: "Chile"},
	{Code: "CO", DialCode: "57", NameZH: "哥伦比亚", NameEN: "Colombia"},
	{Code: "ZA", DialCode: "27", NameZH: "南非", NameEN: "South Africa"},
	{Code: "EG", DialCode: "20", NameZH: "埃及", NameEN: "Egypt"},
	{Code: "NG", DialCode: "234", NameZH: "尼日利亚", NameEN: "Nigeria"},
	{Code: "IL", DialCode: "972", NameZH: "以色列", NameEN: "Israel"},
	{Code: "PK", DialCode: "92", NameZH: "巴基斯坦", NameEN: "Pakistan"},
	{Code: "BD", DialCode: "880", NameZH: "孟加拉国", NameEN: "Bangladesh"},
	{Code: "UA", DialCode: "380", NameZH: "乌克兰", NameEN: "Ukraine"},
}

// DefaultDialCountryCode 保底国家（用户要求默认美国 +1）
const DefaultDialCountryCode = "US"

var dialByCode map[string]DialCountry

func init() {
	dialByCode = make(map[string]DialCountry, len(dialCountries))
	for _, c := range dialCountries {
		dialByCode[c.Code] = c
	}
}

// ListDialCountries 返回区号列表副本
func ListDialCountries() []DialCountry {
	out := make([]DialCountry, len(dialCountries))
	copy(out, dialCountries)
	return out
}

// GetDialCountry 按 ISO 取条目；未知返回 false
func GetDialCountry(code string) (DialCountry, bool) {
	c, ok := dialByCode[strings.ToUpper(strings.TrimSpace(code))]
	return c, ok
}

// CountryFromLanguage 根据界面语言猜默认国家（zh→CN，其它非空→US；空串返回空表示交给保底）
func CountryFromLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return ""
	}
	lang = strings.ReplaceAll(lang, "_", "-")
	switch {
	case strings.HasPrefix(lang, "zh"):
		return "CN"
	default:
		return DefaultDialCountryCode
	}
}

// CountryFromCDNHeaders 从 CDN/反代头读取国家码（Cloudflare / CloudFront 等）
func CountryFromCDNHeaders(getHeader func(string) string) string {
	headers := []string{
		"CF-IPCountry",
		"CloudFront-Viewer-Country",
		"X-Country-Code",
		"X-AppEngine-Country",
		"X-Vercel-IP-Country",
	}
	for _, h := range headers {
		v := strings.ToUpper(strings.TrimSpace(getHeader(h)))
		if len(v) == 2 && v != "XX" && v != "T1" && unicode.IsLetter(rune(v[0])) {
			if _, ok := GetDialCountry(v); ok {
				return v
			}
		}
	}
	return ""
}

// ipCountryProvider 单个 IP→国家码 查询服务：名称 + URL 构造 + 响应解析。
type ipCountryProvider struct {
	name  string
	build func(ip string) string
	parse func(body []byte) string
}

// freeIPCountryProviders 免费、无需 API Key 的 IP→国家码 服务池。
// 查询时随机打散顺序，避免所有请求都打到同一家：既分摊各家免费额度，也容忍单点故障。
// 【说明】ip-api.com 免费额度仅 HTTP；其余几家支持免费 HTTPS。字段名均已核对真实响应。
var freeIPCountryProviders = []ipCountryProvider{
	{
		name:  "ip-api.com",
		build: func(ip string) string { return "http://ip-api.com/json/" + ip + "?fields=status,countryCode" },
		parse: func(b []byte) string {
			var r struct {
				Status      string `json:"status"`
				CountryCode string `json:"countryCode"`
			}
			if json.Unmarshal(b, &r) == nil && strings.EqualFold(r.Status, "success") {
				return r.CountryCode
			}
			return ""
		},
	},
	{
		name:  "country.is",
		build: func(ip string) string { return "https://api.country.is/" + ip },
		parse: func(b []byte) string {
			var r struct {
				Country string `json:"country"`
			}
			if json.Unmarshal(b, &r) == nil {
				return r.Country
			}
			return ""
		},
	},
	{
		name:  "ipwho.is",
		build: func(ip string) string { return "https://ipwho.is/" + ip },
		parse: func(b []byte) string {
			var r struct {
				Success     bool   `json:"success"`
				CountryCode string `json:"country_code"`
			}
			if json.Unmarshal(b, &r) == nil && r.Success {
				return r.CountryCode
			}
			return ""
		},
	},
	{
		name:  "freeipapi.com",
		build: func(ip string) string { return "https://freeipapi.com/api/json/" + ip },
		parse: func(b []byte) string {
			var r struct {
				CountryCode string `json:"countryCode"`
			}
			if json.Unmarshal(b, &r) == nil {
				return r.CountryCode
			}
			return ""
		},
	},
}

// customIPCountryLookup 可选的自定义/付费 IP 归属查询钩子（内部持有，未导出）。
// 为避免 utils→services 循环依赖，由上层（如管理端可配置的付费 API）在启动/保存设置时注入。
// 返回合法国家码时优先采用；返回空串则回退到免费服务池。可用 BuildJSONCountryLookup 快速构造。
//
// 【并发安全】管理端保存设置时可能随时调用 SetCustomIPCountryLookup 覆盖钩子，
// 与此同时可能正有请求在 LookupIPCountry 里读取它——裸的包级 func 变量在 Go 里并发读写
// 属于数据竞争（-race 会报），因此用 RWMutex 包一层，读写都走 Get/Set。
var (
	customIPCountryLookupMu sync.RWMutex
	customIPCountryLookup   func(ip string) string
)

// SetCustomIPCountryLookup 并发安全地设置/清除自定义 IP 归属查询钩子。
// 传 nil 表示清除（回退到免费服务池）。供管理端保存「自定义 IP 查询 API」配置时调用。
func SetCustomIPCountryLookup(fn func(ip string) string) {
	customIPCountryLookupMu.Lock()
	customIPCountryLookup = fn
	customIPCountryLookupMu.Unlock()
}

// GetCustomIPCountryLookup 并发安全地读取当前自定义钩子（未设置时返回 nil）。
func GetCustomIPCountryLookup() func(ip string) string {
	customIPCountryLookupMu.RLock()
	defer customIPCountryLookupMu.RUnlock()
	return customIPCountryLookup
}

// maxIPCountryProvidersPerCall 单次查询最多尝试的免费服务数量（含随机首选 + 失败回退）。
// 控制在 2 家：既有容错，又把请求路径延迟上限压在可接受范围。
const maxIPCountryProvidersPerCall = 2

// LookupIPCountry 通过公网 IP 粗略查国家（返回 ISO alpha-2 国家码，失败返回空）。
// 仅在管理端开启「IP 自动匹配国家」时调用。
// 策略：先试自定义/付费 provider（若注入）；否则在免费服务池里随机打散、依次尝试，
// 命中即返回，避免所有流量集中打到单一免费 API。
func LookupIPCountry(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsLinkLocalUnicast() {
		return ""
	}

	// 1) 优先自定义/付费 provider（若上层注入）
	if hook := GetCustomIPCountryLookup(); hook != nil {
		if code := normalizeDialCountryCode(hook(ip)); code != "" {
			return code
		}
	}

	// 2) 免费服务池随机打散后依次尝试，命中即止
	client := &http.Client{Timeout: 2 * time.Second}
	tried := 0
	for _, idx := range rand.Perm(len(freeIPCountryProviders)) {
		if tried >= maxIPCountryProvidersPerCall {
			break
		}
		tried++
		if code := queryIPCountryProvider(client, freeIPCountryProviders[idx], ip); code != "" {
			return code
		}
	}
	return ""
}

// queryIPCountryProvider 向单个 provider 发起查询并返回校验通过的国家码（失败返回空）。
func queryIPCountryProvider(client *http.Client, p ipCountryProvider, ip string) string {
	req, err := http.NewRequest(http.MethodGet, p.build(ip), nil)
	if err != nil {
		return ""
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil {
		return ""
	}
	return normalizeDialCountryCode(p.parse(body))
}

// normalizeDialCountryCode 统一大写并校验是否为已知区号国家；未知返回空。
func normalizeDialCountryCode(raw string) string {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if code == "" {
		return ""
	}
	if _, ok := GetDialCountry(code); ok {
		return code
	}
	return ""
}

// BuildJSONCountryLookup 用「URL 模板 + 顶层国家码字段名」构造一个查询函数。
// urlTemplate 用 {ip} 占位；countryField 为响应 JSON 顶层存放国家码的字段名。
// 供上层把管理端可配置的（含付费）API 注入到 CustomIPCountryLookup，无需改动本包。
func BuildJSONCountryLookup(urlTemplate, countryField string) func(string) string {
	urlTemplate = strings.TrimSpace(urlTemplate)
	countryField = strings.TrimSpace(countryField)
	if urlTemplate == "" || countryField == "" || !strings.Contains(urlTemplate, "{ip}") {
		return nil
	}
	return func(ip string) string {
		client := &http.Client{Timeout: 2 * time.Second}
		u := strings.ReplaceAll(urlTemplate, "{ip}", url.QueryEscape(ip))
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return ""
		}
		resp, err := client.Do(req)
		if err != nil {
			return ""
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return ""
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 16384))
		if err != nil {
			return ""
		}
		var m map[string]any
		if json.Unmarshal(body, &m) != nil {
			return ""
		}
		if v, ok := m[countryField].(string); ok {
			return v
		}
		return ""
	}
}

// ResolvePhoneCountry 解析默认手机号国家。
// 优先级：仅大陆强制 CN → CDN 头 →（可选）IP 查询 → 语言 → 美国保底
func ResolvePhoneCountry(opts ResolvePhoneCountryOptions) (country DialCountry, source string) {
	if opts.CNOnly {
		c, _ := GetDialCountry("CN")
		return c, "cn_only"
	}
	if code := CountryFromCDNHeaders(opts.GetHeader); code != "" {
		c, _ := GetDialCountry(code)
		return c, "header"
	}
	if opts.IPDetectEnabled {
		if code := LookupIPCountry(opts.ClientIP); code != "" {
			c, _ := GetDialCountry(code)
			return c, "ip"
		}
	}
	if code := CountryFromLanguage(opts.Lang); code != "" {
		if c, ok := GetDialCountry(code); ok {
			return c, "lang"
		}
	}
	c, _ := GetDialCountry(DefaultDialCountryCode)
	return c, "default"
}

// ResolvePhoneCountryOptions 国家解析参数
type ResolvePhoneCountryOptions struct {
	CNOnly          bool
	IPDetectEnabled bool
	ClientIP        string
	Lang            string
	GetHeader       func(string) string
}

// ComposeE164 区号 + 国内号码 → E.164（会去掉国内号开头的 0）
func ComposeE164(dialCode, national string) string {
	dialCode = strings.TrimLeft(strings.TrimSpace(dialCode), "+")
	var b strings.Builder
	for _, r := range national {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	n := b.String()
	if strings.HasPrefix(n, "0") {
		n = n[1:]
	}
	if dialCode == "" || n == "" {
		return ""
	}
	return "+" + dialCode + n
}
