package controllers

import (
	"testing"

	"github.com/yeung/system-hardening/backend/models"
)

// TestIsDegradedWindowsUpload 降级上传保护判定单元测试
// 对应 v2.2.4 修复 4：更新重启后环境半就绪导致三组策略全空的降级上传，
// 必须跳过覆盖以保留服务端历史健康数据；同时不得影响域控豁免语义。
func TestIsDegradedWindowsUpload(t *testing.T) {
	tests := []struct {
		name     string
		incoming models.WindowsSystemCheck
		existing models.WindowsSystemCheck
		want     bool
	}{
		{
			name:     "三组全空且已有记录有值 → 触发保护",
			incoming: models.WindowsSystemCheck{},
			existing: models.WindowsSystemCheck{
				MinimumPasswordLength: "8",
				AuditSystemEvents:     "Success",
				ScreenSaverActive:     "1",
			},
			want: true,
		},
		{
			name:     "三组全空且已有记录仅屏保有值 → 触发保护",
			incoming: models.WindowsSystemCheck{},
			existing: models.WindowsSystemCheck{ScreenSaverActive: "1"},
			want:     true,
		},
		{
			name:     "三组全空且已有记录也全空 → 不触发保护（正常覆盖空值）",
			incoming: models.WindowsSystemCheck{},
			existing: models.WindowsSystemCheck{},
			want:     false,
		},
		{
			name: "豁免场景：仅屏保为空、密码/审计有值 → 不触发保护（正常覆盖）",
			incoming: models.WindowsSystemCheck{
				MinimumPasswordLength: "8",
				AuditSystemEvents:     "Success",
				ScreenSaverActive:     "",
			},
			existing: models.WindowsSystemCheck{
				MinimumPasswordLength: "8",
				AuditSystemEvents:     "Success",
				ScreenSaverActive:     "1",
			},
			want: false,
		},
		{
			name: "正常完整上传 → 不触发保护",
			incoming: models.WindowsSystemCheck{
				MinimumPasswordLength: "8",
				AuditSystemEvents:     "Success",
				ScreenSaverActive:     "1",
			},
			existing: models.WindowsSystemCheck{
				MinimumPasswordLength: "6",
				AuditSystemEvents:     "Failure",
				ScreenSaverActive:     "0",
			},
			want: false,
		},
		{
			name: "仅密码为空、审计/屏保有值 → 不触发保护",
			incoming: models.WindowsSystemCheck{
				AuditSystemEvents: "Success",
				ScreenSaverActive: "1",
			},
			existing: models.WindowsSystemCheck{MinimumPasswordLength: "8"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDegradedWindowsUpload(&tt.incoming, &tt.existing); got != tt.want {
				t.Errorf("isDegradedWindowsUpload() = %v, want %v", got, tt.want)
			}
		})
	}
}
