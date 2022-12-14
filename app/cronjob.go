package app

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
)

func (a *App) ListAvailableCronJobNames(
	ctx context.Context,
) ([]string, error) {
	return a.cronJobService.ListAvailableCronJobs(
		context.Background(),
	)
}

func (a *App) GetCronJobConfig(
	ctx context.Context,
	jobName string,
) (*batchv1.CronJob, error) {
	return a.cronJobService.GetCronJob(
		context.Background(),
		jobName,
	)
}
