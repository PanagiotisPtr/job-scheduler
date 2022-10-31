package repository

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
)

type CronJobRepository interface {
	GetCronJobNames(ctx context.Context) ([]string, error)
	GetCronJob(ctx context.Context, name string) (batchv1.CronJob, error)
}
