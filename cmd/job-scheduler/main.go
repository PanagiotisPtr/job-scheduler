package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v48/github"
	"github.com/gorilla/mux"
	"github.com/panagiotisptr/job-scheduler/config"
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

func errorResponse(w http.ResponseWriter, err error, logger *zap.Logger) {
	w.WriteHeader(http.StatusInternalServerError)
	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})

	if err != nil {
		logger.Sugar().Error(
			"failed to marshal error: ",
			err,
		)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		logger.Sugar().Error(
			"failed to write to response: ",
			err,
		)
	}
}

func Bootstrap(
	cfg *config.Config,
	lc fx.Lifecycle,
	logger *zap.Logger,
	cronJobService *service.CronJobService,
	kubernetesService *service.KubernetesService,
) {
	r := mux.NewRouter()
	// ideally there'd be a controller and an
	// application service here but I'll leave it like this
	// for now
	r.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(cfg)
		_, err := w.Write(b)
		if err != nil {
			logger.Sugar().Error(
				"failed to write to response: ",
				err,
			)
		}
	})
	r.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		cronJobNames, err := cronJobService.ListAvailableJobs(
			context.Background(),
		)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		b, err := json.Marshal(struct {
			JobNames []string `json:"jobNames"`
		}{
			JobNames: cronJobNames,
		})
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		_, err = w.Write(b)
		if err != nil {
			logger.Sugar().Error(
				"failed to write to response: ",
				err,
			)
		}
	})
	r.HandleFunc("/jobs/{jobName}", func(w http.ResponseWriter, r *http.Request) {
		jobName, ok := mux.Vars(r)["jobName"]
		if !ok {
			errorResponse(
				w,
				fmt.Errorf("could not find job"),
				logger,
			)
		}
		cronJob, err := cronJobService.GetJob(
			context.Background(),
			jobName,
		)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		b, err := json.Marshal(cronJob)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		_, err = w.Write(b)
		if err != nil {
			logger.Sugar().Error(
				"failed to write to response: ",
				err,
			)
		}
	})
	r.HandleFunc("/runningJobs", func(w http.ResponseWriter, r *http.Request) {
		cronJobs, err := kubernetesService.ListRunningJobs(
			context.Background(),
		)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		b, err := json.Marshal(cronJobs)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		_, err = w.Write(b)
		if err != nil {
			logger.Sugar().Error(
				"failed to write to response: ",
				err,
			)
		}
	})
	r.HandleFunc("/jobs/{jobName}/start", func(w http.ResponseWriter, r *http.Request) {
		jobName, ok := mux.Vars(r)["jobName"]
		if !ok {
			errorResponse(
				w,
				fmt.Errorf("could not find job"),
				logger,
			)
		}
		cronJob, err := cronJobService.GetJob(
			context.Background(),
			jobName,
		)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		err = kubernetesService.StartJob(
			context.Background(),
			cronJob,
		)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/jobs/{jobName}/stop", func(w http.ResponseWriter, r *http.Request) {
		jobName, ok := mux.Vars(r)["jobName"]
		if !ok {
			errorResponse(
				w,
				fmt.Errorf("could not find job"),
				logger,
			)
		}
		cronJob, err := cronJobService.GetJob(
			context.Background(),
			jobName,
		)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		err = kubernetesService.StopJob(
			context.Background(),
			cronJob,
		)
		if err != nil {
			errorResponse(w, err, logger)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

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
			config.ProvideConfigFromRemote,
			parser.ProvideCronJobParser,
			githubRepo.ProvideGitHubCronJobRepository,
			kubeRepo.ProvideKubernetesRepository,
			service.ProvideCronJobService,
			service.ProvideKubernetesService,
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
