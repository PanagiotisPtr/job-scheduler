package repository

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
)

type KubernetesRepository interface {
	StartJob(ctx context.Context, cj batchv1.CronJob) error
	StopJob(ctx context.Context, cj batchv1.CronJob) error
	GetRunningJobs(ctx context.Context) ([]batchv1.CronJob, error)
}
