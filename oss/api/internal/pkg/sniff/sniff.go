package sniff

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// 本文件：按文件头魔数校验后缀，拦截可执行文件。

const HeaderSize = 512

type sniffKind string

const (
	kindUnknown sniffKind = ""
	kindJPEG    sniffKind = "jpeg"
	kindPNG     sniffKind = "png"
	kindGIF     sniffKind = "gif"
	kindWEBP    sniffKind = "webp"
	kindBMP     sniffKind = "bmp"
	kindPDF     sniffKind = "pdf"
	kindZIP     sniffKind = "zip"
	kindMP4     sniffKind = "mp4"
	kindWebM    sniffKind = "webm"
	kindWAV     sniffKind = "wav"
	kindMP3     sniffKind = "mp3"
	kindExec    sniffKind = "exec"
)

var kindExts = map[sniffKind][]string{
	kindJPEG: {".jpg", ".jpeg"},
	kindPNG:  {".png"},
	kindGIF:  {".gif"},
	kindWEBP: {".webp"},
	kindBMP:  {".bmp"},
	kindPDF:  {".pdf"},
	kindZIP:  {".zip", ".xlsx", ".xlsm", ".docx", ".pptx", ".apk", ".jar"},
	kindMP4:  {".mp4", ".m4v", ".m4a", ".mov"},
	kindWebM: {".webm", ".mkv"},
	kindWAV:  {".wav"},
	kindMP3:  {".mp3"},
}

// PeekHeader 读文件头后 Seek 回起点，供魔数校验。
func PeekHeader(rs io.ReadSeeker, n int) ([]byte, error) {
	if n <= 0 {
		return nil, nil
	}
	buf := make([]byte, n)
	got, err := io.ReadFull(rs, buf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		buf = buf[:got]
		err = nil
	}
	if _, seekErr := rs.Seek(0, io.SeekStart); seekErr != nil {
		return nil, seekErr
	}
	return buf, err
}

// HeaderMatchesExt 能识别的文件头必须与后缀一致；可执行文件一律拒绝。
func HeaderMatchesExt(ext string, header []byte) error {
	kind := sniffKindOf(header)
	if kind == kindUnknown {
		return nil
	}
	if kind == kindExec {
		return fmt.Errorf("executable content is not allowed")
	}
	ext = strings.ToLower(ext)
	for _, allowed := range kindExts[kind] {
		if allowed == ext {
			return nil
		}
	}
	return fmt.Errorf("file content does not match extension %s", ext)
}

func sniffKindOf(b []byte) sniffKind {
	switch {
	case bytes.HasPrefix(b, []byte{'M', 'Z'}):
		return kindExec
	case len(b) >= 4 && b[0] == 0x7f && b[1] == 'E' && b[2] == 'L' && b[3] == 'F':
		return kindExec
	case bytes.HasPrefix(b, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return kindPNG
	case len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff:
		return kindJPEG
	case bytes.HasPrefix(b, []byte("GIF87a")), bytes.HasPrefix(b, []byte("GIF89a")):
		return kindGIF
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")):
		return kindWEBP
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WAVE")):
		return kindWAV
	case bytes.HasPrefix(b, []byte("%PDF")):
		return kindPDF
	case bytes.HasPrefix(b, []byte{'P', 'K', 0x03, 0x04}), bytes.HasPrefix(b, []byte{'P', 'K', 0x05, 0x06}):
		return kindZIP
	case len(b) >= 8 && bytes.Equal(b[4:8], []byte("ftyp")):
		return kindMP4
	case len(b) >= 4 && bytes.Equal(b[:4], []byte{0x1a, 0x45, 0xdf, 0xa3}):
		return kindWebM
	case bytes.HasPrefix(b, []byte("ID3")):
		return kindMP3
	case bytes.HasPrefix(b, []byte("BM")):
		return kindBMP
	default:
		return kindUnknown
	}
}
