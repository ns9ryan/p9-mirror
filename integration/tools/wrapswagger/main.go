// 把 goctls 生成的 swagger 200 响应包成实际 HTTP 外壳 {code,msg,data}。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	for _, name := range []string{"swagger.json", "swagger.yaml"} {
		path := filepath.Join(root, "docs", name)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if filepath.Ext(name) == ".yaml" {
			// yaml 由 goctls 再生成；只修补 json，yaml 用 json 再转太重，这里同样读 json 逻辑不处理 yaml 结构
			// 仍跑一遍：若是 yaml 文件则跳过（goctls yaml 与 json 同源，以 json 为准挂 UI）
			continue
		}
		if err := wrapFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "wrap %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Println("wrapped", path)
	}
}

func wrapFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return err
	}

	info, _ := doc["info"].(map[string]any)
	if info == nil {
		info = map[string]any{}
		doc["info"] = info
	}
	info["description"] = "P9 第三方接入：IP/GEO。路径前缀 /p9/integration/ipgeo。成功体为 {code:0,msg,data}；失败无 data，HTTP status=code。全部接口需 Authorization: Bearer。"

	if sec, ok := doc["securityDefinitions"].(map[string]any); ok {
		if token, ok := sec["Token"].(map[string]any); ok {
			token["description"] = "JWT，格式：Bearer <access_token>"
		}
	}

	paths, _ := doc["paths"].(map[string]any)
	for _, item := range paths {
		methods, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for _, opVal := range methods {
			op, ok := opVal.(map[string]any)
			if !ok {
				continue
			}
			resps, ok := op["responses"].(map[string]any)
			if !ok {
				continue
			}
			r200, ok := resps["200"].(map[string]any)
			if !ok {
				continue
			}
			schema, _ := r200["schema"].(map[string]any)
			if schema == nil {
				continue
			}
			// 已包过则跳过
			if props, _ := schema["properties"].(map[string]any); props != nil {
				if _, hasCode := props["code"]; hasCode {
					if _, hasData := props["data"]; hasData {
						continue
					}
				}
			}
			r200["description"] = "成功"
			r200["schema"] = map[string]any{
				"type": "object",
				"required": []any{"code", "msg"},
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "integer",
						"description": "业务码，成功恒为 0",
						"example":     0,
					},
					"msg": map[string]any{
						"type":        "string",
						"description": "成功为 ok（可随语言）",
						"example":     "ok",
					},
					"data": schema,
				},
			}
			resps["400"] = errResp("参数错误")
			resps["401"] = errResp("未登录 / Token 无效")
			resps["403"] = errResp("无权限")
			resps["404"] = errResp("资源不存在")
			resps["500"] = errResp("服务错误")
		}
	}

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}

func errResp(desc string) map[string]any {
	return map[string]any{
		"description": desc,
		"schema": map[string]any{
			"type": "object",
			"required": []any{"code", "msg"},
			"properties": map[string]any{
				"code": map[string]any{"type": "integer", "description": "与 HTTP status 相同"},
				"msg":  map[string]any{"type": "string", "description": "错误原因"},
			},
		},
	}
}
