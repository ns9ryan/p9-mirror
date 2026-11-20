package object

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// 本文件：对象 key 规划与从 url 还原 key。

var ErrInvalidKey = errors.New("invalid key")

// 有厅：{operator_code}/{scene}/{yyyyMMdd}/{32hex}{.ext}；站点 Logo：logo/{OP_站点编码}/{yyyyMMdd}/{32hex}{.ext}
// 无厅：{scene}/{yyyyMMdd}/{32hex}{.ext}
// 总网默认 VIP 图标：vipicon/{等级下限}-{等级上限}/{32hex}.png
// 分站 VIP 图标：vipicon/{operator_code}/{等级下限}-{等级上限}/{32hex}.png

var (
	plannedAvatarKeyRe = regexp.MustCompile(
		`^avatar-resource/avatars/(?:[0-9a-f]{32}|avatar[1-8]-[0-9a-f]{16})\.png$`,
	)
	plannedAvatarFrameKeyRe = regexp.MustCompile(
		`^avatar-resource/avatar-frames/(?:[0-9a-f]{32}|frames[1-6]-[0-9a-f]{16})\.png$`,
	)
	plannedVipIconKeyRe = regexp.MustCompile(
		`^vipicon/(0-9|10-19|20-29|30-39|40-49|50-59|60-69|70-79|80-89|90-99)/([0-9a-f]{32})(\.png)$`,
	)
	plannedOperatorVipIconKeyRe = regexp.MustCompile(
		`^vipicon/([a-zA-Z0-9][a-zA-Z0-9_-]{0,63})/` +
			`(0-9|10-19|20-29|30-39|40-49|50-59|60-69|70-79|80-89|90-99)/[0-9a-f]{32}\.png$`,
	)
	// 站点 Logo 固定在 logo 目录下，依次捕获站点编码、日期、随机文件名和后缀。
	plannedSiteLogoKeyRe = regexp.MustCompile(
		`^logo/(OP_[a-zA-Z0-9_-]{1,61})/([0-9]{8})/([0-9a-f]{32})(\.[a-z0-9]+)?$`,
	)
	plannedKeyWithOpRe = regexp.MustCompile(
		`^([a-zA-Z0-9][a-zA-Z0-9_-]{0,63})/([a-z][a-z0-9_]{0,31})/([0-9]{8})/([0-9a-f]{32})(\.[a-z0-9]+)?$`,
	)
	plannedKeyNoOpRe = regexp.MustCompile(
		`^([a-z][a-z0-9_]{0,31})/([0-9]{8})/([0-9a-f]{32})(\.[a-z0-9]+)?$`,
	)
	plannedKeyFindRe = regexp.MustCompile(
		`(?:^|/)(` +
			`avatar-resource/(?:` +
			`avatars/(?:[0-9a-f]{32}|avatar[1-8]-[0-9a-f]{16})|` +
			`avatar-frames/(?:[0-9a-f]{32}|frames[1-6]-[0-9a-f]{16}))\.png|` +
			`vipicon/(?:[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}/)?` +
			`(?:0-9|10-19|20-29|30-39|40-49|50-59|60-69|70-79|80-89|90-99)/[0-9a-f]{32}\.png|` +
			`(?:(?:[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}/)?[a-z][a-z0-9_]{0,31}|logo/OP_[a-zA-Z0-9_-]{1,61})/` +
			`[0-9]{8}/[0-9a-f]{32}(?:\.[a-z0-9]+)?)`,
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

// PlanObjectKey 规划对象路径。OP_ 站点 Logo 按场景、站点分组，其他路径沿用原规则。
func PlanObjectKey(scene, ext, operatorCode string) (string, error) {
	// 总站素材放在统一目录，不带分站编码。
	if scene == "default_avatar" || scene == "default_avatar_frame" {
		if ext != ".png" {
			return "", ErrInvalidKey
		}
		return PlanAvatarResourceKey(scene)
	}
	// VIP 图标必须明确等级组，只能通过专用规划方法生成路径。
	if scene == "vipicon" {
		return "", ErrInvalidKey
	}
	name, err := randomObjectName(ext)
	if err != nil {
		return "", err
	}
	day := time.Now().Format("20060102")
	base := scene + "/" + day + "/" + name
	if op := SanitizeOperatorCode(operatorCode); op != "" {
		// OP_ 站点使用 logo/站点/日期/文件名，其他场景或编码继续沿用原目录规则。
		if scene == "logo" && strings.HasPrefix(op, "OP_") && len(op) > 3 {
			return "logo/" + op + "/" + day + "/" + name, nil
		}
		return op + "/" + base, nil
	}
	return base, nil
}

// PlanVipIconKey 按固定等级组规划总网默认图标路径，不覆盖同组的旧图片。
func PlanVipIconKey(minLevel int64) (string, error) {
	if minLevel < 0 || minLevel > 90 || minLevel%10 != 0 {
		return "", ErrInvalidKey
	}
	name, err := randomObjectName(".png")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("vipicon/%d-%d/%s", minLevel, minLevel+9, name), nil
}

// PlanOperatorVipIconKey 复用总站等级及随机文件名规则，必须加上有效的登录站点目录。
func PlanOperatorVipIconKey(operatorCode string, minLevel int64) (string, error) {
	operatorCode = SanitizeOperatorCode(operatorCode)
	if operatorCode == "" {
		return "", ErrInvalidKey
	}
	key, err := PlanVipIconKey(minLevel)
	if err != nil {
		return "", err
	}
	return "vipicon/" + operatorCode + "/" + strings.TrimPrefix(key, "vipicon/"), nil
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
// 优先匹配 VIP 等级目录和站点 Logo 目录，再兼容有厅四段和无厅三段。
func ParsePlannedKeyParts(key string) (operatorCode, scene, ext string, ok bool) {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if plannedAvatarKeyRe.MatchString(key) {
		return "", "default_avatar", ".png", true
	}
	if plannedAvatarFrameKeyRe.MatchString(key) {
		return "", "default_avatar_frame", ".png", true
	}
	// 先识别分站目录，避免把本站素材误认成总站默认图标。
	if matches := plannedOperatorVipIconKeyRe.FindStringSubmatch(key); len(matches) == 3 {
		return matches[1], "vipicon", ".png", true
	}
	if plannedVipIconKeyRe.MatchString(key) {
		return "", "vipicon", ".png", true
	}
	// 从 Logo 专用目录还原站点身份及场景，供上传结果与素材归属校验复用。
	if matches := plannedSiteLogoKeyRe.FindStringSubmatch(key); len(matches) >= 5 {
		return matches[1], "logo", strings.ToLower(matches[4]), true
	}
	if m := plannedKeyWithOpRe.FindStringSubmatch(key); len(m) >= 6 {
		return m[1], m[2], strings.ToLower(m[5]), true
	}
	if m := plannedKeyNoOpRe.FindStringSubmatch(key); len(m) >= 5 {
		return "", m[1], strings.ToLower(m[4]), true
	}
	return "", "", "", false
}

// PlanAvatarResourceKey 每次替换生成新文件名，避免覆盖已被会员使用的旧图片。
func PlanAvatarResourceKey(scene string) (string, error) {
	directory := ""
	switch scene {
	case "default_avatar":
		directory = "avatars"
	case "default_avatar_frame":
		directory = "avatar-frames"
	default:
		return "", ErrInvalidKey
	}
	name, err := randomObjectName(".png")
	if err != nil {
		return "", err
	}
	return "avatar-resource/" + directory + "/" + name, nil
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
