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
	"github.com/panagiotisptr/job-scheduler/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func ProvideGitHubClient(
	cfg *config.Config,
) (*github.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubCinfig.AccessToken},
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
	service *service.CronJobService,
) {
	r := mux.NewRouter()
	r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello"))
		if err != nil {
			logger.Sugar().Error(
				"failed to write to response: ",
				err,
			)
		}
	})
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
		cronJobNames, err := service.ListAvailableJobs(
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
		cronJob, err := service.GetJob(
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
			config.ProvideConfig,
			parser.ProvideCronJobParser,
			githubRepo.ProvideGitHubCronJobRepository,
			service.ProvideCronJobService,
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
