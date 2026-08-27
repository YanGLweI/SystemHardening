package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/middleware"
	"github.com/yeung/system-hardening/backend/models"
	"gorm.io/gorm"
)

type TaskController struct{}

// NewTaskController create new TaskController instance
func NewTaskController() *TaskController {
	return &TaskController{}
}

// TriggerCheckTaskRequest trigger check task request
type TriggerCheckTaskRequest struct {
	ClientUUID string `json:"client_uuid" binding:"required"`
}

// SubmitTaskResultRequest submit task result request
type SubmitTaskResultRequest struct {
	Status       string         `json:"status" binding:"required,oneof=executing completed failed"`
	ErrorMessage string         `json:"error_message"`
	ResultData   models.JSONMap `json:"result_data"`
}

// TriggerCheckTask 触发立即检查任务
func (tc *TaskController) TriggerCheckTask(c *gin.Context) {
	var req TriggerCheckTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	clientUUID := req.ClientUUID

	// 1. 检查并发限制（同一客户端只能有一个 pending/executing 任务）
	var existingTask models.CheckTask
	err := database.DB.Where("client_uuid = ? AND status IN ('pending', 'executing')", clientUUID).First(&existingTask).Error

	if err == nil && existingTask.ID > 0 {
		c.JSON(409, gin.H{
			"error":   "该客户端已有正在执行的任务",
			"task_id": existingTask.TaskID,
		})
		return
	}

	// 2. 生成任务 ID
	taskID := generateTaskID()

	// 3. 创建任务记录
	task := models.CheckTask{
		TaskID:      taskID,
		ClientUUID:  clientUUID,
		TriggeredBy: getAdminUsername(c),
		Status:      "pending",
		IssuedAt:    &time.Time{},
		CreatedAt:   time.Now(),
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(200, gin.H{
		"message": "任务已创建",
		"task_id": task.TaskID,
		"status":  task.Status,
	})
}

// GetTaskStatus 查询任务状态
func (tc *TaskController) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	var task models.CheckTask
	if err := database.DB.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(200, task)
}

// GetPendingTasksForClient 客户端拉取待执行任务
func (tc *TaskController) GetPendingTasksForClient(c *gin.Context) {
	clientUUID := c.GetHeader("X-Client-UUID")
	if clientUUID == "" {
		c.JSON(400, gin.H{"error": "Missing client UUID in header"})
		return
	}

	// 可选：同时支持从 token 中解析 client_uuid
	tokenStr := c.GetHeader("X-Client-Token")
	if tokenStr != "" {
		var token models.ClientToken
		if err := database.DB.Where("short_token = ? AND expires_at > ?", tokenStr, time.Now()).First(&token).Error; err == nil {
			clientUUID = token.ClientUUID
		}
	}

	var tasks []models.CheckTask
	err := database.DB.Where("client_uuid = ? AND status = ?", clientUUID, models.StatusPending).Find(&tasks).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	// 更新任务状态为 sent
	now := time.Now()
	for i := range tasks {
		tasks[i].Status = "sent"
		tasks[i].IssuedAt = &now
		database.DB.Save(&tasks[i])
	}

	c.JSON(200, gin.H{"tasks": tasks})
}

// SubmitTaskResult 客户端上报执行结果
func (tc *TaskController) SubmitTaskResult(c *gin.Context) {
	taskID := c.Param("id")

	var req SubmitTaskResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	var task models.CheckTask
	if err := database.DB.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      req.Status,
		"retry_count": task.RetryCount + 1,
	}

	// 根据状态更新不同字段
	if req.Status == "executing" {
		// 任务开始执行
		updates["started_at"] = &now
	} else if req.Status == "completed" || req.Status == "failed" {
		// 任务完成或失败，设置完成时间
		updates["completed_at"] = now

		if req.Status == "failed" {
			updates["error_message"] = req.ErrorMessage
		} else if req.Status == "completed" {
			// 保存结果数据（转换为 JSON 字符串）
			if req.ResultData != nil {
				jsonData, err := json.Marshal(req.ResultData)
				if err != nil {
					log.Printf("Failed to marshal result data: %v", err)
				} else {
					updates["result_summary"] = string(jsonData)
				}
			}
		}
	}

	if err := database.DB.Model(&task).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(200, gin.H{"message": "Result received"})
}

// DeleteTask 删除任务（用于卡死任务重试：删除当前任务后可重新创建）
func (tc *TaskController) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")

	result := database.DB.Where("task_id = ?", taskID).Delete(&models.CheckTask{})
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Failed to delete task"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	}

	log.Printf("Task deleted: %s (by %s)", taskID, getAdminUsername(c))
	c.JSON(200, gin.H{"message": "Task deleted"})
}

// Helper functions

// generateTaskID 生成唯一任务 ID
func generateTaskID() string {
	timestamp := time.Now().Format("20060102150405") // 精确到秒
	randomSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	return fmt.Sprintf("chk_%s%s", timestamp, randomSuffix)
}

// getAdminUsername 获取当前登录管理员用户名
func getAdminUsername(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return "system"
	}

	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	username, err := middleware.ExtractUsername(tokenStr, globalConfig.JWT)
	if err != nil {
		return "system"
	}

	return username
}

// GetClientLatestTask 获取客户端的最新任务（如果有）
func (tc *TaskController) GetClientLatestTask(c *gin.Context) {
	clientUUID := c.Param("client_uuid")

	if clientUUID == "" {
		c.JSON(400, gin.H{"error": "Missing required parameter: client_uuid"})
		return
	}

	var task models.CheckTask
	if err := database.DB.Where("client_uuid = ? AND status IN ('pending', 'sent', 'executing')", clientUUID).
		Order("created_at DESC").
		First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(204, gin.H{}) // 没有进行中任务
		} else {
			c.JSON(500, gin.H{"error": "Failed to query task"})
		}
		return
	}

	c.JSON(200, gin.H{
		"task_id":        task.TaskID,
		"status":         task.Status,
		"error_message":  task.ErrorMessage,
		"result_summary": task.ResultSummary,
		"retry_count":    task.RetryCount,
	})
}
