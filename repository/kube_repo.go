package repository

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
)

// KubernetesRepository a repository to interface with the
// kubernetes client
type KubernetesRepository interface {
	// StartJob Start a cron job
	StartCronJob(ctx context.Context, cj *batchv1.CronJob) error

	// StopJob Stop a cron job
	StopCronJob(ctx context.Context, cj *batchv1.CronJob) error

	// GetRunningJobs get list of names of running jobs
	GetRunningCronJobs(ctx context.Context) ([]string, error)
}
