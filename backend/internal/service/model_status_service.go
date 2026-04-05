package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ModelStatusService struct {
	repo           ModelStatusTargetRepository
	accountRepo    AccountRepository
	accountTestSvc *AccountTestService
}

func NewModelStatusService(
	repo ModelStatusTargetRepository,
	accountRepo AccountRepository,
	accountTestSvc *AccountTestService,
) *ModelStatusService {
	return &ModelStatusService{
		repo:           repo,
		accountRepo:    accountRepo,
		accountTestSvc: accountTestSvc,
	}
}

func (s *ModelStatusService) CreateTarget(ctx context.Context, target *ModelStatusTarget) (*ModelStatusTarget, error) {
	if target == nil {
		return nil, errors.New("target is required")
	}
	target = cloneModelStatusTarget(target)
	if err := s.normalizeTarget(ctx, target, true); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, target)
}

func (s *ModelStatusService) GetTarget(ctx context.Context, id int64) (*ModelStatusTarget, error) {
	if id <= 0 {
		return nil, errors.New("invalid target id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *ModelStatusService) ListTargets(ctx context.Context, includeDisabled bool) ([]*ModelStatusTarget, error) {
	return s.repo.List(ctx, includeDisabled)
}

func (s *ModelStatusService) UpdateTarget(ctx context.Context, target *ModelStatusTarget) (*ModelStatusTarget, error) {
	if target == nil || target.ID <= 0 {
		return nil, errors.New("invalid target")
	}
	existing, err := s.repo.GetByID(ctx, target.ID)
	if err != nil {
		return nil, err
	}
	merged := *existing
	if strings.TrimSpace(target.Name) != "" {
		merged.Name = target.Name
	}
	if target.AccountID > 0 {
		merged.AccountID = target.AccountID
	}
	if strings.TrimSpace(target.ModelID) != "" {
		merged.ModelID = target.ModelID
	}
	if target.CheckIntervalSeconds > 0 {
		merged.CheckIntervalSeconds = target.CheckIntervalSeconds
	}
	if target.TimeoutSeconds > 0 {
		merged.TimeoutSeconds = target.TimeoutSeconds
	}
	merged.Enabled = target.Enabled
	if err := s.normalizeTarget(ctx, &merged, false); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, &merged)
}

func (s *ModelStatusService) DeleteTarget(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid target id")
	}
	return s.repo.Delete(ctx, id)
}

func (s *ModelStatusService) ListChecks(ctx context.Context, targetID int64, limit int) ([]*ModelStatusCheck, error) {
	if targetID <= 0 {
		return nil, errors.New("invalid target id")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	return s.repo.ListChecks(ctx, targetID, limit)
}

func (s *ModelStatusService) GetOverview(ctx context.Context, includeDisabled bool) (*ModelStatusOverview, error) {
	return s.repo.GetOverview(ctx, includeDisabled)
}

func (s *ModelStatusService) RunTargetCheck(ctx context.Context, targetID int64) (*ModelStatusTarget, error) {
	if targetID <= 0 {
		return nil, errors.New("invalid target id")
	}
	if s.accountTestSvc == nil {
		return nil, errors.New("account test service is unavailable")
	}
	target, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if _, err := s.accountRepo.GetByID(ctx, target.AccountID); err != nil {
		return nil, fmt.Errorf("load account for target: %w", err)
	}

	timeout := time.Duration(target.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := s.accountTestSvc.RunTestBackground(runCtx, target.AccountID, target.ModelID)
	if err != nil {
		// RunTestBackground normally folds errors into result; keep this as last-resort fallback.
		now := time.Now().UTC()
		msg := err.Error()
		result = &ScheduledTestResult{
			Status:       ModelStatusFailed,
			ErrorMessage: msg,
			StartedAt:    now,
			FinishedAt:   now,
		}
	}

	check := &ModelStatusCheck{
		TargetID:     target.ID,
		Status:       normalizeModelStatus(result.Status),
		ErrorMessage: strings.TrimSpace(result.ErrorMessage),
		ResponseText: strings.TrimSpace(result.ResponseText),
		StartedAt:    result.StartedAt,
		FinishedAt:   result.FinishedAt,
	}
	if result.LatencyMs > 0 {
		latency := result.LatencyMs
		check.LatencyMs = &latency
	}

	var nextCheckAt *time.Time
	if target.Enabled {
		next := time.Now().UTC().Add(time.Duration(target.CheckIntervalSeconds) * time.Second)
		nextCheckAt = &next
	}

	return s.repo.RecordCheck(ctx, target.ID, check, nextCheckAt)
}

func (s *ModelStatusService) normalizeTarget(ctx context.Context, target *ModelStatusTarget, create bool) error {
	target.Name = strings.TrimSpace(target.Name)
	target.ModelID = strings.TrimSpace(target.ModelID)
	target.LatestStatus = normalizeModelStatus(target.LatestStatus)

	if target.AccountID <= 0 {
		return errors.New("account_id is required")
	}
	account, err := s.accountRepo.GetByID(ctx, target.AccountID)
	if err != nil {
		return fmt.Errorf("invalid account_id: %w", err)
	}
	if account == nil {
		return errors.New("account not found")
	}

	if target.ModelID == "" {
		return errors.New("model_id is required")
	}
	if target.CheckIntervalSeconds <= 0 {
		target.CheckIntervalSeconds = 300
	}
	if target.CheckIntervalSeconds < 60 {
		return errors.New("check_interval_seconds must be at least 60")
	}
	if target.TimeoutSeconds <= 0 {
		target.TimeoutSeconds = 45
	}
	if target.TimeoutSeconds < 5 || target.TimeoutSeconds > 300 {
		return errors.New("timeout_seconds must be between 5 and 300")
	}

	if target.Name == "" {
		target.Name = fmt.Sprintf("%s / %s", strings.TrimSpace(account.Name), target.ModelID)
	}

	if create {
		target.LatestStatus = ModelStatusUnknown
		if target.Enabled {
			now := time.Now().UTC()
			target.NextCheckAt = &now
		} else {
			target.NextCheckAt = nil
		}
	} else {
		if target.Enabled {
			if target.NextCheckAt == nil {
				now := time.Now().UTC()
				target.NextCheckAt = &now
			}
		} else {
			target.NextCheckAt = nil
		}
	}

	return nil
}

func normalizeModelStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case ModelStatusSuccess:
		return ModelStatusSuccess
	case ModelStatusFailed:
		return ModelStatusFailed
	default:
		return ModelStatusUnknown
	}
}

func cloneModelStatusTarget(target *ModelStatusTarget) *ModelStatusTarget {
	if target == nil {
		return nil
	}
	cp := *target
	return &cp
}
