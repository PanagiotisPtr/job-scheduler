package app

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
)

func (a *App) ListRunningJobs(
	ctx context.Context,
) ([]batchv1.CronJob, error) {
	return a.kubeService.ListRunningJobs(
		ctx,
	)
}

func (a *App) StartJob(
	ctx context.Context,
	jobName string,
) error {
	cronJob, err := a.cronJobService.GetJob(
		ctx,
		jobName,
	)
	if err != nil {
		return err
	}

	return a.kubeService.StartJob(
		ctx,
		cronJob,
	)
}

func (a *App) StopJob(
	ctx context.Context,
	jobName string,
) error {
	cronJob, err := a.cronJobService.GetJob(
		ctx,
		jobName,
	)
	if err != nil {
		return err
	}

	return a.kubeService.StartJob(
		ctx,
		cronJob,
	)
}
