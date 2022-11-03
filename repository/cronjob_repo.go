package repository

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
)

// CronJobRepository interfaces with the GitHub API to get available cronjobs
type CronJobRepository interface {
	// GetCronJobNames get list of available cronjob names
	GetCronJobNames(ctx context.Context) ([]string, error)

	// GetCronJob get cronjob configuration
	GetCronJob(ctx context.Context, name string) (*batchv1.CronJob, error)
}
