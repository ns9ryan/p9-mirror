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

func TestPlanSiteLogoObjectKey(testContext *testing.T) {
	for _, operatorCode := range []string{"OP_57BE18", "OP_001", "OP_a-b_9"} {
		testContext.Run(operatorCode, func(testContext *testing.T) {
			key, err := PlanObjectKey("logo", ".png", operatorCode)
			if err != nil {
				testContext.Fatal(err)
			}
			if !strings.HasPrefix(key, "logo/"+operatorCode+"/"+time.Now().Format("20060102")+"/") {
				testContext.Fatalf("unexpected logo directory: %q", key)
			}
			gotOperator, scene, extension, ok := ParsePlannedKeyParts(key)
			if !ok || gotOperator != operatorCode || scene != "logo" || extension != ".png" {
				testContext.Fatalf("key did not round trip: %q", key)
			}
		})
	}
}

func TestSiteLogoKeyCompatibility(testContext *testing.T) {
	for _, testCase := range []struct {
		key          string
		operatorCode string
		scene        string
	}{
		{"logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png", "OP_57BE18", "logo"},
		{"OP_57BE18/logo/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png", "OP_57BE18", "logo"},
		{"OP_57BE18/public/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png", "OP_57BE18", "public"},
		{"logo/public/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png", "logo", "public"},
		{"demo/logo/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png", "demo", "logo"},
		{"logo/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png", "", "logo"},
	} {
		testContext.Run(testCase.key, func(testContext *testing.T) {
			operatorCode, scene, extension, ok := ParsePlannedKeyParts(testCase.key)
			if !ok || operatorCode != testCase.operatorCode || scene != testCase.scene || extension != ".png" {
				testContext.Fatalf("operator=%q scene=%q extension=%q ok=%v", operatorCode, scene, extension, ok)
			}
			if _, ok := PlannedKeyMatchesScene(testCase.key, testCase.scene); !ok {
				testContext.Fatal("key must resolve to its upload scene")
			}
		})
	}
}

func TestSiteLogoKeyFromLocator(testContext *testing.T) {
	const key = "logo/OP_57BE18/20261027/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"
	const base = "https://cdn.example.com/assets"
	for _, locator := range []string{
		key,
		base + "/" + key,
		base + "/oss-bucket/" + key,
		"http://127.0.0.1:9200/oss-bucket/" + key,
		"https://cdn.example.com/extra/assets/" + key,
	} {
		if got := KeyFromLocator(locator, base); got != key {
			testContext.Fatalf("locator=%q key=%q", locator, got)
		}
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
