package api

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// HandleReload 处理 nginx reload 请求
func HandleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 验证配置
	validateCmd := exec.Command("nginx", "-t")
	validateOutput, err := validateCmd.CombinedOutput()
	validateStr := string(validateOutput)

	fmt.Printf("[nginx -t] %s\n", validateStr)

	if err != nil {
		http.Error(w, fmt.Sprintf("Nginx validation failed: %s", validateStr), http.StatusBadRequest)
		return
	}

	// 重载 nginx
	reloadCmd := exec.Command("nginx", "-s", "reload")
	reloadOutput, err := reloadCmd.CombinedOutput()
	reloadStr := string(reloadOutput)

	fmt.Printf("[nginx -s reload] %s\n", reloadStr)

	if err != nil {
		http.Error(w, fmt.Sprintf("Nginx reload failed: %s", reloadStr), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success":true,"validate":"%s","reload":"%s"}`,
		strings.TrimSpace(validateStr),
		strings.TrimSpace(reloadStr),
	)
}

// StartReloadAPI 启动 reload API 服务
func StartReloadAPI(port string) {
	http.HandleFunc("/internal/nginx/reload", HandleReload)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	fmt.Printf("Reload API listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
