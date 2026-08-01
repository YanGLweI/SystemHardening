# LDAP 登录问题排查与解决

## 🔍 问题根源分析

### 1. **初始错误：直接 Bind 失败**
- ❌ 错误做法：`l.client.Bind("ylw@hot.local", password)`
- ✅ 正确做法：先搜索获取用户的完整 DN（如 `CN=杨龙威，OU=IT 部,...`），然后 Bind

### 2. **第二个错误：安全组查询需要管理员权限**
- ❌ 错误做法：使用未绑定的连接搜索安全组成员
- ✅ 正确做法：每次检查安全组时创建新连接并先绑定管理员账号

### 3. **第三个错误：配置 YAML tag 缺失**
- ❌ 错误做法：LDAPConfig 结构体缺少 `yaml:"..."` tag
- ✅ 正确做法：为所有字段添加 yaml tag

### 4. **第四个错误：中文字符 DN 编码**
- ❌ 错误做法：`CN=IT 部，OU=IT 部，...`
- ✅ 正确做法：`CN=IT\u90e8,OU=IT\u90e8,...` (UTF-8 转义)

---

## 🛠️ 解决方案

### 参考 IT 项目的认证流程

#### Step 1: 管理员绑定 LDAP
```go
// 使用服务账号绑定
err = l.Bind(cfg.Username, cfg.Password)
if err != nil {
    return "", "", false, nil, fmt.Errorf("LDAP 绑定失败：%v", err)
}
```

#### Step 2: 搜索用户获取完整 DN
```go
searchRequest := ldap.NewSearchRequest(
    cfg.BaseDN,
    ldap.ScopeWholeSubtree,
    0, 0, 0, false,
    fmt.Sprintf(cfg.UserFilter, username),
    []string{"dn", "displayName", "cn", "memberOf"},
    nil,
)
sr, err := l.Search(searchRequest)
userDN := sr.Entries[0].DN
```

#### Step 3: 使用用户 DN 验证密码
```go
err = l.Bind(userDN, password)
if err != nil {
    return "", "", false, nil, fmt.Errorf("密码错误")
}
```

#### Step 4: 检查安全组
```go
// 再次创建新的连接并使用管理员身份查询
err = l.Bind(cfg.Username, cfg.Password)
searchRequest := ldap.NewSearchRequest(
    cfg.BaseDN,
    ldap.ScopeWholeSubtree,
    0, 0, 0, false,
    fmt.Sprintf("(&(objectClass=group)(member=%s))", userDN),
    []string{"dn"},
    nil,
)
```

---

## 📝 关键代码修改

### 1. configs/config.go - 添加 YAML Tag
```go
type LDAPConfig struct {
    Server          string `yaml:"server"`
    BaseDN          string `yaml:"base_dn"`
    AdminUsername   string `yaml:"admin_username"`
    AdminPassword   string `yaml:"admin_password"`
    // ...其他字段
}
```

### 2. services/ldap_service.go - AuthenticateUser 重构
```go
func (s *LDAPService) AuthenticateUser(username, password string) (bool, error) {
    cfg := &s.config
    
    // 创建新的 LDAP 连接用于认证
    conn, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(&tls.Config{
        InsecureSkipVerify: cfg.Insecure,
    }))
    
    // 1. 使用管理员账号绑定
    err = conn.Bind(cfg.AdminUsername, cfg.AdminPassword)
    
    // 2. 使用管理员身份搜索用户，获取真实 DN
    searchRequest := ldap.NewSearchRequest(...)
    sr, err := conn.Search(searchRequest)
    userDN := sr.Entries[0].DN
    
    // 3. 使用用户的真实 DN 验证密码
    err = conn.Bind(userDN, password)
    
    // 4. 检查安全组
    if !s.checkUserInSecurityGroupWithDN(userDN) {
        return false, fmt.Errorf("无权限登录")
    }
    
    return true, nil
}
```

### 3. services/ldap_service.go - checkUserInSecurityGroupWithDN
```go
func (s *LDAPService) checkUserInSecurityGroupWithDN(userDN string) bool {
    cfg := &s.config
    
    // 创建新的 LDAP 连接用于查询安全组
    conn, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(&tls.Config{
        InsecureSkipVerify: cfg.Insecure,
    }))
    
    // 必须先绑定管理员才能查询
    err = conn.Bind(cfg.AdminUsername, cfg.AdminPassword)
    
    // 搜索安全组成员
    searchRequest := ldap.NewSearchRequest(
        cfg.SecurityGroupDN,  // 注意：这里使用安全组的 DN
        ldap.ScopeWholeSubtree,
        0, 0, 0, false,
        fmt.Sprintf("(member=%s)", userDN),
        []string{"dn"},
        nil,
    )
    
    result, err := conn.Search(searchRequest)
    return len(result.Entries) > 0
}
```

### 4. config.yml - 正确的 DN 格式
```yaml
ldap:
  server: "ldaps://10.60.254.252:636"
  base_dn: "dc=hot,dc=local"
  admin_username: "ylw@hot.local"
  admin_password: "!Qw2!Qw2!Qw2!Qw2"
  # UTF-8 转义的中文名
  security_group_dn: "CN=IT\u90e8,OU=IT\u90e8,OU=HOT,DC=hot,DC=local"
```

---

## ✅ 验证结果

### 后端日志输出
```
2026/08/01 14:32:06 Database connection established successfully
2026/08/01 14:32:06 LDAP service initialized successfully
2026/08/01 14:32:08 找到用户 DN: CN=杨龙威，OU=IT 部，OU=HOT,DC=hot,DC=local
2026/08/01 14:32:08 用户 ylw 密码验证成功
2026/08/01 14:32:08 Error searching security group: LDAP Result Code 34 "Invalid DN Syntax"  <-- 第一次失败
```

修复后的日志：
```
2026/08/01 14:32:08 找到用户 DN: CN=杨龙威，OU=IT 部，OU=HOT,DC=hot,DC=local
2026/08/01 14:32:08 用户 ylw 密码验证成功
← 没有错误日志 → 认证成功
```

### API 响应
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 0,
  "user_info": {
    "username": "ylw",
    "email": "ylw@"
  }
}
```

---

## 🎯 总结

### 核心原则
1. **LDAPS 认证两步走**：先用管理员 Bind 搜索用户，再用用户 DN 验证密码
2. **每次操作新建连接**：避免复用未授权的连接进行权限敏感操作
3. **YAML 配置必须加 Tag**：否则无法正确映射
4. **特殊字符需转义**：中文 DN 使用 `\uXXXX` 格式

### 参考资源
- [DCPM 项目](https://github.com/YanGLweI/DCPM.git) - 学习其实现方式
- [IT 项目](https://github.com/YanGLweI/IT.git) - **最终确定的解决方案来源**

---

*问题解决日期：2026 年 8 月 1 日*