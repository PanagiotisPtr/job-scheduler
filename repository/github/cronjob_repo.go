package github

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/go-github/v48/github"
	"github.com/panagiotisptr/job-scheduler/config"
	"github.com/panagiotisptr/job-scheduler/parser"
	"github.com/panagiotisptr/job-scheduler/repository"
	"go.uber.org/fx"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
)

const (
	syncTime         = time.Minute * 5
	timeoutThreshold = time.Minute * 2
)

func isYaml(path string) bool {
	return strings.Contains(path, ".yml") ||
		strings.Contains(path, ".yaml")
}

type GitHubCronJobRepository struct {
	logger        *zap.Logger
	client        *github.Client
	cronJobs      map[string]config.GitHubRepositoryArgs
	cronJobParser *parser.CronJobParser
}

func ProvideGitHubCronJobRepository(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger *zap.Logger,
	client *github.Client,
	p *parser.CronJobParser,
) (repository.CronJobRepository, error) {
	repo := &GitHubCronJobRepository{
		logger:        logger,
		cronJobs:      make(map[string]config.GitHubRepositoryArgs),
		client:        client,
		cronJobParser: p,
	}

	ticker := time.NewTicker(syncTime)
	stop := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutThreshold))
			defer cancel()
			err := repo.sync(timeoutCtx, cfg.GitHubConfig.Locations)
			if err != nil {
				repo.logger.Sugar().Error("failed to sync cronjobs with GitHub: ", err)
			}
			go func() {
				select {
				case <-ticker.C:
					repo.logger.Sugar().Info("syncing cronjobs with github")
					tctx, cl := context.WithTimeout(ctx, time.Duration(timeoutThreshold))
					defer cl()
					err := repo.sync(tctx, cfg.GitHubConfig.Locations)
					if err != nil {
						repo.logger.Sugar().Error("failed to sync cronjobs with GitHub: ", err)
					}
					repo.logger.Sugar().Info("cronjobs synced")
				case <-stop:
					ticker.Stop()
					return
				}
			}()

			return nil
		},

		OnStop: func(ctx context.Context) error {
			stop <- struct{}{}

			return nil
		},
	})

	return repo, nil
}

func (r *GitHubCronJobRepository) sync(
	ctx context.Context,
	locations []config.GitHubRepositoryArgs,
) error {
	completed := make(chan struct{})
	var err error

	go func() {
		for _, location := range locations {
			paths := []string{location.Path}

			for len(paths) > 0 {
				p := paths[len(paths)-1]
				paths = paths[:len(paths)-1]

				_, content, _, err := r.client.Repositories.GetContents(
					ctx,
					location.Owner,
					location.Name,
					p,
					nil,
				)
				if err != nil {
					r.logger.With(
						zap.String("owner", location.Owner),
						zap.String("name", location.Name),
						zap.String("path", p),
					).Sugar().Error(
						"failed to get repository contents: ",
						err,
					)
					continue
				}

				for _, c := range content {
					switch c.GetType() {
					case "dir":
						if c.Path != nil {
							paths = append(paths, c.GetPath())
						}
					case "file":
						if !isYaml(c.GetPath()) {
							continue
						}
						reader, err := r.getFileReader(
							ctx,
							config.GitHubRepositoryArgs{
								Owner: location.Owner,
								Name:  location.Name,
								Path:  c.GetPath(),
							},
						)
						if err != nil {
							r.logger.With(
								zap.String("owner", location.Owner),
								zap.String("name", location.Name),
								zap.String("path", p),
							).Sugar().Error(
								"failed to get reader for file: ",
								err,
							)
							continue
						}
						defer reader.Close()
						cronJobs := r.cronJobParser.ParseCronJobConfigs(
							reader,
						)
						for _, cj := range cronJobs {
							// one yaml file could have multiple cron jobs
							r.cronJobs[cj.Name] = config.GitHubRepositoryArgs{
								Owner: location.Owner,
								Name:  location.Name,
								Path:  c.GetPath(),
							}
						}
					}
				}
			}
		}

		completed <- struct{}{}
	}()

	select {
	case <-completed:
		return err
	case <-ctx.Done():
		return fmt.Errorf("failed to sync files. Context timeout")
	}
}

func (r *GitHubCronJobRepository) getFileReader(
	ctx context.Context,
	location config.GitHubRepositoryArgs,
) (io.ReadCloser, error) {
	reader, _, err := r.client.Repositories.DownloadContents(
		ctx,
		location.Owner,
		location.Name,
		location.Path,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (r *GitHubCronJobRepository) GetCronJobNames(
	ctx context.Context,
) ([]string, error) {
	names := []string{}
	for name := range r.cronJobs {
		names = append(names, name)
	}

	return names, nil
}

func (r *GitHubCronJobRepository) GetCronJob(
	ctx context.Context,
	name string,
) (batchv1.CronJob, error) {
	var cronJob batchv1.CronJob
	location, ok := r.cronJobs[name]
	if !ok {
		return cronJob, fmt.Errorf("could not find cronjob with name: %s", name)
	}

	reader, err := r.getFileReader(
		ctx,
		location,
	)
	if err != nil {
		r.logger.With(
			zap.String("owner", location.Owner),
			zap.String("name", location.Name),
			zap.String("path", location.Path),
		).Sugar().Error(
			"failed to get reader for file: ",
			err,
		)
		return cronJob, err
	}

	cronJobs := r.cronJobParser.ParseCronJobConfigs(
		reader,
	)

	for _, cj := range cronJobs {
		if cj.Name == name {
			return cj, nil
		}
	}

	r.logger.Info("no")
	return cronJob, fmt.Errorf(
		"failed to find cronjob with name: %s",
		name,
	)
}
