package service

import (
	"context"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	modelStatusRunnerTick       = time.Minute
	modelStatusRunnerBatchLimit = 50
	modelStatusRunnerWorkers    = 4
)

type ModelStatusRunnerService struct {
	repo    ModelStatusTargetRepository
	service *ModelStatusService

	startOnce sync.Once
	stopOnce  sync.Once
	stopCh    chan struct{}
}

func NewModelStatusRunnerService(
	repo ModelStatusTargetRepository,
	service *ModelStatusService,
) *ModelStatusRunnerService {
	return &ModelStatusRunnerService{
		repo:    repo,
		service: service,
	}
}

func (s *ModelStatusRunnerService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		s.stopCh = make(chan struct{})
		go s.run()
		logger.LegacyPrintf("service.model_status_runner", "[ModelStatusRunner] started (tick=%s)", modelStatusRunnerTick)
	})
}

func (s *ModelStatusRunnerService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.stopCh != nil {
			close(s.stopCh)
		}
	})
}

func (s *ModelStatusRunnerService) run() {
	s.runOnce()
	ticker := time.NewTicker(modelStatusRunnerTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runOnce()
		case <-s.stopCh:
			return
		}
	}
}

func (s *ModelStatusRunnerService) runOnce() {
	if s == nil || s.repo == nil || s.service == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	due, err := s.repo.ListDue(ctx, time.Now().UTC(), modelStatusRunnerBatchLimit)
	if err != nil {
		logger.LegacyPrintf("service.model_status_runner", "[ModelStatusRunner] list due failed: %v", err)
		return
	}
	if len(due) == 0 {
		return
	}

	sem := make(chan struct{}, modelStatusRunnerWorkers)
	var wg sync.WaitGroup
	for _, target := range due {
		targetID := target.ID
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if _, err := s.service.RunTargetCheck(ctx, targetID); err != nil {
				logger.LegacyPrintf("service.model_status_runner", "[ModelStatusRunner] target=%d failed: %v", targetID, err)
			}
		}()
	}
	wg.Wait()
}
