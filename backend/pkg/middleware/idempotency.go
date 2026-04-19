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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

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

		tx, err := db.DB.Begin()
		if err != nil {
			utils.Fail(c, 500, "幂等校验失败")
			c.Abort()
			return
		}
		now := time.Now().Unix()
		if err := models.DeleteExpiredIdempotencyKeyTx(tx, idemKey, userID, scope, now); err != nil {
			_ = tx.Rollback()
			utils.Fail(c, 500, "幂等校验失败")
			c.Abort()
			return
		}

		item, err := models.GetActiveIdempotencyKeyTx(tx, idemKey, userID, scope, now)
		if err == nil && item != nil {
			_ = tx.Rollback()
			if item.RequestHash != requestHash {
				utils.Fail(c, 409, "幂等键已被其他请求占用")
			} else {
				utils.Fail(c, 409, "请勿重复提交")
			}
			c.Abort()
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			utils.Fail(c, 500, "幂等校验失败")
			c.Abort()
			return
		}

		expireAt := time.Unix(now, 0).Add(ttl).Unix()
		err = models.CreateIdempotencyKeyTx(tx, idemKey, userID, scope, requestHash, expireAt)
		if err != nil {
			_ = tx.Rollback()
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				utils.Fail(c, 409, "请勿重复提交")
			} else {
				utils.Fail(c, 500, "幂等校验失败")
			}
			c.Abort()
			return
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			utils.Fail(c, 500, "幂等校验失败")
			c.Abort()
			return
		}

		c.Set("idempotencyKey", idemKey)
		c.Set("idempotencyScope", scope)
		c.Set("idempotencyRequestHash", requestHash)

		if len(bodyBytes) > 0 {
			var bodyMap map[string]any
			if json.Unmarshal(bodyBytes, &bodyMap) == nil {
				c.Set("idempotencyBody", bodyMap)
			}
		}

		c.Next()
	}
}

