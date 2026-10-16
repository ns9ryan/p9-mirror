package sniff

import (
	"bytes"
	"testing"
)

func TestHeaderMatchesExt(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 1}
	jpeg := []byte{0xff, 0xd8, 0xff, 0xe0}
	zip := []byte{'P', 'K', 0x03, 0x04}
	exe := []byte{'M', 'Z', 0x90, 0x00}
	csv := []byte("id,name\n1,a\n")
	mp4 := make([]byte, 12)
	copy(mp4[4:], "ftypisom")

	cases := []struct {
		ext    string
		header []byte
		ok     bool
	}{
		{".png", png, true},
		{".jpg", jpeg, true},
		{".jpeg", jpeg, true},
		{".zip", zip, true},
		{".xlsx", zip, true},
		{".mp4", mp4, true},
		{".csv", csv, true},
		{".png", jpeg, false},
		{".png", exe, false},
		{".jpg", png, false},
		{".txt", exe, false},
		{".txt", csv, true},
		{"", csv, true},
	}
	for _, tc := range cases {
		err := HeaderMatchesExt(tc.ext, tc.header)
		if tc.ok && err != nil {
			t.Fatalf("%s: %v", tc.ext, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s should reject", tc.ext)
		}
	}
}

func TestPeekHeaderRewinds(t *testing.T) {
	raw := append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, bytes.Repeat([]byte{1}, 20)...)
	rs := bytes.NewReader(raw)
	got, err := PeekHeader(rs, HeaderSize)
	if err != nil {
		t.Fatal(err)
	}
	if sniffKindOf(got) != kindPNG {
		t.Fatalf("kind %q", sniffKindOf(got))
	}
	buf := make([]byte, 8)
	n, err := rs.Read(buf)
	if err != nil || n != 8 || !bytes.Equal(buf, raw[:8]) {
		t.Fatalf("seek rewind failed n=%d err=%v buf=%q", n, err, buf)
	}
}

func TestSniffKindOf(t *testing.T) {
	if sniffKindOf(nil) != kindUnknown {
		t.Fatal("empty")
	}
	if sniffKindOf([]byte("GIF89a....")) != kindGIF {
		t.Fatal("gif")
	}
	if sniffKindOf([]byte("RIFF1234WEBP")) != kindWEBP {
		t.Fatal("webp")
	}
	if err := HeaderMatchesExt(".gif", []byte("GIF87a")); err != nil {
		t.Fatal(err)
	}
}
