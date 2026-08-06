package configs

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	LDAP     LDAPConfig `yaml:"ldap"`
	JWT      JWTConfig `yaml:"jwt"`
	Packages PackageConfig `yaml:"packages"`
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type LDAPConfig struct {
	Server          string `yaml:"server"`
	BaseDN          string `yaml:"base_dn"`
	DomainSuffix    string `yaml:"domain_suffix"`
	UseTLS          bool   `yaml:"use_tls"`
	Insecure        bool   `yaml:"insecure"`
	CertPath        string `yaml:"cert_path"`
	AdminUsername   string `yaml:"admin_username"`
	AdminPassword   string `yaml:"admin_password"`
	UserFilter      string `yaml:"user_filter"`
	SecurityGroupDN string `yaml:"security_group_dn"`
}

type JWTConfig struct {
	SecretKey  string `yaml:"secret_key"`
	ExpiryHour int    `yaml:"expiry_hour"`
}

type PackageConfig struct {
	LinuxPackageDir  string `yaml:"linux_package_dir"`
	WindowsPackageDir string `yaml:"windows_package_dir"`
}

func LoadConfig() *Config {
	// 尝试加载 config.yml 文件
	data, err := os.ReadFile("config.yml")
	if err != nil {
		fmt.Printf("Warning: config.yml not found: %v\n", err)
		fmt.Println("Please ensure config.yml exists in the backend directory.")
		os.Exit(1)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Error parsing config.yml: %v\n", err)
		os.Exit(1)
	}

	// 允许环境变量覆盖 YAML 配置
	config.Server.Port = getEnvOrDefault("SERVER_PORT", config.Server.Port)
	config.Database.Host = getEnvOrDefault("DB_HOST", config.Database.Host)
	config.Database.Port = getIntOrDefault("DB_PORT", config.Database.Port)
	config.Database.User = getEnvOrDefault("DB_USER", config.Database.User)
	config.Database.Password = getEnvOrDefault("DB_PASSWORD", config.Database.Password)
	config.Database.DBName = getEnvOrDefault("DB_NAME", config.Database.DBName)
	config.LDAP.Server = getEnvOrDefault("LDAP_SERVER", config.LDAP.Server)
	config.LDAP.BaseDN = getEnvOrDefault("LDAP_BASE_DN", config.LDAP.BaseDN)
	config.LDAP.DomainSuffix = getEnvOrDefault("LDAP_DOMAIN_SUFFIX", config.LDAP.DomainSuffix)
	config.LDAP.UseTLS = getBoolOrDefault("LDAP_USE_TLS", config.LDAP.UseTLS)
	config.LDAP.Insecure = getBoolOrDefault("LDAP_INSECURE", config.LDAP.Insecure)
	config.LDAP.CertPath = getEnvOrDefault("LDAP_CERT_PATH", config.LDAP.CertPath)
	config.LDAP.AdminUsername = getEnvOrDefault("LDAP_USERNAME", config.LDAP.AdminUsername)
	config.LDAP.AdminPassword = getEnvOrDefault("LDAP_PASSWORD", config.LDAP.AdminPassword)
	config.LDAP.UserFilter = getEnvOrDefault("LDAP_USER_FILTER", config.LDAP.UserFilter)
	config.LDAP.SecurityGroupDN = getEnvOrDefault("LDAP_SECURITY_GROUP_DN", config.LDAP.SecurityGroupDN)
	config.JWT.SecretKey = getEnvOrDefault("JWT_SECRET", config.JWT.SecretKey)
	config.JWT.ExpiryHour = getIntOrDefault("JWT_EXPIRY_HOUR", config.JWT.ExpiryHour)
	config.Packages.LinuxPackageDir = getEnvOrDefault("PACKAGES_LINUX_DIR", config.Packages.LinuxPackageDir)
	config.Packages.WindowsPackageDir = getEnvOrDefault("PACKAGES_WINDOWS_DIR", config.Packages.WindowsPackageDir)

	return &config
}


func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
