package utils

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
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

// LookupIPCountry 通过公网 IP 粗略查国家（ip-api.com，无 key；失败返回空）
// 仅在管理端开启「IP 自动匹配国家」时调用。
func LookupIPCountry(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsLinkLocalUnicast() {
		return ""
	}

	client := &http.Client{Timeout: 2 * time.Second}
	url := "http://ip-api.com/json/" + ip + "?fields=status,countryCode"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return ""
	}
	var parsedResp struct {
		Status      string `json:"status"`
		CountryCode string `json:"countryCode"`
	}
	if json.Unmarshal(body, &parsedResp) != nil {
		return ""
	}
	if !strings.EqualFold(parsedResp.Status, "success") {
		return ""
	}
	code := strings.ToUpper(strings.TrimSpace(parsedResp.CountryCode))
	if _, ok := GetDialCountry(code); ok {
		return code
	}
	return ""
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
