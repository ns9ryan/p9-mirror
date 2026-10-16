package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

// registerSwagger 仅在开发 / 测试环境挂 /swagger；线上 Mode=pro 不注册。
// 直接读仓库 docs/swagger.json（仅本地开发用，不 embed、不复制）。
func registerSwagger(server *rest.Server, mode string) {
	if mode != service.DevMode && mode != service.TestMode {
		return
	}
	spec, err := readSwaggerJSON()
	if err != nil {
		logx.Errorf("swagger: skip /swagger, %v", err)
		return
	}
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>Integration API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      spec: %s,
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis],
      layout: "BaseLayout"
    });
  </script>
</body>
</html>
`, spec)

	server.AddRoutes([]rest.Route{
		{
			Method: http.MethodGet,
			Path:   "/swagger",
			Handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(html))
			},
		},
	})
}

// go run ./api → docs/swagger.json；在 api/ 下启动 → ../docs/swagger.json
func readSwaggerJSON() ([]byte, error) {
	for _, p := range []string{"docs/swagger.json", "../docs/swagger.json"} {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("docs/swagger.json not found (cwd relative)")
}
