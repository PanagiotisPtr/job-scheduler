package service

import (
	"context"

	"github.com/panagiotisptr/job-scheduler/repository"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
)

type KubernetesService struct {
	repo   repository.KubernetesRepository
	logger *zap.Logger
}

func ProvideKubernetesService(
	repo repository.KubernetesRepository,
	logger *zap.Logger,
) (*KubernetesService, error) {
	return &KubernetesService{
		repo:   repo,
		logger: logger,
	}, nil
}

func (s *KubernetesService) ListRunningJobs(
	ctx context.Context,
) ([]batchv1.CronJob, error) {
	return s.repo.GetRunningJobs(ctx)
}

func (s *KubernetesService) StartJob(
	ctx context.Context,
	cj batchv1.CronJob,
) error {
	return s.repo.StartJob(ctx, cj)
}

func (s *KubernetesService) StopJob(
	ctx context.Context,
	cj batchv1.CronJob,
) error {
	return s.repo.StopJob(ctx, cj)
}
