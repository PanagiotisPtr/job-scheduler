package memory

import (
	"context"

	"github.com/panagiotisptr/job-scheduler/repository"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
)

type KubernetesMemoryRepository struct {
	jobs   map[string]struct{}
	logger *zap.Logger
}

func ProvideKubernetesMemoryRepository(
	logger *zap.Logger,
) repository.KubernetesRepository {
	logger.Sugar().Info("using in-memory kubernetes repository. Changes are not applied to the cluster")
	return &KubernetesMemoryRepository{
		jobs:   make(map[string]struct{}),
		logger: logger,
	}
}

func (r *KubernetesMemoryRepository) StartCronJob(
	ctx context.Context,
	cj *batchv1.CronJob,
) error {
	r.jobs[cj.Name] = struct{}{}
	return nil
}

func (r *KubernetesMemoryRepository) StopCronJob(
	ctx context.Context,
	cj *batchv1.CronJob,
) error {
	delete(r.jobs, cj.Name)
	return nil
}

func (r *KubernetesMemoryRepository) GetRunningCronJobs(
	ctx context.Context,
) ([]string, error) {
	names := []string{}
	for n := range r.jobs {
		names = append(names, n)
	}

	return names, nil
}
