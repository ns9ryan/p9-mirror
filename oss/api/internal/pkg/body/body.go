package body

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

// ParseForm 解析 multipart。手写 Content-Type（没有真实 boundary / 空 body）时给出明确错误。
func ParseForm(r *http.Request, maxMemory int64) error {
	if r.ContentLength == 0 {
		return errors.New("empty body: uncheck Content-Type in headers, send form-data file")
	}
	ct := r.Header.Get("Content-Type")
	media, params, err := mime.ParseMediaType(ct)
	if err != nil || !strings.HasPrefix(media, "multipart/") {
		return fmt.Errorf("content-type must be multipart/form-data, got %q", ct)
	}
	boundary := strings.TrimSpace(params["boundary"])
	if boundary == "" || strings.Contains(strings.ToLower(boundary), "calculated") {
		return errors.New("multipart boundary missing: uncheck Content-Type in headers")
	}
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if IsTooLarge(err) {
			return err
		}
		return fmt.Errorf("invalid multipart: %v", err)
	}
	return nil
}

// File 取名为 name 的文件字段。
func File(r *http.Request, name string) (multipart.File, *multipart.FileHeader, error) {
	src, hdr, err := r.FormFile(name)
	if err == nil {
		return src, hdr, nil
	}
	return nil, nil, fmt.Errorf("no file provided (%s)", fileHint(r))
}

func fileHint(r *http.Request) string {
	var files, values []string
	if r.MultipartForm != nil {
		for k := range r.MultipartForm.File {
			files = append(files, k)
		}
		for k := range r.MultipartForm.Value {
			values = append(values, k)
		}
	}
	return fmt.Sprintf("content-length=%d files=%v fields=%v", r.ContentLength, files, values)
}

type Prepared struct {
	Body    io.ReadSeeker
	Size    int64
	Cleanup func()
}

func (p Prepared) Close() {
	if p.Cleanup != nil {
		p.Cleanup()
	}
}

// Prepare 把 multipart 文件变成 ReadSeeker；不能 Seek 时落到临时文件。
func Prepare(src multipart.File, headerSize, maxSize int64) (Prepared, error) {
	if rs, ok := src.(io.ReadSeeker); ok {
		size := headerSize
		if size <= 0 {
			n, err := seekerSize(rs)
			if err != nil {
				return Prepared{}, err
			}
			size = n
		} else if _, err := rs.Seek(0, io.SeekStart); err != nil {
			return Prepared{}, err
		}
		return Prepared{Body: rs, Size: size}, nil
	}

	tmp, err := os.CreateTemp("", "oss-upload-*")
	if err != nil {
		return Prepared{}, err
	}
	cleanup := func() {
		name := tmp.Name()
		_ = tmp.Close()
		_ = os.Remove(name)
	}
	n, err := io.Copy(tmp, io.LimitReader(src, maxSize+1))
	if err != nil {
		cleanup()
		return Prepared{}, err
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		cleanup()
		return Prepared{}, err
	}
	return Prepared{Body: tmp, Size: n, Cleanup: cleanup}, nil
}

func seekerSize(rs io.ReadSeeker) (int64, error) {
	end, err := rs.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = rs.Seek(0, io.SeekStart)
	return end, err
}

// IsTooLarge 是否因请求体超过 MaxBytesReader 上限而失败。
func IsTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}
