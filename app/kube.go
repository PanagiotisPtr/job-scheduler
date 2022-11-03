package app

import (
	"context"
)

func (a *App) ListRunningJobs(
	ctx context.Context,
) ([]string, error) {
	return a.kubeService.ListRunningCronJobs(
		ctx,
	)
}

func (a *App) StartJob(
	ctx context.Context,
	jobName string,
) error {
	cronJob, err := a.cronJobService.GetCronJob(
		ctx,
		jobName,
	)
	if err != nil {
		return err
	}

	return a.kubeService.StartCronJob(
		ctx,
		cronJob,
	)
}

func (a *App) StopJob(
	ctx context.Context,
	jobName string,
) error {
	cronJob, err := a.cronJobService.GetCronJob(
		ctx,
		jobName,
	)
	if err != nil {
		return err
	}

	return a.kubeService.StopCronJob(
		ctx,
		cronJob,
	)
}
