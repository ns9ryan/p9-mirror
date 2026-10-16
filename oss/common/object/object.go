package object

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// 本文件：对象 key 规划与从 url 还原 key。

var ErrInvalidKey = errors.New("invalid key")

// 有厅：{operator_code}/{scene}/{yyyyMMdd}/{32hex}{.ext}
// 无厅：{scene}/{yyyyMMdd}/{32hex}{.ext}

var (
	plannedKeyWithOpRe = regexp.MustCompile(
		`^([a-zA-Z0-9][a-zA-Z0-9_-]{0,63})/([a-z][a-z0-9_]{0,31})/([0-9]{8})/([0-9a-f]{32})(\.[a-z0-9]+)?$`,
	)
	plannedKeyNoOpRe = regexp.MustCompile(
		`^([a-z][a-z0-9_]{0,31})/([0-9]{8})/([0-9a-f]{32})(\.[a-z0-9]+)?$`,
	)
	plannedKeyFindRe = regexp.MustCompile(
		`(?:^|/)((?:[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}/)?[a-z][a-z0-9_]{0,31}/[0-9]{8}/[0-9a-f]{32}(?:\.[a-z0-9]+)?)`,
	)
	operatorCodeRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)
)

func randomObjectName(ext string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]) + ext, nil
}

// SanitizeOperatorCode 目录段厅编码；空或非法时返回空串。
func SanitizeOperatorCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" || !operatorCodeRe.MatchString(code) {
		return ""
	}
	return code
}

// PlanObjectKey 规划对象路径。有厅时 operator_code 在最前；无厅则直接 scene/...。
func PlanObjectKey(scene, ext, operatorCode string) (string, error) {
	name, err := randomObjectName(ext)
	if err != nil {
		return "", err
	}
	day := time.Now().Format("20060102")
	base := scene + "/" + day + "/" + name
	if op := SanitizeOperatorCode(operatorCode); op != "" {
		return op + "/" + base, nil
	}
	return base, nil
}

// ExtAllowed 空白名单表示允许任意后缀。
func ExtAllowed(allowed []string, ext string) bool {
	if len(allowed) == 0 {
		return true
	}
	ext = strings.ToLower(ext)
	for _, a := range allowed {
		if a == ext {
			return true
		}
	}
	return false
}

// ParsePlannedKey 校验并拆出 key 的 scene 与后缀。
func ParsePlannedKey(key string) (scene, ext string, ok bool) {
	_, scene, ext, ok = ParsePlannedKeyParts(key)
	return scene, ext, ok
}

// ParsePlannedKeyParts 拆出 operator_code（无厅为空）、scene、后缀。
// 优先匹配有厅四段，再匹配无厅三段。
func ParsePlannedKeyParts(key string) (operatorCode, scene, ext string, ok bool) {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if m := plannedKeyWithOpRe.FindStringSubmatch(key); len(m) >= 6 {
		return m[1], m[2], strings.ToLower(m[5]), true
	}
	if m := plannedKeyNoOpRe.FindStringSubmatch(key); len(m) >= 5 {
		return "", m[1], strings.ToLower(m[4]), true
	}
	return "", "", "", false
}

// PlannedKeyMatchesScene 检查 key 是否属于指定 scene。
func PlannedKeyMatchesScene(key, scene string) (ext string, ok bool) {
	got, ext, ok := ParsePlannedKey(key)
	if !ok || got != scene {
		return "", false
	}
	return ext, true
}

// PersistURL 拼 {file_url}/{key}，不含桶名。
func PersistURL(base, key string) string {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if key == "" {
		return ""
	}
	if base == "" {
		return key
	}
	return base + "/" + key
}

// parseExactKey 整段或去掉 path-style 桶名后的规划 key。
// 有厅与无厅都能匹配时优先更长的（带 operator_code）。
func parseExactKey(s string) string {
	s = strings.TrimPrefix(strings.TrimSpace(s), "/")
	if s == "" {
		return ""
	}
	best := ""
	if _, _, ok := ParsePlannedKey(s); ok {
		best = s
	}
	if i := strings.IndexByte(s, '/'); i > 0 {
		rest := s[i+1:]
		if _, _, ok := ParsePlannedKey(rest); ok {
			// path-style：{bucket}/{key}；桶名含 - 时优先 rest，避免误把桶当厅。
			if best == "" || strings.Contains(s[:i], "-") {
				return rest
			}
			// 否则保留更长的有厅 key（如 demo/public/... 不要剥成 public/...）
			if len(rest) > len(best) {
				return rest
			}
		}
	}
	return best
}

func findPlannedKey(s string) string {
	s = strings.TrimPrefix(strings.TrimSpace(s), "/")
	if s == "" {
		return ""
	}
	if key := parseExactKey(s); key != "" {
		return key
	}
	best := ""
	for _, g := range plannedKeyFindRe.FindAllStringSubmatch("/"+s, -1) {
		if len(g) < 2 {
			continue
		}
		if _, _, ok := ParsePlannedKey(g[1]); ok && len(g[1]) >= len(best) {
			best = g[1]
		}
	}
	return best
}

// KeyFromLocator 从 url 或裸 key 还原对象 key。
func KeyFromLocator(raw, fileBase string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		return findPlannedKey(raw)
	}

	path := ""
	if u, err := url.Parse(raw); err == nil && u.Scheme != "" && u.Host != "" {
		path = strings.TrimPrefix(u.Path, "/")
	}

	fileBase = strings.TrimRight(strings.TrimSpace(fileBase), "/")
	if fileBase != "" {
		if bu, err := url.Parse(fileBase); err == nil && bu.Scheme != "" && bu.Host != "" {
			fbPath := strings.Trim(bu.Path, "/")
			if fbPath != "" && path != "" && strings.HasPrefix(path, fbPath+"/") {
				if key := findPlannedKey(strings.TrimPrefix(path, fbPath+"/")); key != "" {
					return key
				}
			}
		}
		for _, c := range []string{raw, path} {
			if c != "" && strings.HasPrefix(c, fileBase+"/") {
				if key := findPlannedKey(strings.TrimPrefix(c, fileBase+"/")); key != "" {
					return key
				}
			}
		}
	}
	return findPlannedKey(path)
}
