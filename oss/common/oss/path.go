package oss

import (
	"bytes"
	"fmt"
	"io"
)

// toReadSeeker 把业务侧 io.Reader 转成 S3 SDK 需要的 ReadSeeker。
// 大文件应传 *os.File 等 ReadSeeker 并带 ContentLength，避免落入 ReadAll。
func toReadSeeker(r io.Reader, size int64) (io.ReadSeeker, int64, error) {
	// 已能 Seek：不拷贝，只回到起点。优先用调用方给的 size。
	if rs, ok := r.(io.ReadSeeker); ok {
		if size > 0 {
			if _, err := rs.Seek(0, io.SeekStart); err != nil {
				return nil, 0, err
			}
			return rs, size, nil
		}
		n, err := seekerSize(rs)
		return rs, n, err
	}
	// Buffer 已在内存：NewReader 共用底层数组，不再拷一份。
	if buf, ok := r.(*bytes.Buffer); ok {
		return bytes.NewReader(buf.Bytes()), int64(buf.Len()), nil
	}
	// 一次性 Reader（管道、HTTP Body 等）：整份读进内存。大文件可能 OOM。
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, 0, fmt.Errorf("oss: read body: %w", err)
	}
	return bytes.NewReader(data), int64(len(data)), nil
}

// seekerSize 用 SeekEnd 量长度，再 SeekStart 以便后续上传。
func seekerSize(rs io.ReadSeeker) (int64, error) {
	end, err := rs.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = rs.Seek(0, io.SeekStart)
	return end, err
}
