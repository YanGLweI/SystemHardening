package main

import (
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// CollectHardwareUUID 采集 SMBIOS UUID，优先级：
// 1. /sys/class/dmi/id/product_uuid（需 root，systemd 服务默认满足）
// 2. dmidecode -s system-uuid
// 校验标准 UUID 格式并拒绝全 0/全 F/占位值；均不可用返回空串（降级不报错）
func CollectHardwareUUID() string {
	// 优先读取 sysfs
	if data, err := os.ReadFile("/sys/class/dmi/id/product_uuid"); err == nil {
		uuid := strings.TrimSpace(string(data))
		if IsValidHardwareUUID(uuid) {
			return strings.ToUpper(uuid)
		}
		log.Printf("⚠️ product_uuid 无效（%s），尝试 dmidecode", uuid)
	}

	// 回退：dmidecode
	if out, err := exec.Command("dmidecode", "-s", "system-uuid").Output(); err == nil {
		uuid := strings.TrimSpace(string(out))
		if IsValidHardwareUUID(uuid) {
			return strings.ToUpper(uuid)
		}
		log.Printf("⚠️ dmidecode system-uuid 无效（%s）", uuid)
	}

	// 无 DMI 环境（容器等）：返回空串，心跳/注册不携带该字段
	return ""
}

var hardwareUUIDRegex = regexp.MustCompile(`(?i)^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$`)

// IsValidHardwareUUID 校验标准硬件 UUID 格式（8-4-4-4-12 十六进制，大小写不敏感），
// 并拒绝全 0 / 全 F 等无效 SMBIOS 值（与 Windows 客户端 is_valid_hardware_uuid 同源）
func IsValidHardwareUUID(s string) bool {
	t := strings.TrimSpace(s)
	if !hardwareUUIDRegex.MatchString(t) {
		return false
	}
	hexOnly := strings.ReplaceAll(t, "-", "")
	allZero := true
	allF := true
	for _, c := range strings.ToLower(hexOnly) {
		if c != '0' {
			allZero = false
		}
		if c != 'f' {
			allF = false
		}
	}
	return !allZero && !allF
}

// IsPlaceholderSerial BIOS 序列号占位值黑名单（与 Windows is_placeholder_serial / 后端 isPlaceholderHardwareUUID 同源）
func IsPlaceholderSerial(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "default string", "to be filled by o.e.m.", "to be filled by oem",
		"not specified", "none", "null", "n/a", "unknown", "o.e.m.",
		"system serial number", "chassis serial number", "baseboard serial number",
		"serial", "0", "0000000000", "123456789", "0123456789":
		return true
	}
	return false
}

// GetHostname 实时主机名：os.Hostname()，失败返回空串（服务端空值不更新）
func GetHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

// GetLocalIP 主 IP：net.Dial("udp","8.8.8.8:80") 取本地地址（不发包）；
// 失败回退遍历网卡取首个非回环 IPv4；均失败返回空串
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			return addr.IP.String()
		}
	}

	// 回退：遍历网卡
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 != nil {
				return ip4.String()
			}
		}
	}
	return ""
}
