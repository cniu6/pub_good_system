package public

import (
	"fst/backend/app/services"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// GeoController 公开地理/区号相关接口
type GeoController struct{}

// NewGeoController 创建地理控制器
func NewGeoController() *GeoController {
	return &GeoController{}
}

// PhoneCountryResponse 默认手机号国家探测结果
type PhoneCountryResponse struct {
	CountryCode     string `json:"country_code"`
	DialCode        string `json:"dial_code"`
	NameZH          string `json:"name_zh"`
	NameEN          string `json:"name_en"`
	Source          string `json:"source"` // cn_only | header | ip | lang | default
	MobileCNOnly    bool   `json:"mobile_cn_only"`
	IPDetectEnabled bool   `json:"ip_detect_enabled"`
}

// DialCountriesResponse 区号列表
type DialCountriesResponse struct {
	Items []utils.DialCountry `json:"items"`
}

// DetectPhoneCountry 探测默认手机号国家（大厂式区号选择的默认值）
// @Summary 探测手机号默认国家
// @Description 优先级：仅大陆强制CN → CDN国家头 →（可选）IP查询 → 语言 → 美国+1保底。query.lang 建议传前端当前语言（zhCN/enUS）
// @Tags Public-配置
// @Produce json
// @Param lang query string false "当前界面语言，如 zhCN / enUS / zh-CN"
// @Success 200 {object} utils.Response{data=PhoneCountryResponse}
// @Router /api/v1/public/geo/phone-country [get]
func (ctrl *GeoController) DetectPhoneCountry(c *gin.Context) {
	cnOnly := services.GetGlobalMobileCNOnly()
	ipDetect := services.GetGlobalMobileIPCountryDetect()
	// IP 探测仅在「允许国际号」时生效
	if cnOnly {
		ipDetect = false
	}

	lang := c.Query("lang")
	country, source := utils.ResolvePhoneCountry(utils.ResolvePhoneCountryOptions{
		CNOnly:          cnOnly,
		IPDetectEnabled: ipDetect,
		ClientIP:        utils.GetClientIP(c),
		Lang:            lang,
		GetHeader:       c.GetHeader,
	})

	utils.Success(c, PhoneCountryResponse{
		CountryCode:     country.Code,
		DialCode:        country.DialCode,
		NameZH:          country.NameZH,
		NameEN:          country.NameEN,
		Source:          source,
		MobileCNOnly:    cnOnly,
		IPDetectEnabled: ipDetect,
	})
}

// ListDialCountries 返回常用国家区号列表（前后端可共用；前端也有静态副本）
// @Summary 国家区号列表
// @Tags Public-配置
// @Produce json
// @Success 200 {object} utils.Response{data=DialCountriesResponse}
// @Router /api/v1/public/geo/dial-countries [get]
func (ctrl *GeoController) ListDialCountries(c *gin.Context) {
	utils.Success(c, DialCountriesResponse{Items: utils.ListDialCountries()})
}

// RegisterRoutes 注册地理相关公开路由
func (ctrl *GeoController) RegisterRoutes(group *gin.RouterGroup) {
	geo := group.Group("/geo")
	{
		geo.GET("/phone-country", ctrl.DetectPhoneCountry)
		geo.GET("/dial-countries", ctrl.ListDialCountries)
	}
}
