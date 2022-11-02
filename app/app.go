package app

import (
	"github.com/panagiotisptr/job-scheduler/service"
	"go.uber.org/zap"
)

type App struct {
	logger         *zap.Logger
	cronJobService *service.CronJobService
	kubeService    *service.KubernetesService
}

func ProvideApp(
	logger *zap.Logger,
	cronJobService *service.CronJobService,
	kubeService *service.KubernetesService,
) *App {
	return &App{
		logger:         logger,
		cronJobService: cronJobService,
		kubeService:    kubeService,
	}
}
