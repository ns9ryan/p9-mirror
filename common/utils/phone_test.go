package utils

import "testing"

func TestCallingCodeByRegion(t *testing.T) {
	cases := []struct {
		region string
		code   string
	}{
		{"my", "60"},
		{"CN", "86"},
		{"HK", "852"},
		{"MO", "853"},
		{"TW", "886"},
		{"RU", "7"},
		{"KZ", "7"},
		{"US", "1"},
		{"BR", "55"},
	}
	for _, c := range cases {
		code, ok := CallingCodeByRegion(c.region)
		if !ok || code != c.code {
			t.Fatalf("%s=%s ok=%v want=%s", c.region, code, ok, c.code)
		}
	}
	if _, ok := CallingCodeByRegion("XX"); ok {
		t.Fatal("XX should be unsupported")
	}
}

func TestCallingCodesComplete(t *testing.T) {
	list := CallingCodes()
	if n := len(list); n < 190 {
		t.Fatalf("calling codes=%d, want >= 190", n)
	}
	seen := make(map[string]struct{}, len(list))
	for _, item := range list {
		if item.Region == "" || item.Code == "" {
			t.Fatalf("empty entry %+v", item)
		}
		if _, ok := seen[item.Region]; ok {
			t.Fatalf("duplicate region %s", item.Region)
		}
		seen[item.Region] = struct{}{}
		if _, ok := callingCodeNSN[item.Code]; !ok {
			t.Fatalf("missing nsn range for %s", item.Code)
		}
	}
}

func TestValidPhone(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"60-123456789", true},
		{"+60-123456789", true},
		{" 60-123456789 ", true},
		{"65-12345678", true},
		{"1-2025550123", true},
		{"86-13800138000", true},
		{"+86-13800138000", true},
		{"852-91234567", true},
		{"886-912345678", true},
		{"7-9123456789", true},
		{"44-7911123456", true},
		{"33-612345678", true},
		{"81-9012345678", true},
		{"971-501234567", true},
		{"234-8012345678", true},
		{"86-1380013800", false},
		{"999-123456789", false},
		{"60-12345", false},
		{"60_123456789", false},
		{"60123456789", false},
		{"", false},
		{"60-", false},
		{"-123456789", false},
		{"60-12a456789", false},
	}
	for _, c := range cases {
		if got := ValidPhone(c.in); got != c.want {
			t.Fatalf("%q got=%v want=%v", c.in, got, c.want)
		}
	}
}

func TestParseAndFormatPhone(t *testing.T) {
	code, nsn, ok := ParsePhone("+86-13800138000")
	if !ok || code != "86" || nsn != "13800138000" {
		t.Fatalf("parse=%s %s %v", code, nsn, ok)
	}
	if got := FormatPhone("86", "13800138000"); got != "86-13800138000" {
		t.Fatalf("format=%q", got)
	}
	if FormatPhone("", "123") != "" {
		t.Fatal("empty calling code")
	}
}
