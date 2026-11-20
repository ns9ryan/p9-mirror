package manage

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xerr"
	contract "oa.98ent.com/p9/operator-crm/pkg/memberexport"
)

const (
	memberExportMemberIDColumnIndex   = 0
	memberExportMemberCodeColumnIndex = 1
	memberExportMemberCodeLength      = 12
	memberExportIntegerBase           = 10
	memberExportMemberIDBitSize       = 64
	memberExportCSVHeaderRowCount     = 1
	memberExportCSVOverflowBytes      = 1
)

// verifyCSV 校验完整 CSV 的固定表头、会员标识、唯一性与实际行数。
// 即使最终不导出会员 ID，也先完成完整文件校验。
func verifyCSV(ctx context.Context, csvFile *os.File, expectedRowCount int64) error {
	validationError := func() error { return xerr.InternalServerError("export CSV validation failed") }
	if expectedRowCount < 0 || expectedRowCount > contract.MaxRows {
		return validationError()
	}
	if _, err := csvFile.Seek(0, io.SeekStart); err != nil {
		return validationError()
	}
	reader := csv.NewReader(io.LimitReader(csvFile, contract.MaxBytes+memberExportCSVOverflowBytes))
	reader.FieldsPerRecord = len(contract.Header)
	header, err := reader.Read()
	if err != nil || !slices.Equal(header, contract.Header) {
		return validationError()
	}
	seenMemberIDs := make(map[int64]struct{}, expectedRowCount)
	var rowCount int64
	for {
		if err = ctx.Err(); err != nil {
			return exportHTTPError(ctx, err)
		}
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return validationError()
		}
		memberID, parseErr := strconv.ParseInt(strings.TrimPrefix(row[memberExportMemberIDColumnIndex], "'"), memberExportIntegerBase, memberExportMemberIDBitSize)
		if parseErr != nil || memberID <= 0 {
			return validationError()
		}
		memberCode := strings.TrimPrefix(row[memberExportMemberCodeColumnIndex], "'")
		if len(memberCode) != memberExportMemberCodeLength {
			return validationError()
		}
		for _, digit := range memberCode {
			if digit < '0' || digit > '9' {
				return validationError()
			}
		}
		if _, ok := seenMemberIDs[memberID]; ok {
			return validationError()
		}
		seenMemberIDs[memberID] = struct{}{}
		rowCount++
		if rowCount > contract.MaxRows {
			return validationError()
		}
	}
	if rowCount != expectedRowCount {
		return validationError()
	}
	_, err = csvFile.Seek(0, io.SeekStart)
	if err != nil {
		return validationError()
	}
	return nil
}

// projectCSV 按请求列顺序生成新文件；输入必须已经通过完整 CSV 校验。
func projectCSV(ctx context.Context, sourceFile *os.File, columnIndices []int, expectedRowCount int64) (projectedFile *os.File, err error) {
	if _, err = sourceFile.Seek(0, io.SeekStart); err != nil {
		return nil, xerr.InternalServerError("CSV projection seek failed")
	}
	projectedFile, err = os.CreateTemp("", "member-export-columns-*.csv")
	if err != nil {
		return nil, xerr.InternalServerError("CSV projection temporary file failed")
	}
	temporaryFile := projectedFile
	// 投影中途失败时关闭并移除输出文件，避免残留不完整 CSV。
	defer func() {
		if err != nil {
			temporaryFile.Close()
			if cleanupErr := os.Remove(temporaryFile.Name()); cleanupErr != nil {
				logx.WithContext(ctx).Error("member export projection cleanup failed")
			}
		}
	}()
	reader := csv.NewReader(sourceFile)
	reader.FieldsPerRecord = len(contract.Header)
	writer := csv.NewWriter(projectedFile)
	writer.UseCRLF = true
	var recordCount int64
	for {
		if ctx.Err() != nil {
			return nil, exportHTTPError(ctx, ctx.Err())
		}
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, xerr.InternalServerError("CSV projection read failed")
		}
		row := make([]string, len(columnIndices))
		for outputColumnIndex, sourceColumnIndex := range columnIndices {
			if sourceColumnIndex < 0 || sourceColumnIndex >= len(record) {
				return nil, xerr.InternalServerError("CSV projection index invalid")
			}
			row[outputColumnIndex] = record[sourceColumnIndex]
		}
		if writeErr := writer.Write(row); writeErr != nil {
			return nil, xerr.InternalServerError("CSV projection write failed")
		}
		recordCount++
	}
	writer.Flush()
	if writer.Error() != nil || recordCount != expectedRowCount+memberExportCSVHeaderRowCount {
		return nil, xerr.InternalServerError("CSV projection count or write failed")
	}
	if _, err = projectedFile.Seek(0, io.SeekStart); err != nil {
		return nil, xerr.InternalServerError("CSV projection seek failed")
	}
	return projectedFile, nil
}
