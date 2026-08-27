package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// TaskListResponse 客户端获取待执行任务响应
type TaskListResponse struct {
	Tasks []CheckTask `json:"tasks"`
}

// CheckTask 待执行任务
type CheckTask struct {
	TaskID     string     `json:"task_id"`
	ClientUUID string     `json:"client_uuid"`
	Status     string     `json:"status"`
	IssuedAt   *time.Time `json:"issued_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TaskResultRequest 上报任务结果请求
type TaskResultRequest struct {
	Status       string         `json:"status"`
	ErrorMessage string         `json:"error_message,omitempty"`
	ResultData   map[string]any `json:"result_data,omitempty"`
}

// FetchPendingTasks 拉取待执行任务
func FetchPendingTasks() ([]CheckTask, error) {
	token := tokenManager.GetShortToken()
	if token == "" {
		return nil, fmt.Errorf("no token available")
	}

	url := fmt.Sprintf("%s/api/client/tasks/pending", config.ServerURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("X-Client-Token", token)

	// 检查 UUID 是否已设置
	if clientUUID == "" {
		log.Printf("[TASK] WARNING: clientUUID is empty! This may cause 400 errors.")
	} else {
		log.Printf("[TASK] Sending X-Client-UUID: %s", clientUUID)
	}
	req.Header.Set("X-Client-UUID", clientUUID) // 发送 UUID

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
		return nil, fmt.Errorf("get pending tasks failed: HTTP %d, body: %s", resp.StatusCode, string(body))
	}

	var result TaskListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}

	log.Printf("[TASK] Fetched %d pending tasks", len(result.Tasks))
	return result.Tasks, nil
}

// SubmitTaskResult 上报任务执行结果
func SubmitTaskResult(taskID string, status string, errorMessage string, resultData map[string]any) error {
	token := tokenManager.GetShortToken()
	if token == "" {
		return fmt.Errorf("no token available")
	}

	url := fmt.Sprintf("%s/api/client/tasks/%s/result", config.ServerURL, taskID)

	payload := TaskResultRequest{
		Status:       status,
		ErrorMessage: errorMessage,
		ResultData:   resultData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload failed: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Token", token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("submit task result failed: HTTP %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("[TASK] Result submitted for task %s: status=%s", taskID, status)
	return nil
}

// ProcessPendingTasks 处理待执行任务
func ProcessPendingTasks() {
	tasks, err := FetchPendingTasks()
	if err != nil {
		log.Printf("[TASK] Failed to fetch pending tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	for _, task := range tasks {
		processSingleTask(task)
	}
}

// processSingleTask 处理单个任务
func processSingleTask(task CheckTask) {
	log.Printf("[TASK] Processing task %s for client %s", task.TaskID, task.ClientUUID)

	// 1. 更新任务状态为 executing
	err := SubmitTaskResult(task.TaskID, "executing", "", nil)
	if err != nil {
		log.Printf("[TASK] Failed to update task status to executing: %v", err)
		return
	}

	startTime := time.Now()

	// 2. 执行加固检查（复用现有的 runDailyCheck 逻辑）
	runDailyCheck()

	duration := time.Since(startTime)

	// 3. 上报执行结果
	err = SubmitTaskResult(
		task.TaskID,
		"completed",
		"",
		map[string]any{
			"success":          true,
			"duration_seconds": duration.Seconds(),
		},
	)
}
