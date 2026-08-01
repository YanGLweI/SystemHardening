package services

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/yeung/system-hardening/backend/configs"
)

type LDAPService struct {
	config        configs.LDAPConfig
	client        *ldap.Conn
	adminConn     *ldap.Conn
	securityGroup string
}

// NewLDAPService 初始化 LDAP 服务
func NewLDAPService(cfg configs.LDAPConfig) (*LDAPService, error) {
	svc := &LDAPService{
		config:        cfg,
		securityGroup: cfg.SecurityGroupDN,
	}

	// 连接 LDAP 服务器（不执行 adminBind）
	err := svc.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %v", err)
	}

	log.Println("LDAP service initialized successfully")
	return svc, nil
}

// connect 连接到 LDAP 服务器
func (s *LDAPService) connect() error {
	var err error

	// 配置 TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: s.config.Insecure,
		ServerName:         "", // LDAPS 不需要 SNI
	}

	// 如果启用了证书校验且提供了证书路径，加载 CA 证书
	if !s.config.Insecure && s.config.CertPath != "" {
		certPool, err := loadCAPem(s.config.CertPath)
		if err != nil {
			log.Printf("Warning: Failed to load CA cert, using insecure mode: %v", err)
			tlsConfig.InsecureSkipVerify = true
		} else {
			tlsConfig.RootCAs = certPool
		}
	}

	// 使用 ldap.DialURL 连接（支持 ldaps://协议）
	s.client, err = ldap.DialURL(s.config.Server, ldap.DialWithTLSConfig(tlsConfig))
	if err != nil {
		return fmt.Errorf("failed to dial LDAP: %v", err)
	}

	return nil
}


// AuthenticateUser 验证用户密码并检查安全组权限
func (s *LDAPService) AuthenticateUser(username, password string) (bool, error) {
	cfg := &s.config
	
	// 创建新的 LDAP 连接用于认证
	conn, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(&tls.Config{
		InsecureSkipVerify: cfg.Insecure,
	}))
	if err != nil {
		return false, fmt.Errorf("连接 LDAP 失败：%v", err)
	}
	defer conn.Close()

	// 1. 使用管理员账号绑定
	err = conn.Bind(cfg.AdminUsername, cfg.AdminPassword)
	if err != nil {
		log.Printf("管理员绑定失败：%v", err)
		return false, fmt.Errorf("系统认证服务异常")
	}

	// 2. 使用管理员身份搜索用户
	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		0,
		0,
		0,
		false,
		fmt.Sprintf(cfg.UserFilter, username),
		[]string{"dn", "displayName", "cn", "memberOf"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		log.Printf("搜索用户失败：%v", err)
		return false, fmt.Errorf("用户不存在")
	}

	if len(sr.Entries) == 0 {
		log.Printf("用户%s不存在", username)
		return false, fmt.Errorf("用户不存在")
	}

	userEntry := sr.Entries[0]
	userDN := userEntry.DN
	log.Printf("找到用户 DN: %s", userDN)

	// 3. 使用用户的真实 DN 验证密码
	err = conn.Bind(userDN, password)
	if err != nil {
		log.Printf("密码验证失败：%v", err)
		return false, fmt.Errorf("用户名或密码错误")
	}

	log.Printf("用户%s密码验证成功", username)

	// 4. 检查安全组
	if !s.checkUserInSecurityGroupWithDN(userDN) {
		return false, fmt.Errorf("无权限登录")
	}

	log.Printf("用户%s认证成功且属于允许的安全组", username)
	return true, nil
}

// searchUserByLoginName 通过登录名搜索用户 DN
func (s *LDAPService) searchUserByLoginName(loginName string) (*ldap.SearchResult, error) {
	searchRequest := ldap.NewSearchRequest(
		s.config.BaseDN,
		ldap.ScopeWholeSubtree,
		0,
		0,
		0,
		false,
		fmt.Sprintf("(&(|(userPrincipalName=%s)(sAMAccountName=%s))(objectClass=user))", loginName, loginName),
		[]string{"dn", "userPrincipalName", "sAMAccountName"},
		nil,
	)
	
	result, err := s.client.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// checkUserInSecurityGroupWithDN 检查用户是否属于指定的安全组
func (s *LDAPService) checkUserInSecurityGroupWithDN(userDN string) bool {
	cfg := &s.config
	
	// 创建新的 LDAP 连接用于查询安全组
	conn, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(&tls.Config{
		InsecureSkipVerify: cfg.Insecure,
	}))
	if err != nil {
		log.Printf("连接 LDAP 失败：%v", err)
		return false
	}
	defer conn.Close()

	// 1. 使用管理员账号绑定
	err = conn.Bind(cfg.AdminUsername, cfg.AdminPassword)
	if err != nil {
		log.Printf("管理员绑定失败：%v", err)
		return false
	}

	// 2. 搜索安全组成员
	searchRequest := ldap.NewSearchRequest(
		cfg.SecurityGroupDN,
		ldap.ScopeWholeSubtree,
		0,
		0,
		0,
		false,
		fmt.Sprintf("(member=%s)", userDN),
		[]string{"dn"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		log.Printf("Error searching security group: %v", err)
		return false
	}

	// 如果找到结果，说明用户在安全组中
	return len(result.Entries) > 0
}

// getUserDNByUsername 通过用户名/邮箱获取用户的完整 DN
func (s *LDAPService) getUserDNByUsername(loginName string) string {
	// 如果是邮箱格式，直接构建 UPN 格式的 DN
	if strings.Contains(loginName, "@") {
		return fmt.Sprintf("%s@%s", loginName, s.config.DomainSuffix)
	}
	
	// 如果不是邮箱格式，尝试搜索
	sr, err := s.searchUserByLoginName(loginName)
	if err != nil || len(sr.Entries) == 0 {
		return ""
	}
	
	return sr.Entries[0].DN
}

// GetUserDetails 获取用户详细信息
func (s *LDAPService) GetUserDetails(username string) (map[string]string, error) {
	// 构建用户的 DN（使用 UPN 格式）
	userDN := fmt.Sprintf("%s@%s", username, s.config.DomainSuffix)
	
	// 搜索用户信息
	searchRequest := ldap.NewSearchRequest(
		userDN,
		ldap.ScopeBaseObject,
		0,
		0,
		0,
		false,
		"(objectClass=*)",
		[]string{"cn", "mail", "telephoneNumber", "title"},
		nil,
	)

	result, err := s.client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search user: %v", err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	userInfo := make(map[string]string)
	for _, attr := range result.Entries[0].Attributes {
		userInfo[attr.Name] = attr.Values[0]
	}

	return userInfo, nil
}

// Close 关闭 LDAP 连接
func (s *LDAPService) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.adminConn != nil {
		s.adminConn.Close()
	}
}

// loadCAPem 从文件加载 CA 证书
func loadCAPem(certPath string) (*x509.CertPool, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert file: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certData) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return certPool, nil
}
