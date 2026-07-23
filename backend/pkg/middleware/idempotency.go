package middleware

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/utils"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// staleProcessingSeconds processing 状态超过该秒数视为僵死锁，允许同 key 重新占坑
const staleProcessingSeconds int64 = 120

// errAbortIdempotency 事务内检测到重复/冲突幂等键时中止（已写响应）
var errAbortIdempotency = errors.New("idempotency abort")

// RequireIdempotency 幂等中间件：
//  1. 请求进入时以 status=processing 占坑（防并发双花）
//  2. 业务成功（utils.Success / SuccessMsg）后标记 completed，同 key 不可再跑
//  3. 业务失败则删除占坑，同 key 可重试
func RequireIdempotency(scope string, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" {
			c.Next()
			return
		}

		idemKey := strings.TrimSpace(c.GetHeader("X-Idempotency-Key"))
		if idemKey == "" {
			utils.Fail(c, 400, "缺少 X-Idempotency-Key")
			c.Abort()
			return
		}
		if len(idemKey) > 120 {
			utils.Fail(c, 400, "X-Idempotency-Key 过长")
			c.Abort()
			return
		}

		userIDAny, exists := c.Get("userID")
		if !exists {
			utils.Fail(c, 401, "用户未登录")
			c.Abort()
			return
		}
		userID, ok := userIDAny.(uint64)
		if !ok || userID == 0 {
			utils.Fail(c, 401, "用户未登录")
			c.Abort()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			utils.Fail(c, 400, "读取请求体失败")
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		hash := sha256.Sum256(bodyBytes)
		requestHash := hex.EncodeToString(hash[:])

		now := time.Now().Unix()
		txErr := db.DB.Transaction(func(tx *gorm.DB) error {
			if err := models.DeleteExpiredIdempotencyKeyTx(tx, idemKey, userID, scope, now); err != nil {
				return err
			}
			staleBefore := now - staleProcessingSeconds
			if err := models.DeleteStaleProcessingIdempotencyKeyTx(tx, idemKey, userID, scope, staleBefore); err != nil {
				return err
			}

			item, err := models.GetActiveIdempotencyKeyTx(tx, idemKey, userID, scope, now)
			if err == nil && item != nil {
				if item.RequestHash != requestHash {
					utils.Fail(c, 409, "幂等键已被其他请求占用")
				} else if item.Status == models.IdempotencyStatusCompleted {
					utils.Fail(c, 409, "请勿重复提交")
				} else {
					utils.Fail(c, 409, "请求处理中，请稍后再试")
				}
				c.Abort()
				return errAbortIdempotency
			}
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			expireAt := time.Unix(now, 0).Add(ttl).Unix()
			if err := models.CreateIdempotencyKeyTx(tx, idemKey, userID, scope, requestHash, expireAt); err != nil {
				return err
			}
			return nil
		})
		if txErr == errAbortIdempotency {
			return
		}
		if txErr != nil {
			if db.IsDuplicateKeyError(txErr) {
				utils.Fail(c, 409, "请勿重复提交")
			} else {
				utils.Fail(c, 500, "幂等校验失败")
			}
			c.Abort()
			return
		}

		c.Set("idempotencyKey", idemKey)
		c.Set("idempotencyScope", scope)
		c.Set("idempotencyRequestHash", requestHash)
		c.Set("idempotencyUserID", userID)
		// 默认未成功；只有 utils.Success / SuccessMsg 会置 true
		c.Set(utils.CtxBizOK, false)

		if len(bodyBytes) > 0 {
			var bodyMap map[string]any
			if json.Unmarshal(bodyBytes, &bodyMap) == nil {
				c.Set("idempotencyBody", bodyMap)
			}
		}

		c.Next()

		// 业务结束后：成功→completed；失败→释放 key 允许同 key 重试
		finalizeIdempotency(c, idemKey, userID, scope)
	}
}

func finalizeIdempotency(c *gin.Context, idemKey string, userID uint64, scope string) {
	bizOK, _ := c.Get(utils.CtxBizOK)
	ok, _ := bizOK.(bool)
	if ok {
		// 业务已成功：必须尽量落到 completed，否则僵死 processing 被清理后同 key 会再跑一遍业务（如重复建充值单）
		var lastErr error
		for i := 0; i < 3; i++ {
			if err := models.MarkIdempotencyCompleted(idemKey, userID, scope); err == nil {
				return
			} else {
				lastErr = err
				time.Sleep(time.Duration(20*(i+1)) * time.Millisecond)
			}
		}
		requestHash, _ := c.Get("idempotencyRequestHash")
		hashStr, _ := requestHash.(string)
		expireAt := time.Now().Add(10 * time.Minute).Unix()
		if err := models.ForceUpsertIdempotencyCompleted(idemKey, userID, scope, hashStr, expireAt); err != nil {
			log.Printf("[Idempotency] CRITICAL mark completed failed after retry+upsert key=%s user=%d scope=%s markErr=%v upsertErr=%v",
				idemKey, userID, scope, lastErr, err)
		} else {
			log.Printf("[Idempotency] mark completed recovered via upsert key=%s user=%d scope=%s after=%v",
				idemKey, userID, scope, lastErr)
		}
		return
	}
	if err := models.DeleteIdempotencyKey(idemKey, userID, scope); err != nil {
		log.Printf("[Idempotency] release failed key=%s user=%d scope=%s err=%v", idemKey, userID, scope, err)
	}
}
