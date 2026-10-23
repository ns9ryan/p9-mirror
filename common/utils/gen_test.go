package utils

import (
	"strings"
	"testing"
	"unicode"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func TestGenInvalidLength(t *testing.T) {
	if _, err := Gen(0); err == nil {
		t.Fatal("expected invalid length")
	}
	if _, err := Gen(-1); err == nil {
		t.Fatal("expected invalid length")
	}
}

func TestGenLengthAndAlphabet(t *testing.T) {
	for _, n := range []int{1, 3, 5, ShortIDLength, LongIDLength} {
		id, err := Gen(n)
		if err != nil {
			t.Fatalf("n=%d err=%v", n, err)
		}
		if len(id) != n {
			t.Fatalf("n=%d len=%d id=%q", n, len(id), id)
		}
		for _, r := range id {
			if !strings.ContainsRune(base62Alphabet, r) || unicode.IsSpace(r) {
				t.Fatalf("n=%d invalid char %q in %q", n, r, id)
			}
		}
	}
}

func TestGenShortAndLongID(t *testing.T) {
	shortID, err := GenShortID()
	if err != nil || len(shortID) != ShortIDLength {
		t.Fatalf("short=%q err=%v", shortID, err)
	}
	longID, err := GenLongID()
	if err != nil || len(longID) != LongIDLength {
		t.Fatalf("long=%q err=%v", longID, err)
	}
}

func TestGenUniqueness(t *testing.T) {
	const n = 10_000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id, err := GenShortID()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate %q at %d", id, i)
		}
		seen[id] = struct{}{}
	}
}

func TestGenRandomPassword(t *testing.T) {
	if _, err := GenRandomPassword(7); err == nil {
		t.Fatal("expected invalid length")
	}
	pwd, err := GenRandomPassword(16)
	if err != nil || len(pwd) != 16 {
		t.Fatalf("pwd=%q err=%v", pwd, err)
	}
	var lower, upper, digit, special bool
	for _, r := range pwd {
		switch {
		case strings.ContainsRune(passwordLower, r):
			lower = true
		case strings.ContainsRune(passwordUpper, r):
			upper = true
		case strings.ContainsRune(passwordDigit, r):
			digit = true
		case strings.ContainsRune(passwordSpecial, r):
			special = true
		default:
			t.Fatalf("invalid char %q in %q", r, pwd)
		}
	}
	if !lower || !upper || !digit || !special {
		t.Fatalf("missing class in %q", pwd)
	}
}
