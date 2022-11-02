package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v48/github"
	"github.com/gorilla/mux"
	"github.com/panagiotisptr/job-scheduler/app"
	"github.com/panagiotisptr/job-scheduler/config"
	"github.com/panagiotisptr/job-scheduler/controller"
	"github.com/panagiotisptr/job-scheduler/parser"
	githubRepo "github.com/panagiotisptr/job-scheduler/repository/github"
	kubeRepo "github.com/panagiotisptr/job-scheduler/repository/kubernetes"
	"github.com/panagiotisptr/job-scheduler/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ProvideKuberentesClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	return kubernetes.NewForConfig(config)
}

func ProvideGitHubClient(
	cfg *config.Config,
) (*github.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubConfig.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client, nil
}

func ProvideLogger() *zap.Logger {
	logger, _ := zap.NewProduction()

	return logger
}

func ProvideMuxRouter() *mux.Router {
	return mux.NewRouter()
}

func Bootstrap(
	cfg *config.Config,
	lc fx.Lifecycle,
	logger *zap.Logger,
	r *mux.Router,

	// need these here to invoke them
	cronJobController *controller.CronJobController,
	kubeController *controller.KubernetesController,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go http.ListenAndServe(fmt.Sprintf(":%d", cfg.Service.Port), r)

			return nil
		},
	})
}

func main() {
	app := fx.New(
		fx.Provide(
			ProvideLogger,
			ProvideGitHubClient,
			ProvideKuberentesClientset,
			ProvideMuxRouter,
			config.ProvideRemoteConfig,
			parser.ProvideCronJobParser,
			githubRepo.ProvideGitHubCronJobRepository,
			kubeRepo.ProvideKubernetesRepository,
			service.ProvideCronJobService,
			service.ProvideKubernetesService,
			app.ProvideApp,
			controller.ProvideCronJobController,
			controller.ProvideKubernetesController,
		),
		fx.Invoke(Bootstrap),
		fx.WithLogger(
			func(logger *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: logger}
			},
		),
	)

	app.Run()
}
