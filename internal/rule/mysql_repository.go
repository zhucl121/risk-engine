// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// mysqlRepository implements RuleRepository backed by a MySQL/PostgreSQL database.
// It uses the standard database/sql package so it works with any driver.
type mysqlRepository struct {
	db *sql.DB
}

// NewMySQLRepository returns a RuleRepository backed by the given *sql.DB.
// The db must already be open and have max_open_conns / max_idle_conns configured.
func NewMySQLRepository(db *sql.DB) RuleRepository {
	return &mysqlRepository{db: db}
}

const selectCols = `
	id, rule_key, name, group_name, scene_code, priority,
	condition_dsl, COALESCE(condition_ast, ''), action_decision,
	COALESCE(action_risk_code, ''), action_score, status, version,
	created_at, updated_at`

func scanRecord(row interface {
	Scan(...any) error
}) (*RuleRecord, error) {
	var r RuleRecord
	var condAST string
	err := row.Scan(
		&r.ID, &r.RuleKey, &r.Name, &r.GroupName, &r.SceneCode, &r.Priority,
		&r.ConditionDSL, &condAST, &r.ActionDecision,
		&r.ActionRiskCode, &r.ActionScore, &r.Status, &r.Version,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if condAST != "" {
		r.ConditionAST = []byte(condAST)
	}
	return &r, nil
}

func (m *mysqlRepository) ListActive(ctx context.Context, sceneCode string) ([]*RuleRecord, error) {
	q := `SELECT ` + selectCols + ` FROM risk_rules WHERE status = 1`
	args := []any{}
	if sceneCode != "" {
		q += ` AND scene_code = ?`
		args = append(args, sceneCode)
	}
	q += ` ORDER BY priority DESC`

	rows, err := m.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("rule.ListActive: %w", err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func (m *mysqlRepository) ListUpdatedSince(ctx context.Context, since time.Time) ([]*RuleRecord, error) {
	q := `SELECT ` + selectCols + ` FROM risk_rules WHERE updated_at > ? ORDER BY updated_at ASC`
	rows, err := m.db.QueryContext(ctx, q, since)
	if err != nil {
		return nil, fmt.Errorf("rule.ListUpdatedSince: %w", err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func scanRows(rows *sql.Rows) ([]*RuleRecord, error) {
	var out []*RuleRecord
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (m *mysqlRepository) GetByID(ctx context.Context, id int64) (*RuleRecord, error) {
	q := `SELECT ` + selectCols + ` FROM risk_rules WHERE id = ?`
	row := m.db.QueryRowContext(ctx, q, id)
	r, err := scanRecord(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("rule.GetByID: %w", err)
	}
	return r, nil
}

func (m *mysqlRepository) Create(ctx context.Context, r *RuleRecord) (int64, error) {
	q := `
	INSERT INTO risk_rules
		(rule_key, name, group_name, scene_code, priority,
		 condition_dsl, condition_ast, action_decision, action_risk_code,
		 action_score, status, version)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 1)`

	var condAST interface{}
	if len(r.ConditionAST) > 0 {
		condAST = string(r.ConditionAST)
	}

	res, err := m.db.ExecContext(ctx, q,
		r.RuleKey, r.Name, r.GroupName, r.SceneCode, r.Priority,
		r.ConditionDSL, condAST, r.ActionDecision, r.ActionRiskCode,
		r.ActionScore,
	)
	if err != nil {
		return 0, fmt.Errorf("rule.Create: %w", err)
	}
	return res.LastInsertId()
}

func (m *mysqlRepository) Update(ctx context.Context, r *RuleRecord) error {
	q := `
	UPDATE risk_rules SET
		name = ?, group_name = ?, scene_code = ?, priority = ?,
		condition_dsl = ?, condition_ast = ?, action_decision = ?,
		action_risk_code = ?, action_score = ?, status = ?,
		version = version + 1, updated_at = NOW()
	WHERE id = ? AND version = ?`

	var condAST interface{}
	if len(r.ConditionAST) > 0 {
		condAST = string(r.ConditionAST)
	}

	res, err := m.db.ExecContext(ctx, q,
		r.Name, r.GroupName, r.SceneCode, r.Priority,
		r.ConditionDSL, condAST, r.ActionDecision, r.ActionRiskCode,
		r.ActionScore, r.Status, r.ID, r.Version,
	)
	if err != nil {
		return fmt.Errorf("rule.Update: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rule.Update rows: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("rule.Update: optimistic lock conflict or rule %d not found", r.ID)
	}
	return nil
}

func (m *mysqlRepository) SoftDelete(ctx context.Context, id int64) error {
	q := `UPDATE risk_rules SET status = 0, updated_at = NOW() WHERE id = ?`
	_, err := m.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("rule.SoftDelete(%d): %w", id, err)
	}
	return nil
}
