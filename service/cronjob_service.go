package service

import (
	"context"

	"github.com/panagiotisptr/job-scheduler/repository"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
)

type CronJobService struct {
	repo   repository.CronJobRepository
	logger *zap.Logger
}

func ProvideCronJobService(
	repo repository.CronJobRepository,
	logger *zap.Logger,
) (*CronJobService, error) {
	return &CronJobService{
		repo:   repo,
		logger: logger,
	}, nil
}

func (s *CronJobService) ListAvailableJobs(
	ctx context.Context,
) ([]string, error) {
	return s.repo.GetCronJobNames(ctx)
}

func (s *CronJobService) GetJob(
	ctx context.Context,
	name string,
) (batchv1.CronJob, error) {
	return s.repo.GetCronJob(ctx, name)
}
