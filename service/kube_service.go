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

func (s *KubernetesService) ListRunningCronJobs(
	ctx context.Context,
) ([]string, error) {
	return s.repo.GetRunningCronJobs(ctx)
}

func (s *KubernetesService) StartCronJob(
	ctx context.Context,
	cj *batchv1.CronJob,
) error {
	return s.repo.StartCronJob(ctx, cj)
}

func (s *KubernetesService) StopCronJob(
	ctx context.Context,
	cj *batchv1.CronJob,
) error {
	return s.repo.StopCronJob(ctx, cj)
}
