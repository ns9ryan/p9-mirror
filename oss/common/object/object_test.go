package object

import (
	"strings"
	"testing"
	"time"
)

func TestParsePlannedKey(t *testing.T) {
	op, scene, ext, ok := ParsePlannedKeyParts("public/20260825/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.mp4")
	if !ok || op != "" || scene != "public" || ext != ".mp4" {
		t.Fatalf("no op: op=%q scene=%q ext=%q ok=%v", op, scene, ext, ok)
	}
	op, scene, ext, ok = ParsePlannedKeyParts("demo/public/20260825/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg")
	if !ok || op != "demo" || scene != "public" || ext != ".jpg" {
		t.Fatalf("with op: op=%q scene=%q ext=%q ok=%v", op, scene, ext, ok)
	}
	op, scene, ext, ok = ParsePlannedKeyParts("op-a/kyc/20260826/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.png")
	if !ok || op != "op-a" || scene != "kyc" || ext != ".png" {
		t.Fatalf("hyphen op: op=%q scene=%q ext=%q ok=%v", op, scene, ext, ok)
	}
	if _, _, _, ok := ParsePlannedKeyParts("_/public/20260825/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.mp4"); ok {
		t.Fatal("underscore placeholder should not be accepted as operator")
	}
	if _, _, ok := ParsePlannedKey("../secret"); ok {
		t.Fatal("rejected path should fail")
	}
}

func TestPlannedKeyMatchesScene(t *testing.T) {
	key := "demo/private/20260825/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg"
	ext, ok := PlannedKeyMatchesScene(key, "private")
	if !ok || ext != ".jpg" {
		t.Fatalf("match: ext=%q ok=%v", ext, ok)
	}
	if _, ok := PlannedKeyMatchesScene(key, "public"); ok {
		t.Fatal("scene mismatch should fail")
	}
	key = "private/20260825/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg"
	ext, ok = PlannedKeyMatchesScene(key, "private")
	if !ok || ext != ".jpg" {
		t.Fatalf("no-op match: ext=%q ok=%v", ext, ok)
	}
}

func TestPlanObjectKeyPrefix(t *testing.T) {
	key, err := PlanObjectKey("public", ".png", "")
	if err != nil {
		t.Fatal(err)
	}
	op, scene, ext, ok := ParsePlannedKeyParts(key)
	if !ok || op != "" || scene != "public" || ext != ".png" {
		t.Fatalf("empty op: %q", key)
	}
	if !strings.HasPrefix(key, "public/") {
		t.Fatalf("want public/ prefix, got %q", key)
	}
	if !strings.Contains(key, "/"+time.Now().Format("20060102")+"/") {
		t.Fatalf("missing today in %q", key)
	}

	key, err = PlanObjectKey("public", ".png", "demo")
	if err != nil {
		t.Fatal(err)
	}
	op, scene, ext, ok = ParsePlannedKeyParts(key)
	if !ok || op != "demo" || scene != "public" || ext != ".png" {
		t.Fatalf("with operator: %q", key)
	}
	if !strings.HasPrefix(key, "demo/public/") {
		t.Fatalf("want demo/public/ prefix, got %q", key)
	}

	key, err = PlanObjectKey("public", ".png", "../bad")
	if err != nil {
		t.Fatal(err)
	}
	op, _, _, ok = ParsePlannedKeyParts(key)
	if !ok || op != "" {
		t.Fatalf("invalid operator should omit prefix: %q", key)
	}
}

func TestExtAllowedEmptyMeansAll(t *testing.T) {
	if !ExtAllowed(nil, ".webp") || !ExtAllowed([]string{}, ".exe") {
		t.Fatal("empty allowlist should allow any ext")
	}
	if ExtAllowed([]string{".jpg"}, ".png") {
		t.Fatal("explicit list should reject others")
	}
	if !ExtAllowed([]string{".jpg"}, ".jpg") {
		t.Fatal("explicit list should allow listed ext")
	}
}

func TestPersistURL(t *testing.T) {
	got := PersistURL("https://resource.example.com/res", "demo/public/20260824/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png")
	want := "https://resource.example.com/res/demo/public/20260824/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestKeyFromLocator(t *testing.T) {
	const key = "demo/public/20260824/fda7417083cbc74acbc753a83c924e25.png"
	const base = "https://resource.crm6688.com/res"
	cases := []string{
		key,
		base + "/" + key,
		base + "/oss-bucket/" + key,
		"http://127.0.0.1:9200/oss-bucket/" + key,
	}
	for _, raw := range cases {
		if got := KeyFromLocator(raw, base); got != key {
			t.Fatalf("locator %q: got %q", raw, got)
		}
	}
	plat := "public/20260824/fda7417083cbc74acbc753a83c924e25.png"
	if got := KeyFromLocator(base+"/oss-bucket/"+plat, base); got != plat {
		t.Fatalf("no-op key: got %q want %q", got, plat)
	}
	if got := KeyFromLocator(plat, base); got != plat {
		t.Fatalf("bare no-op: got %q", got)
	}
	if KeyFromLocator("https://cdn.example/other.png", base) != "" {
		t.Fatal("unrelated path should be empty")
	}
}
