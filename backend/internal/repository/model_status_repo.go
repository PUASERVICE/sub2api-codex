package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type modelStatusRepository struct {
	db *sql.DB
}

func NewModelStatusRepository(db *sql.DB) service.ModelStatusTargetRepository {
	return &modelStatusRepository{db: db}
}

func (r *modelStatusRepository) Create(ctx context.Context, target *service.ModelStatusTarget) (*service.ModelStatusTarget, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO model_status_targets (
			name, account_id, model_id, check_interval_seconds, timeout_seconds, enabled,
			latest_status, latest_latency_ms, latest_error_message, latest_response_text,
			consecutive_failures, last_checked_at, last_success_at, last_failure_at, next_check_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			NOW(), NOW()
		)
		RETURNING id
	`,
		target.Name,
		target.AccountID,
		target.ModelID,
		target.CheckIntervalSeconds,
		target.TimeoutSeconds,
		target.Enabled,
		target.LatestStatus,
		nullableInt64(target.LatestLatencyMs),
		target.LatestErrorMessage,
		target.LatestResponseText,
		target.ConsecutiveFailures,
		target.LastCheckedAt,
		target.LastSuccessAt,
		target.LastFailureAt,
		target.NextCheckAt,
	)
	var id int64
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *modelStatusRepository) GetByID(ctx context.Context, id int64) (*service.ModelStatusTarget, error) {
	row := r.db.QueryRowContext(ctx, modelStatusTargetSelectSQL+` WHERE t.id = $1`, id)
	target, err := scanModelStatusTarget(row)
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (r *modelStatusRepository) List(ctx context.Context, includeDisabled bool) ([]*service.ModelStatusTarget, error) {
	var (
		rows *sql.Rows
		err  error
	)
	query := modelStatusTargetSelectSQL
	if includeDisabled {
		query += ` ORDER BY t.enabled DESC, CASE t.latest_status WHEN 'failed' THEN 0 WHEN 'unknown' THEN 1 ELSE 2 END, t.updated_at DESC`
		rows, err = r.db.QueryContext(ctx, query)
	} else {
		query += ` WHERE t.enabled = true ORDER BY CASE t.latest_status WHEN 'failed' THEN 0 WHEN 'unknown' THEN 1 ELSE 2 END, t.updated_at DESC`
		rows, err = r.db.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanModelStatusTargets(rows)
}

func (r *modelStatusRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]*service.ModelStatusTarget, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, modelStatusTargetSelectSQL+`
		WHERE t.enabled = true
		  AND (t.next_check_at IS NULL OR t.next_check_at <= $1)
		ORDER BY COALESCE(t.next_check_at, t.created_at) ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanModelStatusTargets(rows)
}

func (r *modelStatusRepository) Update(ctx context.Context, target *service.ModelStatusTarget) (*service.ModelStatusTarget, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE model_status_targets
		SET
			name = $2,
			account_id = $3,
			model_id = $4,
			check_interval_seconds = $5,
			timeout_seconds = $6,
			enabled = $7,
			next_check_at = $8,
			updated_at = NOW()
		WHERE id = $1
	`,
		target.ID,
		target.Name,
		target.AccountID,
		target.ModelID,
		target.CheckIntervalSeconds,
		target.TimeoutSeconds,
		target.Enabled,
		target.NextCheckAt,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, target.ID)
}

func (r *modelStatusRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM model_status_targets WHERE id = $1`, id)
	return err
}

