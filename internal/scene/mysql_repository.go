// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scene

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// mysqlRepository implements ExtraParamRepository backed by MySQL/PostgreSQL.
type mysqlRepository struct {
	db *sql.DB
}

// NewMySQLRepository returns an ExtraParamRepository backed by the given *sql.DB.
func NewMySQLRepository(db *sql.DB) ExtraParamRepository {
	return &mysqlRepository{db: db}
}

const selectParamCols = `
	id, scene_code, param_key, param_type, required,
	COALESCE(default_val, ''), COALESCE(description, ''),
	status, version, created_at, updated_at`

func scanSpec(row interface{ Scan(...any) error }) (*ExtraParamSpec, error) {
	var s ExtraParamSpec
	var required int8
	err := row.Scan(
		&s.ID, &s.SceneCode, &s.ParamKey, &s.ParamType, &required,
		&s.DefaultVal, &s.Description,
		&s.Status, &s.Version, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.Required = required == 1
	return &s, nil
}

func scanSpecs(rows *sql.Rows) ([]*ExtraParamSpec, error) {
	var out []*ExtraParamSpec
	for rows.Next() {
		s, err := scanSpec(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *mysqlRepository) ListByScene(ctx context.Context, sceneCode string) ([]*ExtraParamSpec, error) {
	q := `SELECT ` + selectParamCols + `
		FROM scene_extra_params
		WHERE scene_code = ? AND status = 1
		ORDER BY param_key ASC`
	rows, err := r.db.QueryContext(ctx, q, sceneCode)
	if err != nil {
		return nil, fmt.Errorf("scene.ListByScene(%s): %w", sceneCode, err)
	}
	defer rows.Close() //nolint:errcheck
	return scanSpecs(rows)
}

func (r *mysqlRepository) ListUpdatedSince(ctx context.Context, since time.Time) ([]*ExtraParamSpec, error) {
	q := `SELECT ` + selectParamCols + `
		FROM scene_extra_params
		WHERE updated_at > ?
		ORDER BY updated_at ASC`
	rows, err := r.db.QueryContext(ctx, q, since)
	if err != nil {
		return nil, fmt.Errorf("scene.ListUpdatedSince: %w", err)
	}
	defer rows.Close() //nolint:errcheck
	return scanSpecs(rows)
}

func (r *mysqlRepository) GetByKey(ctx context.Context, sceneCode, paramKey string) (*ExtraParamSpec, error) {
	q := `SELECT ` + selectParamCols + `
		FROM scene_extra_params
		WHERE scene_code = ? AND param_key = ?`
	row := r.db.QueryRowContext(ctx, q, sceneCode, paramKey)
	s, err := scanSpec(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("scene extra param %s/%s not found", sceneCode, paramKey)
	}
	if err != nil {
		return nil, fmt.Errorf("scene.GetByKey(%s/%s): %w", sceneCode, paramKey, err)
	}
	return s, nil
}

func (r *mysqlRepository) Upsert(ctx context.Context, spec *ExtraParamSpec) error {
	required := int8(0)
	if spec.Required {
		required = 1
	}
	var defVal interface{}
	if spec.DefaultVal != "" {
		defVal = spec.DefaultVal
	}
	var desc interface{}
	if spec.Description != "" {
		desc = spec.Description
	}

	q := `
	INSERT INTO scene_extra_params
		(scene_code, param_key, param_type, required, default_val, description, status)
	VALUES (?, ?, ?, ?, ?, ?, 1)
	ON DUPLICATE KEY UPDATE
		param_type   = VALUES(param_type),
		required     = VALUES(required),
		default_val  = VALUES(default_val),
		description  = VALUES(description),
		status       = 1,
		version      = version + 1`

	_, err := r.db.ExecContext(ctx, q,
		spec.SceneCode, spec.ParamKey, spec.ParamType,
		required, defVal, desc,
	)
	if err != nil {
		return fmt.Errorf("scene.Upsert(%s/%s): %w", spec.SceneCode, spec.ParamKey, err)
	}
	return nil
}

func (r *mysqlRepository) Delete(ctx context.Context, sceneCode, paramKey string) error {
	q := `UPDATE scene_extra_params SET status = 0, version = version + 1
		  WHERE scene_code = ? AND param_key = ?`
	_, err := r.db.ExecContext(ctx, q, sceneCode, paramKey)
	if err != nil {
		return fmt.Errorf("scene.Delete(%s/%s): %w", sceneCode, paramKey, err)
	}
	return nil
}
