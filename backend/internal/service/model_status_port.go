package service

import (
	"context"
	"time"
)

const (
	ModelStatusUnknown = "unknown"
	ModelStatusSuccess = "success"
	ModelStatusFailed  = "failed"
)

type ModelStatusTarget struct {
	ID                   int64      `json:"id"`
	Name                 string     `json:"name"`
	AccountID            int64      `json:"account_id"`
	AccountName          string     `json:"account_name"`
	AccountPlatform      string     `json:"account_platform"`
	AccountStatus        string     `json:"account_status"`
	ModelID              string     `json:"model_id"`
	CheckIntervalSeconds int        `json:"check_interval_seconds"`
	TimeoutSeconds       int        `json:"timeout_seconds"`
	Enabled              bool       `json:"enabled"`
	LatestStatus         string     `json:"latest_status"`
	LatestLatencyMs      *int64     `json:"latest_latency_ms"`
	LatestErrorMessage   string     `json:"latest_error_message"`
	LatestResponseText   string     `json:"latest_response_text"`
	ConsecutiveFailures  int        `json:"consecutive_failures"`
	LastCheckedAt        *time.Time `json:"last_checked_at"`
	LastSuccessAt        *time.Time `json:"last_success_at"`
	LastFailureAt        *time.Time `json:"last_failure_at"`
	NextCheckAt          *time.Time `json:"next_check_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ModelStatusCheck struct {
	ID           int64     `json:"id"`
	TargetID     int64     `json:"target_id"`
	Status       string    `json:"status"`
	LatencyMs    *int64    `json:"latency_ms"`
	ErrorMessage string    `json:"error_message"`
	ResponseText string    `json:"response_text"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type ModelStatusOverview struct {
	TotalTargets      int      `json:"total_targets"`
	EnabledTargets    int      `json:"enabled_targets"`
	HealthyTargets    int      `json:"healthy_targets"`
	FailedTargets     int      `json:"failed_targets"`
	UnknownTargets    int      `json:"unknown_targets"`
	AverageLatencyMs  *float64 `json:"average_latency_ms"`
	LastCheckedTarget *string  `json:"last_checked_target"`
}

type ModelStatusTargetRepository interface {
	Create(ctx context.Context, target *ModelStatusTarget) (*ModelStatusTarget, error)
	GetByID(ctx context.Context, id int64) (*ModelStatusTarget, error)
	List(ctx context.Context, includeDisabled bool) ([]*ModelStatusTarget, error)
	ListDue(ctx context.Context, now time.Time, limit int) ([]*ModelStatusTarget, error)
	Update(ctx context.Context, target *ModelStatusTarget) (*ModelStatusTarget, error)
	Delete(ctx context.Context, id int64) error
	RecordCheck(ctx context.Context, targetID int64, check *ModelStatusCheck, nextCheckAt *time.Time) (*ModelStatusTarget, error)
	ListChecks(ctx context.Context, targetID int64, limit int) ([]*ModelStatusCheck, error)
	GetOverview(ctx context.Context, includeDisabled bool) (*ModelStatusOverview, error)
}
