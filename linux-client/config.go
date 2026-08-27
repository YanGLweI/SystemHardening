package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 客户端配置
type Config struct {
	ServerURL   string `yaml:"server_url"`
	LocalDBPath string `yaml:"local_db_path"`
	DeviceName  string `yaml:"device_name"`
	IPAddress   string `yaml:"ip_address"`
	ScriptPath  string `yaml:"script_path,omitempty"`
}

// LoadConfig 从 YAML 文件加载配置
func LoadConfig(filePath string) (Config, error) {
	config := Config{
		ScriptPath: "/opt/linux-hardening-client/scripts/System_Check-1.2.sh",
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}

	return config, nil
}
