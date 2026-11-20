package member

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
)

type ContactBindingCodeFilters struct {
	RecordID      int64
	MemberID      int64
	Channel       int32
	Status        *int32
	CreatedFrom   int64
	CreatedBefore int64
}

type ContactBindingCodeRecord struct {
	VerificationCode string
	ID               int64
	MemberID         int64
	Channel          int32
	Status           int32
	FailedAttempts   int32
	CreatedAt        time.Time
	SentAt           *time.Time
	ExpiresAt        time.Time
	UsedAt           *time.Time
}

// ListContactBindingCodes returns records for the authenticated operator only.
func (service *Service) ListContactBindingCodes(ctx context.Context, operatorCode string, filters ContactBindingCodeFilters, offset, limit int) ([]ContactBindingCodeRecord, int64, error) {
	if operatorCode == "" {
		return nil, 0, fmt.Errorf("missing operator code")
	}
	conditions := []string{"operator_code = $1"}
	arguments := []any{operatorCode}
	addFilter := func(column, comparison string, value any) {
		arguments = append(arguments, value)
		conditions = append(conditions, fmt.Sprintf("%s %s $%d", column, comparison, len(arguments)))
	}
	if filters.RecordID > 0 {
		addFilter("id", "=", filters.RecordID)
	}
	if filters.MemberID > 0 {
		addFilter("member_id", "=", filters.MemberID)
	}
	if filters.Channel > 0 {
		addFilter("channel", "=", filters.Channel)
	}
	if filters.Status != nil {
		addFilter("status", "=", *filters.Status)
	}
	if filters.CreatedFrom > 0 {
		addFilter("created_at", ">=", time.UnixMilli(filters.CreatedFrom))
	}
	if filters.CreatedBefore > 0 {
		addFilter("created_at", "<", time.UnixMilli(filters.CreatedBefore))
	}

	// One statement keeps the count and page on the same database snapshot.
	query := `WITH matches AS NOT MATERIALIZED (
        SELECT id, member_id, channel, status, failed_attempts, created_at, sent_at, expires_at, used_at, verification_code
        FROM member_contact_binding_code WHERE ` + strings.Join(conditions, " AND ") + `
    ), page AS (
        SELECT * FROM matches ORDER BY created_at DESC, id DESC`
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(arguments)+1, len(arguments)+2) + `
    )
    SELECT summary.total, page.id, page.member_id, page.channel, page.status,
           page.failed_attempts, page.created_at, page.sent_at, page.expires_at, page.used_at, CASE WHEN page.status IN (2, 3, 4) OR page.expires_at <= statement_timestamp() OR page.failed_attempts >= 5
                THEN COALESCE(page.verification_code, '') ELSE '' END
    FROM (SELECT count(*) AS total FROM matches) AS summary
    LEFT JOIN page ON true
    ORDER BY page.created_at DESC, page.id DESC`
	arguments = append(arguments, limit, offset)

	var rows sql.Rows
	if err := service.contactBindingCodeDriver.Query(ctx, query, arguments, &rows); err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]ContactBindingCodeRecord, 0)
	var total int64
	for rows.Next() {
		var recordID *int64
		var verificationCode string
		var memberID *int64
		var channel, status, failedAttempts *int32
		var createdAt, sentAt, expiresAt, usedAt *time.Time
		if err := rows.Scan(&total, &recordID, &memberID, &channel, &status,
			&failedAttempts, &createdAt, &sentAt, &expiresAt, &usedAt, &verificationCode); err != nil {
			return nil, 0, err
		}
		if recordID != nil {
			records = append(records, ContactBindingCodeRecord{
				VerificationCode: verificationCode, ID: *recordID, MemberID: *memberID, Channel: *channel,
				Status: *status, FailedAttempts: *failedAttempts, CreatedAt: *createdAt,
				SentAt: sentAt, ExpiresAt: *expiresAt, UsedAt: usedAt,
			})
		}
	}
	return records, total, rows.Err()
}
