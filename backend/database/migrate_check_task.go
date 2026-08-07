package database

import (
	"github.com/yeung/system-hardening/backend/models"
)

// AutoMigrateCheckTasks auto migrate check_tasks table
func AutoMigrateCheckTasks() error {
	return DB.AutoMigrate(&models.CheckTask{})
}
