package memberlogic

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"oa.98ent.com/p9/common/modules/fieldcrypto"
	"oa.98ent.com/p9/operator-crm/ent"
	contract "oa.98ent.com/p9/operator-crm/pkg/memberexport"
)

func contactValue(cipher *string, permitted bool, cryptor *fieldcrypto.Cryptor) (string, error) {
	if cipher == nil {
		return contract.Missing, nil
	}
	// Without permission a constant mask avoids disclosing contact length or suffix.
	if !permitted {
		return "[REDACTED]", nil
	}
	if cryptor == nil {
		return "", fmt.Errorf("contact cryptor unavailable")
	}
	plain, err := cryptor.Open(*cipher)
	if err != nil || plain == "" {
		return "", fmt.Errorf("contact decryption failed")
	}
	return plain, nil
}

func exportRow(m *ent.Member, policy contract.Policy, cryptor *fieldcrypto.Cryptor) ([]string, error) {
	if m == nil || m.ID <= 0 || len(m.Code) != 12 || m.CreatedAt.IsZero() || m.Edges.Auth == nil || m.Edges.Agent == nil || m.Edges.AutoLevel == nil || m.Edges.VipLevel == nil || m.AgentID <= 0 || m.Edges.Agent.Account == "" || m.Edges.AutoLevel.Name == "" {
		return nil, fmt.Errorf("required export association missing")
	}
	for _, c := range m.Code {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("invalid member code")
		}
	}
	phone, err := contactValue(m.Edges.Auth.PhoneCiphertext, policy.Phone, cryptor)
	if err != nil {
		return nil, err
	}
	email, err := contactValue(m.Edges.Auth.EmailCiphertext, policy.Email, cryptor)
	if err != nil {
		return nil, err
	}
	optional := func(p *string) string {
		if p == nil {
			return contract.Missing
		}
		return *p
	}
	id := func(v int64) string { return strconv.FormatInt(v, 10) }
	optionalID := func(p *int64) string {
		if p == nil {
			return contract.Missing
		}
		return id(*p)
	}
	nickname := contract.Missing
	if m.Edges.Profile != nil {
		nickname = m.Edges.Profile.Nickname
	}
	fixed := make([]string, 0, len(m.Edges.FixedLevelAssignments))
	for _, a := range m.Edges.FixedLevelAssignments {
		if a.Edges.FixedLevel == nil {
			return nil, fmt.Errorf("fixed level missing")
		}
		fixed = append(fixed, a.Edges.FixedLevel.Name)
	}
	sort.Strings(fixed)
	deposit, withdraw := contract.Missing, contract.Missing
	if s := m.Edges.Stats; s != nil {
		if s.TotalDeposit < 0 || s.TotalWithdrawal < 0 {
			return nil, fmt.Errorf("negative funding aggregate")
		}
		deposit, withdraw = id(s.TotalDeposit), id(s.TotalWithdrawal)
	}
	login := contract.Missing
	if m.LastLoginAt != nil {
		login = m.LastLoginAt.UTC().Format(time.RFC3339Nano)
	}
	row := []string{id(m.ID), m.Code, optional(m.Edges.Auth.Username), nickname, phone, email, id(m.Status), id(m.AgentID), optionalID(m.AgentAccountID), optionalID(m.InviterMemberID), m.Edges.Agent.Account, m.Edges.AutoLevel.Name, id(m.Edges.VipLevel.Level), strings.Join(fixed, ";"), m.CreatedAt.UTC().Format(time.RFC3339Nano), login, deposit, withdraw, contract.Unavailable, contract.Unavailable, contract.Unavailable}
	for i := range row {
		row[i] = contract.Cell(row[i], i == 0 || i == 1 || i == 4 || i == 7 || i == 8 || i == 9)
	}
	return row, nil
}
