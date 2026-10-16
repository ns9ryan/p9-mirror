package body

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxJSONBody = 8 << 20

// JSON 用 encoding/json 解 body。不用 go-zero mapping（缺字段、嵌套 map 会误报 invalid request）。
func JSON(r *http.Request, v any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, maxJSONBody+1))
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return errors.New("empty body")
	}
	if len(raw) > maxJSONBody {
		return errors.New("request body too large")
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}
