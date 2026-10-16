package domain

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/node-agent/internal/config"
)

// Manager Nginx 配置管理器
type Manager struct {
	config config.Config
	logx.Logger
}

// NewManager 创建 Nginx 管理器
// confDirPath: nginx 配置目录，未设置时返回 nil
// containerName: docker 容器名称，默认为 p9-nginx
func NewManager(ctx context.Context, config config.Config) *Manager {

	return &Manager{
		config: config,
		Logger: logx.WithContext(ctx),
	}
}

// GenerateConfig 生成域名对应的 nginx 配置文件
func (m *Manager) GenerateConfig(domain string) error {
	if m == nil {
		m.Infof("Nginx 未配置，跳过生成域名配置: domain=%s", domain)
		return nil
	}

	if domain == "" {
		return fmt.Errorf("域名不能为空")
	}

	// 确保配置目录存在
	if err := os.MkdirAll(m.config.Nginx.ConfDirPath, 0755); err != nil {
		m.Errorf("创建 nginx 配置目录失败: path=%s, error=%v", m.config.Nginx.ConfDirPath, err)
		return fmt.Errorf("创建目录失败: %w", err)
	}

	configPath := filepath.Join(m.config.Nginx.ConfDirPath, domain+".conf")
	content := m.buildConfigContent(domain)
	m.Infof("生成 nginx 配置内容: domain=%s, content=\n%s", domain, content)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		m.Errorf("写入 nginx 配置文件失败: path=%s, error=%v", configPath, err)
		return fmt.Errorf("写入文件失败: %w", err)
	}

	m.Infof("Nginx 配置文件已生成: domain=%s, path=%s", domain, configPath)
	return nil
}

// DeleteConfig 删除域名对应的 nginx 配置文件
func (m *Manager) DeleteConfig(domain string) error {
	if m == nil {
		return nil
	}

	if domain == "" {
		return fmt.Errorf("域名不能为空")
	}

	configPath := filepath.Join(m.config.Nginx.ConfDirPath, domain+".conf")
	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			m.Infof("Nginx 配置文件不存在，无需删除: domain=%s, path=%s", domain, configPath)
			return nil
		}
		m.Errorf("删除 nginx 配置文件失败: path=%s, error=%v", configPath, err)
		return fmt.Errorf("删除文件失败: %w", err)
	}

	m.Infof("Nginx 配置文件已删除: domain=%s, path=%s", domain, configPath)
	return nil
}

