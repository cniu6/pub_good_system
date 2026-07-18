package user

import (
	crypto_rand "crypto/rand"
	"fmt"
	"fst/backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// calculateDaysJoined 计算加入天数
func calculateDaysJoined(join_time *int64) int {
	if join_time == nil {
		return 0
	}
	join := time.Unix(*join_time, 0)
	return int(time.Since(join).Hours() / 24)
}

// generateCode 生成6位随机数字验证码
// 使用 crypto/rand 强制随机，不依赖固定格式或种子
func generateCode() string {
	const digits = "0123456789"
	b := make([]byte, 6)
	if _, err := crypto_rand.Read(b); err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	for i := range b {
		b[i] = digits[b[i]%10]
	}
	return string(b)
}

// getLangFromRequest 从请求获取语言
func getLangFromRequest(c *gin.Context, req_lang string) string {
	return utils.ResolveRequestLang(c, req_lang, "zh-CN")
}
