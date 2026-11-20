package member

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/mail"
	"regexp"
	"strconv"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/modules/fieldcrypto"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

const (
	ContactBindingCodeSMS   int32 = 1
	ContactBindingCodeEmail int32 = 2
)

var internationalContactPhone = regexp.MustCompile("^[1-9][0-9]{6,14}$")
var contactVerificationCode = regexp.MustCompile("^[0-9]{6}$")

func normalizeContactBindingTarget(channel int32, target string) (string, error) {
	switch channel {
	case ContactBindingCodeSMS:
		target = fieldcrypto.NormalizePhone(target)
		if internationalContactPhone.MatchString(target) {
			return target, nil
		}
	case ContactBindingCodeEmail:
		target = fieldcrypto.NormalizeEmail(target)
		if address, err := mail.ParseAddress(target); err == nil && address.Address == target && len(target) <= 254 {
			return target, nil
		}
	}
	return "", xerr.BadRequest("common.invalidParam")
}

func contactBindingDigest(cryptor *fieldcrypto.Cryptor, scope, value string) ([]byte, error) {
	digest, err := cryptor.LookupHash("member-contact-binding:" + scope + ":" + value)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(digest)
}

func contactBindingCodeDigest(cryptor *fieldcrypto.Cryptor, operatorCode string, recordID int64, code string) ([]byte, error) {
	return contactBindingDigest(cryptor, "code", operatorCode+":"+strconv.FormatInt(recordID, 10)+":"+code)
}

type PendingContactBindingCode struct {
	RecordID  int64
	Target    string
	Code      string
	ExpiresAt time.Time
}

func (service *Service) CreateContactBindingCode(ctx context.Context, memberID int64, channel int32, target string, cryptor *fieldcrypto.Cryptor) (PendingContactBindingCode, error) {
	var pending PendingContactBindingCode
	target, err := normalizeContactBindingTarget(channel, target)
	if err != nil {
		return pending, err
	}
	operatorCode := ctxdata.OperatorCodeFromCtx(ctx)
	if operatorCode == "" {
		return pending, xerr.Unauthorized("common.unauthorized")
	}
	targetDigest, err := contactBindingDigest(cryptor, "target", target)
	if err != nil {
		return pending, err
	}
	randomCode, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return pending, err
	}
	pending.Target = target
	pending.Code = fmt.Sprintf("%06d", randomCode.Int64())
	err = service.withMemberUpdateTx(ctx, memberID, func(transaction *ent.Tx) error {
		driver := transaction.DialectDriver()
		lockKey := fmt.Sprintf("member-contact-binding:%s:%d:%d", operatorCode, memberID, channel)
		if err := driver.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", []any{lockKey}, nil); err != nil {
			return err
		}
		var rows sql.Rows
		if err := driver.Query(ctx, `SELECT nextval(pg_get_serial_sequence('member_contact_binding_code', 'id'))
            WHERE NOT EXISTS (SELECT 1 FROM member_contact_binding_code
            WHERE operator_code = $1 AND member_id = $2 AND channel = $3
            AND created_at > clock_timestamp() - interval '60 seconds')`, []any{operatorCode, memberID, channel}, &rows); err != nil {
			return err
		}
		found := rows.Next()
		if found {
			err = rows.Scan(&pending.RecordID)
		} else {
			err = rows.Err()
		}
		rows.Close()
		if err != nil {
			return err
		}
		if !found {
			return xerr.Conflict("member.contactBindingCodeTooFrequent")
		}
		codeDigest, err := contactBindingCodeDigest(cryptor, operatorCode, pending.RecordID, pending.Code)
		if err != nil {
			return err
		}
		if err := driver.Exec(ctx, `UPDATE member_contact_binding_code SET status = 3
            WHERE operator_code = $1 AND member_id = $2 AND channel = $3 AND status IN (0, 1)`,
			[]any{operatorCode, memberID, channel}, nil); err != nil {
			return err
		}
		pending.ExpiresAt = time.Now().UTC().Truncate(time.Millisecond).Add(5 * time.Minute)
		return driver.Exec(ctx, `INSERT INTO member_contact_binding_code
            (id, operator_code, member_id, channel, target_hash, code_hash, verification_code, expires_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			[]any{pending.RecordID, operatorCode, memberID, channel, targetDigest, codeDigest, pending.Code, pending.ExpiresAt}, nil)
	})
	return pending, err
}

func (service *Service) FinishContactBindingCodeDelivery(ctx context.Context, recordID int64, accepted bool) error {
	var result sql.Result
	err := service.contactBindingCodeDriver.Exec(ctx, `UPDATE member_contact_binding_code
        SET status = CASE WHEN $1 THEN 1 ELSE 4 END,
            sent_at = CASE WHEN $1 THEN clock_timestamp() ELSE NULL END
        WHERE operator_code = $2 AND id = $3 AND status = 0`,
		[]any{accepted, ctxdata.OperatorCodeFromCtx(ctx), recordID}, &result)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err == nil && updated != 1 {
		return fmt.Errorf("contact binding code delivery record is no longer pending")
	}
	return err
}

func verifyContactBindingCode(ctx context.Context, driver dialect.Driver, memberID int64, channel int32, target string, proof *memberpb.ContactBindingCodeProof, cryptor *fieldcrypto.Cryptor) (bool, error) {
	if proof.GetRecordId() <= 0 || !contactVerificationCode.MatchString(proof.GetVerificationCode()) {
		return false, nil
	}
	operatorCode := ctxdata.OperatorCodeFromCtx(ctx)
	targetDigest, err := contactBindingDigest(cryptor, "target", target)
	if err != nil {
		return false, err
	}
	codeDigest, err := contactBindingCodeDigest(cryptor, operatorCode, proof.GetRecordId(), proof.GetVerificationCode())
	if err != nil {
		return false, err
	}
	// UPDATE locks the record until the surrounding contact transaction commits.
	var rows sql.Rows
	err = driver.Query(ctx, `UPDATE member_contact_binding_code
        SET failed_attempts = failed_attempts + CASE WHEN code_hash = $6 THEN 0 ELSE 1 END
        WHERE operator_code = $1 AND member_id = $2 AND channel = $3 AND id = $4
        AND target_hash = $5 AND status = 1 AND expires_at > clock_timestamp() AND failed_attempts < 5
        RETURNING code_hash = $6`, []any{operatorCode, memberID, channel, proof.GetRecordId(), targetDigest, codeDigest}, &rows)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return false, rows.Err()
	}
	var matches bool
	err = rows.Scan(&matches)
	return matches, err
}
