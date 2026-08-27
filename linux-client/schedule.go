package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// CheckScheduleResponse 服务端下发的加固检查计划
type CheckScheduleResponse struct {
	ScheduleType string `json:"schedule_type"` // daily/weekly/monthly
	CheckTime    string `json:"check_time"`    // HH:mm（半小时粒度）
	Weekday      int    `json:"weekday"`       // 1-7（周一到周日，weekly）
	DayOfMonth   int    `json:"day_of_month"`  // 1-31（monthly）
	UpdatedAt    string `json:"updated_at"`    // 计划最后更新时间（用于变更检测）
}

// ScheduleState 当前生效的计划状态（调度器与计划拉取协程共享）
type ScheduleState struct {
	mu            sync.Mutex
	applied       CheckScheduleResponse // 已应用的计划
	appliedUpdate string                // 已应用的计划 updated_at
	jitter        time.Duration         // 随机抖动（进程生命周期内固定）
	nextCheckTime time.Time             // 下次检查时刻
}

var scheduleState = &ScheduleState{}

// FetchCheckSchedule 从服务端获取加固检查计划
func FetchCheckSchedule() (*CheckScheduleResponse, error) {
	token := tokenManager.GetShortToken()
	if token == "" {
		return nil, fmt.Errorf("no token available")
	}

	url := fmt.Sprintf("%s/api/client/check-schedule", config.ServerURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("X-Client-Token", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get check schedule failed: HTTP %d, body: %s", resp.StatusCode, string(body))
	}

	var result CheckScheduleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}
	return &result, nil
}

// scheduleLoop 定期拉取检查计划（每 5 分钟），计划变化时立即重算下次检查时刻
func scheduleLoop() {
	log.Println("Starting check schedule loop (every 5 minutes)...")

	applyScheduleIfChanged()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		applyScheduleIfChanged()
	}
}

// applyScheduleIfChanged 拉取计划并比对 updated_at，仅在变化时重新生效
func applyScheduleIfChanged() {
	schedule, err := FetchCheckSchedule()
	if err != nil {
		log.Printf("[SCHEDULE] Fetch check schedule failed: %v", err)
		return
	}

	scheduleState.mu.Lock()
	defer scheduleState.mu.Unlock()

	if schedule.UpdatedAt == scheduleState.appliedUpdate {
		return // 计划未变化，无需重新生效
	}

	// 计划变化（或首次应用）：生成一次性随机抖动（0~15 分钟）并重算下次检查时刻
	scheduleState.jitter = time.Duration(rand.Intn(16)) * time.Minute
	scheduleState.applied = *schedule
	scheduleState.appliedUpdate = schedule.UpdatedAt
	scheduleState.nextCheckTime = calcNextCheck(*schedule, time.Now()).Add(scheduleState.jitter)

	log.Printf("[SCHEDULE] ✅ 应用新检查计划：type=%s time=%s weekday=%d day_of_month=%d，下次检查：%s（抖动 %s）",
		schedule.ScheduleType, schedule.CheckTime, schedule.Weekday, schedule.DayOfMonth,
		scheduleState.nextCheckTime.Format("2006-01-02 15:04:05"), scheduleState.jitter)
}

// getNextCheckTime 获取下次检查时刻（零值表示尚未应用计划）
func getNextCheckTime() time.Time {
	scheduleState.mu.Lock()
	defer scheduleState.mu.Unlock()
	return scheduleState.nextCheckTime
}

// recomputeNextCheck 检查执行完成后，基于已应用的计划重算下次检查时刻
func recomputeNextCheck() {
	scheduleState.mu.Lock()
	defer scheduleState.mu.Unlock()

	applied := scheduleState.applied
	if applied.ScheduleType == "" {
		// 从未成功拉取计划时使用兜底默认计划（每天 02:00）
		applied = CheckScheduleResponse{ScheduleType: "daily", CheckTime: "02:00", Weekday: 1, DayOfMonth: 1}
	}
	scheduleState.nextCheckTime = calcNextCheck(applied, time.Now()).Add(scheduleState.jitter)
	log.Printf("[SCHEDULE] 下次检查时刻已重算：%s", scheduleState.nextCheckTime.Format("2006-01-02 15:04:05"))
}

// calcNextCheck 基于计划计算下一个计划时刻（本地时区）
// monthly 遇当月无该日期时取当月最后一天
func calcNextCheck(s CheckScheduleResponse, from time.Time) time.Time {
	hour, minute := parseCheckTime(s.CheckTime)

	switch s.ScheduleType {
	case "weekly":
		// 服务端 weekday 1-7 对应周一到周日；Go Weekday() 周日=0
		target := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, from.Location())
		delta := (s.Weekday - weekdayToISO(from.Weekday()) + 7) % 7
		target = target.AddDate(0, 0, delta)
		if !target.After(from) {
			target = target.AddDate(0, 0, 7)
		}
		return target
	case "monthly":
		// 向后最多找 13 个月，取下一个出现的日期
		for i := 0; i < 13; i++ {
			y, m, _ := from.AddDate(0, i, 0).Date()
			day := s.DayOfMonth
			if last := daysInMonth(y, m); day > last {
				day = last
			}
			target := time.Date(y, m, day, hour, minute, 0, 0, from.Location())
			if target.After(from) {
				return target
			}
		}
	}

	// daily（未知类型也按 daily 兜底）
	target := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, from.Location())
	if !target.After(from) {
		target = target.AddDate(0, 0, 1)
	}
	return target
}

// parseCheckTime 解析 HH:mm，失败时兜底 02:00
func parseCheckTime(checkTime string) (int, int) {
	var h, m int
	if _, err := fmt.Sscanf(checkTime, "%d:%d", &h, &m); err != nil {
		return 2, 0
	}
	return h, m
}

// weekdayToISO 将 Go 的 Weekday 转换为 ISO（周一=1 ... 周日=7）
func weekdayToISO(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}

// daysInMonth 返回指定年月的天数
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