// buildConfigContent 构建 nginx 配置内容
func (m *Manager) buildConfigContent(domain string) string {
	var sb strings.Builder
	ports := m.config.Nginx.PortMap

	sb.WriteString(fmt.Sprintf("# Nginx config for %s\n", domain))
	sb.WriteString(fmt.Sprintf("server {\n"))
	sb.WriteString(fmt.Sprintf("    listen 80;\n"))
	sb.WriteString(fmt.Sprintf("    server_name %s;\n\n", domain))
	sb.WriteString(fmt.Sprintf("    root /var/www/outstation;\n"))
	sb.WriteString(fmt.Sprintf("    index index.html;\n\n"))

	// Core API
	if corePort := ports.Core; corePort != 0 {
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    # Core API\n"))
		sb.WriteString(fmt.Sprintf("    # /admin/core/* → /core/*\n"))
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    location /admin/core/ {\n"))
		sb.WriteString(fmt.Sprintf("        rewrite ^/admin/core/(.*)$ /core/$1 break;\n\n"))
		sb.WriteString(fmt.Sprintf("        proxy_pass http://host.docker.internal:%d;\n\n", corePort))
		sb.WriteString(fmt.Sprintf("        proxy_set_header Host $host;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Real-IP $remote_addr;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-Proto $scheme;\n"))
		sb.WriteString(fmt.Sprintf("    }\n\n"))
	}

	// Base API
	if basePort := ports.Base; basePort != 0 {
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    # Base API\n"))
		sb.WriteString(fmt.Sprintf("    # /admin/base/* → /admin/*\n"))
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    location /admin/base/ {\n"))
		sb.WriteString(fmt.Sprintf("        rewrite ^/admin/base/(.*)$ /admin/$1 break;\n\n"))
		sb.WriteString(fmt.Sprintf("        proxy_pass http://host.docker.internal:%d;\n\n", basePort))
		sb.WriteString(fmt.Sprintf("        proxy_set_header Host $host;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Real-IP $remote_addr;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-Proto $scheme;\n"))
		sb.WriteString(fmt.Sprintf("    }\n\n"))
	}

	// Game API
	if gamePort := ports.Game; gamePort != 0 {
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    # Game API\n"))
		sb.WriteString(fmt.Sprintf("    # /admin/game/* → /operator-game/*\n"))
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    location /admin/game/ {\n"))
		sb.WriteString(fmt.Sprintf("        rewrite ^/admin/game/(.*)$ /operator-game/$1 break;\n\n"))
		sb.WriteString(fmt.Sprintf("        proxy_pass http://host.docker.internal:%d;\n\n", gamePort))
		sb.WriteString(fmt.Sprintf("        proxy_set_header Host $host;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Real-IP $remote_addr;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-Proto $scheme;\n"))
		sb.WriteString(fmt.Sprintf("    }\n\n"))
	}

	// Operator API
	if operatorPort := ports.Operator; operatorPort != 0 {
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    # Operator API\n"))
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    location /admin/operator/ {\n"))
		sb.WriteString(fmt.Sprintf("        proxy_pass http://host.docker.internal:%d;\n\n", operatorPort))
		sb.WriteString(fmt.Sprintf("        proxy_set_header Host $host;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Real-IP $remote_addr;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-Proto $scheme;\n"))
		sb.WriteString(fmt.Sprintf("    }\n\n"))
	}

	// Dispatch API
	if dispatchPort := ports.Dispatch; dispatchPort != 0 {
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    # Dispatch API\n"))
		sb.WriteString(fmt.Sprintf("    # =========================\n"))
		sb.WriteString(fmt.Sprintf("    location /admin/dispatch/ {\n"))
		sb.WriteString(fmt.Sprintf("        proxy_pass http://host.docker.internal:%d;\n\n", dispatchPort))
		sb.WriteString(fmt.Sprintf("        proxy_set_header Host $host;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Real-IP $remote_addr;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n"))
		sb.WriteString(fmt.Sprintf("        proxy_set_header X-Forwarded-Proto $scheme;\n"))
		sb.WriteString(fmt.Sprintf("    }\n\n"))
	}

	// 前端 SPA
	sb.WriteString(fmt.Sprintf("    # =========================\n"))
	sb.WriteString(fmt.Sprintf("    # 前端 SPA\n"))
	sb.WriteString(fmt.Sprintf("    # =========================\n"))
	sb.WriteString(fmt.Sprintf("    location / {\n"))
	sb.WriteString(fmt.Sprintf("        try_files $uri $uri/ /index.html;\n"))
	sb.WriteString(fmt.Sprintf("    }\n"))
	sb.WriteString(fmt.Sprintf("}\n"))

	return sb.String()
}

// ReloadNginx 重载 nginx（通过调用 nginx reload API）
func (m *Manager) ReloadNginx() (string, error) {
	if m == nil || m.config.Nginx.ConfDirPath == "" {
		return "", nil
	}

	// 通过 HTTP API 调用 nginx reload
	containerName := m.config.Nginx.ContainerName
	if containerName == "" {
		containerName = "p9-nginx"
	}

	apiURL := fmt.Sprintf("http://%s:8888/internal/nginx/reload", containerName)
	m.Infof("调用 nginx reload API: url=%s", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		m.Errorf("调用 nginx reload API 失败: url=%s, error=%v", apiURL, err)
		return "", fmt.Errorf("nginx reload 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := string(body)

	if resp.StatusCode != http.StatusOK {
		m.Errorf("nginx reload 失败: status=%d, response=%s", resp.StatusCode, result)
		return result, fmt.Errorf("nginx reload 失败: status=%d", resp.StatusCode)
	}

	m.Infof("Nginx reload 成功: response=%s", result)
	return result, nil
}
