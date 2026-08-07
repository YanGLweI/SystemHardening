package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	// 客户端安装目录
	clientDir = "/opt/linux-hardening-client"
	// 配置文件路径
	configPath = "/opt/linux-hardening-client/config.yaml"
	// 日志目录
	logDir = "/opt/linux-hardening-client/logs"
)

// InstallUpdate 安装更新包（tempPath 为下载器返回的临时 zip 路径）
func InstallUpdate(tempPath string) error {
	log.Println("[UPDATER] Starting update installation...")
	log.Printf("[UPDATER] Found update package: %s", tempPath)

	// 1. 【关键】备份配置文件
	log.Println("[UPDATER] Backing up config file...")
	configBackup := fmt.Sprintf("%s/.config.backup.%d.yaml", logDir, time.Now().Unix())
	if err := copyFile(configPath, configBackup); err != nil {
		return fmt.Errorf("backup config failed: %v", err)
	}
	log.Printf("[UPDATER] Config backed up to: %s", configBackup)

	// 2. 解压到临时目录
	extractDir := fmt.Sprintf("/tmp/systemhardening-extract-%d", time.Now().Unix())
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("create extract dir failed: %v", err)
	}
	defer os.RemoveAll(extractDir)

	log.Println("[UPDATER] Extracting update package...")
	if err := execCommand("unzip", "-o", tempPath, "-d", extractDir); err != nil {
		return fmt.Errorf("extract zip failed: %v", err)
	}

	// 3. 展平版本子目录（zip 内可能包含 linux-hardening-client_vX.Y.Z/ 子目录）
	srcDir, err := flattenExtractDir(extractDir)
	if err != nil {
		return err
	}

	// 4. 校验新二进制存在
	newBinary := filepath.Join(srcDir, "linux-hardening-client")
	if _, err := os.Stat(newBinary); err != nil {
		return fmt.Errorf("new binary not found in package: %v", err)
	}

	// 5. 备份并替换二进制（保持现有安装布局 bin/ + scripts/）
	log.Println("[UPDATER] Backing up current binary...")
	binPath := filepath.Join(clientDir, "bin", "linux-hardening-client")
	binBackup := fmt.Sprintf("%s.backup.%d", binPath, time.Now().Unix())
	if _, err := os.Stat(binPath); err == nil {
		if err := execCommand("mv", binPath, binBackup); err != nil {
			return fmt.Errorf("backup current binary failed: %v", err)
		}
		defer os.Remove(binBackup)
	}

	if err := copyFile(newBinary, binPath); err != nil {
		// 回滚二进制
		if binBackup != "" {
			execCommand("mv", binBackup, binPath)
		}
		return fmt.Errorf("replace binary failed: %v", err)
	}
	if err := execCommand("chmod", "+x", binPath); err != nil {
		log.Printf("Warning: chmod binary failed: %v", err)
	}
	log.Println("[UPDATER] ✅ Binary updated successfully")

	// 6. 更新检查脚本（包内为平铺的 *.sh，安装布局位于 scripts/）
	scriptsDir := filepath.Join(clientDir, "scripts")
	os.MkdirAll(scriptsDir, 0755)
	shFiles, _ := filepath.Glob(filepath.Join(srcDir, "*.sh"))
	for _, sh := range shFiles {
		name := filepath.Base(sh)
		// uninstall.sh 保留在安装根目录，与安装脚本布局一致
		dst := filepath.Join(scriptsDir, name)
		if name == "uninstall.sh" {
			dst = filepath.Join(clientDir, name)
		}
		if err := copyFile(sh, dst); err != nil {
			log.Printf("Warning: update script %s failed: %v", name, err)
			continue
		}
		execCommand("chmod", "+x", dst)
	}
	log.Println("[UPDATER] ✅ Scripts updated successfully")

	// 7. 【关键】恢复配置文件
	log.Println("[UPDATER] Restoring config file...")
	if err := copyFile(configBackup, configPath); err != nil {
		// 回滚二进制
		if binBackup != "" {
			execCommand("mv", binBackup, binPath)
		}
		return fmt.Errorf("restore config failed: %v", err)
	}
	os.Remove(configBackup)
	log.Println("[UPDATER] ✅ Config restored successfully")

	// 8. 清理下载的临时文件和解压目录（重启后进程会被 systemd 终止，defer 不会执行，必须提前清理）
	os.Remove(tempPath)
	os.RemoveAll(extractDir)

	// 9. 重启服务
	log.Println("[UPDATER] Restarting systemd service...")
	if err := execCommand("systemctl", "daemon-reload"); err != nil {
		log.Printf("Warning: systemctl daemon-reload failed: %v", err)
	}

	if err := execCommand("systemctl", "restart", "linux-hardening-client"); err != nil {
		// 重启失败 - 回滚二进制
		log.Printf("[UPDATER] ❌ Service restart failed, rolling back...")
		if binBackup != "" {
			execCommand("mv", binBackup, binPath)
			execCommand("systemctl", "restart", "linux-hardening-client")
		}
		return fmt.Errorf("service restart failed, rolled back to original version")
	}

	log.Println("[UPDATER] ✅ Service restarted successfully")
	log.Printf("[UPDATER] 🎉 Update completed! Current version: %s", version)

	return nil
}

// flattenExtractDir 若解压结果仅包含单个版本子目录，则返回该子目录；否则返回原目录
func flattenExtractDir(extractDir string) (string, error) {
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return "", fmt.Errorf("read extract dir failed: %v", err)
	}
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(extractDir, entries[0].Name())
		log.Printf("[UPDATER] Detected version subdirectory: %s", entries[0].Name())
		return subDir, nil
	}
	return extractDir, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// execCommand 执行命令
func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
