package task

import (
	"context"
	"fmt"
	"log"

	"fst/backend/app/models"
)

// handlers 内置任务：handler_key → 函数。新任务加一行即可。
var handlers = map[string]JobHandler{
	HandlerPruneAutoJobRuns:          handlePruneAutoJobRuns,
	HandlerMarkStuckAutoJobs:         handleMarkStuckAutoJobs,
	HandlerCleanupExpiredIdempotency: handleCleanupExpiredIdempotency,
	HandlerCleanupSessionsCodes:      handleCleanupSessionsCodes,
	HandlerCleanupExpiredOrders:      handleCleanupExpiredOrders,
}

func GetHandler(key string) (JobHandler, bool) {
	h, ok := handlers[key]
	return h, ok
}

func ListHandlerKeys() []string {
	out := make([]string, 0, len(handlers))
	for k := range handlers {
		out = append(out, k)
	}
	return out
}

// errIfCanceled 超时/取消时尽快退出多步骤 handler（单条 SQL 仍可能跑完，但不会继续下一步）
func errIfCanceled(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func handlePruneAutoJobRuns(ctx context.Context, job *JobDefinition) (*HandlerResult, error) {
	_ = job
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	cfg := LoadGlobalConfig()
	deleted, err := PruneSuccessRuns(cfg.RunMaxCount)
	if err != nil {
		return nil, err
	}
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	var warnings []string

	repaired, rerr := RepairBadRunUIDs()
	if rerr != nil {
		log.Printf("[AutoJob] 修复 run_uid 失败: %v", rerr)
		warnings = append(warnings, fmt.Sprintf("修复 run_uid 失败: %v", rerr))
	}
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	renumbered, kept, rerr := MaybeRenumberRunIDsIfNearLimit()
	if rerr != nil {
		log.Printf("[AutoJob] id 重编号检测失败: %v", rerr)
		warnings = append(warnings, fmt.Sprintf("id 重编号检测失败: %v", rerr))
	}
	n, _ := CountRuns()
	msg := fmt.Sprintf("修剪删除 %d 条，当前 %d/%d", deleted, n, cfg.RunMaxCount)
	detail := map[string]interface{}{"deleted": deleted, "row_count": n, "max_count": cfg.RunMaxCount}
	if repaired > 0 {
		msg += fmt.Sprintf("；修复 run_uid %d 条", repaired)
		detail["run_uid_repaired"] = repaired
	}
	if renumbered {
		msg += fmt.Sprintf("；id 已重编号为 1..%d", kept)
		detail["id_renumbered"] = true
		detail["id_renumbered_count"] = kept
	}
	// 子步骤失败之前只打日志，主 handler 仍返回 success，运维在任务运行记录里看不出「部分失败」。
	// 这两步都是维护性兜底操作（非致命），失败了不应该让整个任务标记失败，但要能在 detail 里看到。
	if len(warnings) > 0 {
		msg += fmt.Sprintf("；有 %d 项子步骤失败，见 warnings", len(warnings))
		detail["warnings"] = warnings
	}
	return &HandlerResult{
		Message: msg,
		Detail:  detail,
		Quiet:   deleted == 0 && repaired == 0 && !renumbered && len(warnings) == 0,
	}, nil
}

func handleMarkStuckAutoJobs(ctx context.Context, job *JobDefinition) (*HandlerResult, error) {
	_ = job
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	cfg := LoadGlobalConfig()
	n, err := MarkStuckRuns(cfg.StuckAfterSec)
	if err != nil {
		return nil, err
	}
	return &HandlerResult{
		Message: fmt.Sprintf("标记超时 %d 条", n),
		Detail:  map[string]interface{}{"marked": n},
		Quiet:   n == 0,
	}, nil
}

func handleCleanupExpiredIdempotency(ctx context.Context, job *JobDefinition) (*HandlerResult, error) {
	_ = job
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	aff, err := models.CleanupExpiredIdempotencyKeys()
	if err != nil {
		return nil, err
	}
	return &HandlerResult{
		Message: fmt.Sprintf("清理幂等键 %d", aff),
		Detail:  map[string]interface{}{"affected": aff},
		Quiet:   aff == 0,
	}, nil
}

func handleCleanupSessionsCodes(ctx context.Context, job *JobDefinition) (*HandlerResult, error) {
	_ = job
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	soft, err := models.SoftDeleteExpiredCodes()
	if err != nil {
		return nil, err
	}
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	old, err := models.CleanupOldVerificationCodes()
	if err != nil {
		return nil, err
	}
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	sess, err := models.CleanupExpiredSessions()
	if err != nil {
		return nil, err
	}
	total := soft + old + sess
	return &HandlerResult{
		Message: fmt.Sprintf("验证码/会话清理：软删%d 硬删%d 会话%d", soft, old, sess),
		Detail: map[string]interface{}{
			"codes_soft_deleted": soft,
			"codes_hard_deleted": old,
			"sessions_deleted":   sess,
		},
		Quiet: total == 0,
	}, nil
}

func handleCleanupExpiredOrders(ctx context.Context, job *JobDefinition) (*HandlerResult, error) {
	_ = job
	if err := errIfCanceled(ctx); err != nil {
		return nil, err
	}
	aff, err := models.CancelExpiredOrders()
	if err != nil {
		return nil, err
	}
	return &HandlerResult{
		Message: fmt.Sprintf("取消过期支付单 %d", aff),
		Detail:  map[string]interface{}{"affected": aff},
		Quiet:   aff == 0,
	}, nil
}