func (r *modelStatusRepository) RecordCheck(ctx context.Context, targetID int64, check *service.ModelStatusCheck, nextCheckAt *time.Time) (*service.ModelStatusTarget, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO model_status_checks (
			target_id, status, latency_ms, error_message, response_text,
			started_at, finished_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`,
		targetID,
		normalizeText(check.Status),
		nullableInt64(check.LatencyMs),
		check.ErrorMessage,
		check.ResponseText,
		check.StartedAt,
		check.FinishedAt,
	); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	consecutiveFailureSQL := `CASE WHEN $2 = 'success' THEN 0 ELSE consecutive_failures + 1 END`
	if _, err := tx.ExecContext(ctx, `
		UPDATE model_status_targets
		SET
			latest_status = $2,
			latest_latency_ms = $3,
			latest_error_message = $4,
			latest_response_text = $5,
			consecutive_failures = `+consecutiveFailureSQL+`,
			last_checked_at = $6,
			last_success_at = CASE WHEN $2 = 'success' THEN $6 ELSE last_success_at END,
			last_failure_at = CASE WHEN $2 = 'failed' THEN $6 ELSE last_failure_at END,
			next_check_at = $7,
			updated_at = NOW()
		WHERE id = $1
	`,
		targetID,
		strings.TrimSpace(check.Status),
		nullableInt64(check.LatencyMs),
		check.ErrorMessage,
		check.ResponseText,
		now,
		nextCheckAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, targetID)
}

func (r *modelStatusRepository) ListChecks(ctx context.Context, targetID int64, limit int) ([]*service.ModelStatusCheck, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, target_id, status, latency_ms, error_message, response_text, started_at, finished_at, created_at
		FROM model_status_checks
		WHERE target_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var checks []*service.ModelStatusCheck
	for rows.Next() {
		check, err := scanModelStatusCheck(rows)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	return checks, rows.Err()
}

func (r *modelStatusRepository) GetOverview(ctx context.Context, includeDisabled bool) (*service.ModelStatusOverview, error) {
	clauses := []string{}
	if !includeDisabled {
		clauses = append(clauses, "enabled = true")
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	row := r.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT
			COUNT(*) AS total_targets,
			COUNT(*) FILTER (WHERE enabled = true) AS enabled_targets,
			COUNT(*) FILTER (WHERE latest_status = 'success') AS healthy_targets,
			COUNT(*) FILTER (WHERE latest_status = 'failed') AS failed_targets,
			COUNT(*) FILTER (WHERE latest_status = 'unknown') AS unknown_targets,
			AVG(latest_latency_ms) FILTER (WHERE latest_status = 'success' AND latest_latency_ms IS NOT NULL) AS avg_latency_ms,
			(
				SELECT name
				FROM model_status_targets
				%s
				ORDER BY last_checked_at DESC NULLS LAST
				LIMIT 1
			) AS last_checked_target
		FROM model_status_targets
		%s
	`, where, where))

	var (
		out        service.ModelStatusOverview
		avgLatency sql.NullFloat64
		lastTarget sql.NullString
	)
	if err := row.Scan(
		&out.TotalTargets,
		&out.EnabledTargets,
		&out.HealthyTargets,
		&out.FailedTargets,
		&out.UnknownTargets,
		&avgLatency,
		&lastTarget,
	); err != nil {
		return nil, err
	}
	if avgLatency.Valid {
		v := avgLatency.Float64
		out.AverageLatencyMs = &v
	}
	if lastTarget.Valid {
		v := lastTarget.String
		out.LastCheckedTarget = &v
	}
	return &out, nil
}

const modelStatusTargetSelectSQL = `
SELECT
	t.id,
	t.name,
	t.account_id,
	COALESCE(a.name, '') AS account_name,
	COALESCE(a.platform, '') AS account_platform,
	COALESCE(a.status, '') AS account_status,
	t.model_id,
	t.check_interval_seconds,
	t.timeout_seconds,
	t.enabled,
	t.latest_status,
	t.latest_latency_ms,
	t.latest_error_message,
	t.latest_response_text,
	t.consecutive_failures,
	t.last_checked_at,
	t.last_success_at,
	t.last_failure_at,
	t.next_check_at,
	t.created_at,
	t.updated_at
FROM model_status_targets t
LEFT JOIN accounts a ON a.id = t.account_id
`

type sqlScanner interface {
	Scan(dest ...any) error
}

func scanModelStatusTarget(row sqlScanner) (*service.ModelStatusTarget, error) {
	var (
		target        service.ModelStatusTarget
		latestLatency sql.NullInt64
		lastCheckedAt sql.NullTime
		lastSuccessAt sql.NullTime
		lastFailureAt sql.NullTime
		nextCheckAt   sql.NullTime
	)
	if err := row.Scan(
		&target.ID,
		&target.Name,
		&target.AccountID,
		&target.AccountName,
		&target.AccountPlatform,
		&target.AccountStatus,
		&target.ModelID,
		&target.CheckIntervalSeconds,
		&target.TimeoutSeconds,
		&target.Enabled,
		&target.LatestStatus,
		&latestLatency,
		&target.LatestErrorMessage,
		&target.LatestResponseText,
		&target.ConsecutiveFailures,
		&lastCheckedAt,
		&lastSuccessAt,
		&lastFailureAt,
		&nextCheckAt,
		&target.CreatedAt,
		&target.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if latestLatency.Valid {
		v := latestLatency.Int64
		target.LatestLatencyMs = &v
	}
	if lastCheckedAt.Valid {
		v := lastCheckedAt.Time
		target.LastCheckedAt = &v
	}
	if lastSuccessAt.Valid {
		v := lastSuccessAt.Time
		target.LastSuccessAt = &v
	}
	if lastFailureAt.Valid {
		v := lastFailureAt.Time
		target.LastFailureAt = &v
	}
	if nextCheckAt.Valid {
		v := nextCheckAt.Time
		target.NextCheckAt = &v
	}
	return &target, nil
}

func scanModelStatusTargets(rows *sql.Rows) ([]*service.ModelStatusTarget, error) {
	var targets []*service.ModelStatusTarget
	for rows.Next() {
		target, err := scanModelStatusTarget(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, rows.Err()
}

func scanModelStatusCheck(row sqlScanner) (*service.ModelStatusCheck, error) {
	var (
		check     service.ModelStatusCheck
		latencyMs sql.NullInt64
	)
	if err := row.Scan(
		&check.ID,
		&check.TargetID,
		&check.Status,
		&latencyMs,
		&check.ErrorMessage,
		&check.ResponseText,
		&check.StartedAt,
		&check.FinishedAt,
		&check.CreatedAt,
	); err != nil {
		return nil, err
	}
	if latencyMs.Valid {
		v := latencyMs.Int64
		check.LatencyMs = &v
	}
	return &check, nil
}

func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func normalizeText(v string) string {
	return strings.TrimSpace(v)
}
