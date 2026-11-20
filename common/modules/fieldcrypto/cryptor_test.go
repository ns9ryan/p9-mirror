package fieldcrypto

import (
	"bytes"
	"testing"
)

func testKeys(t *testing.T) (enc, mac []byte) {
	t.Helper()
	enc = bytes.Repeat([]byte{0x11}, 32)
	mac = bytes.Repeat([]byte{0x22}, 32)
	return enc, mac
}

func TestNewInvalidKey(t *testing.T) {
	_, err := New([]byte("short"), []byte("mac"))
	if err != ErrInvalidEncKey {
		t.Fatalf("enc: got %v want ErrInvalidEncKey", err)
	}
	_, err = New(bytes.Repeat([]byte{1}, 32), nil)
	if err != ErrEmptyMacKey {
		t.Fatalf("mac: got %v want ErrEmptyMacKey", err)
	}
}

func TestSealOpenRoundTrip(t *testing.T) {
	c, err := New(testKeys(t))
	if err != nil {
		t.Fatal(err)
	}
	const plain = "60-123456789"
	ct, err := c.Seal(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Open(ct)
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("got %q want %q", got, plain)
	}
}

func TestSealNonceRandom(t *testing.T) {
	c, err := New(testKeys(t))
	if err != nil {
		t.Fatal(err)
	}
	a, err := c.Seal("same")
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Seal("same")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("expected different ciphertexts for same plaintext")
	}
}

func TestLookupHashStable(t *testing.T) {
	c, err := New(testKeys(t))
	if err != nil {
		t.Fatal(err)
	}
	h1, err := c.LookupHash("60-123456789")
	if err != nil {
		t.Fatal(err)
	}
	h2, err := c.LookupHash("60-123456789")
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("unstable hash: %s vs %s", h1, h2)
	}
	if len(h1) != 64 {
		t.Fatalf("hash len=%d want 64", len(h1))
	}
}

func TestOpenWrongKey(t *testing.T) {
	c1, err := New(bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 16))
	if err != nil {
		t.Fatal(err)
	}
	ct, err := c1.Seal("secret")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := New(bytes.Repeat([]byte{9}, 32), bytes.Repeat([]byte{2}, 16))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c2.Open(ct); err != ErrInvalidCiphertext {
		t.Fatalf("got %v want ErrInvalidCiphertext", err)
	}
}

func TestOpenInvalidCiphertext(t *testing.T) {
	c, err := New(testKeys(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Open("%%%"); err != ErrInvalidCiphertext {
		t.Fatalf("got %v want ErrInvalidCiphertext", err)
	}
	if _, err := c.Open(""); err != ErrInvalidCiphertext {
		t.Fatalf("empty: got %v want ErrInvalidCiphertext", err)
	}
}

func TestSealEmpty(t *testing.T) {
	c, err := New(testKeys(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Seal(""); err != ErrEmptyPlaintext {
		t.Fatalf("got %v want ErrEmptyPlaintext", err)
	}
	if _, err := c.LookupHash(""); err != ErrEmptyNormalized {
		t.Fatalf("lookup: got %v want ErrEmptyNormalized", err)
	}
}

func TestNormalize(t *testing.T) {
	if got := NormalizeEmail("  Foo@Bar.COM "); got != "foo@bar.com" {
		t.Fatalf("email: %q", got)
	}
	if got := NormalizePhone(" +60 123 456 "); got != "60123456" {
		t.Fatalf("phone: %q", got)
	}
	if got := NormalizePhone("60-123456789"); got != "60-123456789" {
		t.Fatalf("phone keep dash: %q", got)
	}
}
