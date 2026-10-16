package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"oa.98ent.com/p9/nginx-manager/internal/api"
)

func main() {
	// 启动 nginx
	nginxCmd := exec.Command("nginx", "-g", "daemon off;")
	nginxCmd.Stdout = os.Stdout
	nginxCmd.Stderr = os.Stderr

	fmt.Println("[INIT] Starting nginx...")
	if err := nginxCmd.Start(); err != nil {
		fmt.Printf("[ERROR] Failed to start nginx: %v\n", err)
		os.Exit(1)
	}

	nginxPID := nginxCmd.Process.Pid
	fmt.Printf("[INIT] Nginx started (PID: %d)\n", nginxPID)

	// 启动 Reload API 服务
	go api.StartReloadAPI("8888")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("[INIT] Reload API started on :8888")

	// 监听信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan
	fmt.Println("[SHUTDOWN] Received signal, shutting down...")

	// 优雅关闭 nginx
	syscall.Kill(nginxPID, syscall.SIGTERM)
	nginxCmd.Wait()

	fmt.Println("[SHUTDOWN] Shutdown complete")
}
