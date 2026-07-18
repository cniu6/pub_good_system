package task

import (
	"testing"
	"time"
)

func TestMatchCron_DailyMidnight(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	ok := time.Date(2026, 7, 17, 0, 0, 10, 0, loc)
	if !matchCron("0 0 * * *", ok) {
		t.Fatal("0 0 * * * 应在 00:00 命中")
	}
	bad := time.Date(2026, 7, 17, 0, 1, 0, 0, loc)
	if matchCron("0 0 * * *", bad) {
		t.Fatal("0 0 * * * 不应在 00:01 命中")
	}
}

func TestMatchCron_StepAndList(t *testing.T) {
	loc := time.UTC
	// 每 15 分钟
	at := time.Date(2026, 1, 1, 10, 30, 0, 0, loc)
	if !matchCron("*/15 * * * *", at) {
		t.Fatal("*/15 应命中 :30")
	}
	at2 := time.Date(2026, 1, 1, 10, 31, 0, 0, loc)
	if matchCron("*/15 * * * *", at2) {
		t.Fatal("*/15 不应命中 :31")
	}
	// 列表
	at3 := time.Date(2026, 1, 1, 10, 5, 0, 0, loc)
	if !matchCron("5,10,15 * * * *", at3) {
		t.Fatal("列表 5,10,15 应命中 :05")
	}
}

func TestMatchCron_Invalid(t *testing.T) {
	now := time.Now()
	if matchCron("0 0 * *", now) { // 只有 4 段
		t.Fatal("非法 cron 不应命中")
	}
	if matchCron("", now) {
		t.Fatal("空 cron 不应命中")
	}
}

func TestShouldFire_SkipWhenRunning(t *testing.T) {
	job := &JobDefinition{
		JobCode:         "demo",
		IntervalSeconds: 60,
		LastStatus:      StatusRunning,
		LastFinishedAt:  1,
	}
	if shouldFire(job, time.Now()) {
		t.Fatal("last_status=running 时不应再调度")
	}
}

func TestShouldFire_IntervalDue(t *testing.T) {
	now := time.Now()
	job := &JobDefinition{
		JobCode:         "demo_interval_" + now.Format("150405.000"),
		IntervalSeconds: 60,
		LastStatus:      StatusSuccess,
		LastFinishedAt:  now.Unix() - 120,
		Timezone:        "UTC",
	}
	if !shouldFire(job, now) {
		t.Fatal("已超过 interval 应触发")
	}
	// 刚标记过 fire，5 秒内不应再触发
	if shouldFire(job, now) {
		t.Fatal("5 秒防抖内不应再次触发")
	}
}

func TestShouldFire_IntervalNotDue(t *testing.T) {
	now := time.Now()
	job := &JobDefinition{
		JobCode:         "demo_not_due",
		IntervalSeconds: 3600,
		LastStatus:      StatusSuccess,
		LastFinishedAt:  now.Unix() - 10,
		Timezone:        "UTC",
	}
	if shouldFire(job, now) {
		t.Fatal("未到 interval 不应触发")
	}
}

func TestHandlersCoverPresets(t *testing.T) {
	for _, p := range DefaultPresets() {
		if _, ok := GetHandler(p.HandlerKey); !ok {
			t.Fatalf("预设 %s 的 handler_key=%s 未在 handlers 表中", p.JobCode, p.HandlerKey)
		}
		if p.HandlerKey != p.JobCode {
			t.Fatalf("约定 job_code 与 handler_key 一致: %s vs %s", p.JobCode, p.HandlerKey)
		}
	}
}

func TestRunIDNearLimitBelowMaxInt64(t *testing.T) {
	const maxInt64 = uint64(1<<63 - 1)
	if runIDNearLimit >= maxInt64 {
		t.Fatalf("水位 %d 应小于 MaxInt64 %d", runIDNearLimit, maxInt64)
	}
	if maxInt64-runIDNearLimit != 1_000_000 {
		t.Fatalf("应预留 100 万余量，实际差 %d", maxInt64-runIDNearLimit)
	}
}

func TestMarshalDetail(t *testing.T) {
	if marshalDetail(nil) != "" {
		t.Fatal("nil detail 应为空串")
	}
	s := marshalDetail(map[string]interface{}{"n": 1})
	if s == "" || s[0] != '{' {
		t.Fatalf("unexpected detail: %q", s)
	}
}

func TestNewRunUID(t *testing.T) {
	u := newRunUID()
	if len(u) != 36 {
		t.Fatalf("run_uid 长度应为 36，实际 %d: %q", len(u), u)
	}
	if u[8] != '-' || u[13] != '-' || u[18] != '-' || u[23] != '-' {
		t.Fatalf("run_uid 格式不对: %q", u)
	}
}

func TestStuckLimitSec(t *testing.T) {
	if stuckLimitSec(0, 0) != 600 {
		t.Fatal("默认应为 600")
	}
	if stuckLimitSec(120, 300) != 300 {
		t.Fatal("应取 timeout 与全局较大值")
	}
	if stuckLimitSec(900, 300) != 900 {
		t.Fatal("全局更大时应取全局")
	}
	if stuckLimitSec(10, 10) != 60 {
		t.Fatal("至少 60 秒")
	}
}
