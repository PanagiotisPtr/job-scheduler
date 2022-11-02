package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/panagiotisptr/job-scheduler/app"
	"go.uber.org/zap"
)

type CronJobController struct {
	logger *zap.Logger
	app    *app.App
}

func ProvideCronJobController(
	logger *zap.Logger,
	r *mux.Router,
	app *app.App,
) (*CronJobController, error) {
	c := &CronJobController{
		logger: logger,
		app:    app,
	}

	r.HandleFunc("/static/jobs", c.listJobs)
	r.HandleFunc("/static/jobs/{jobName}", c.getJob)

	return c, nil
}

func (c *CronJobController) listJobs(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second*2,
	)
	defer cancel()
	res, err := c.app.ListAvailableCronJobNames(ctx)
	if err != nil {
		errorResponse(
			w,
			err,
			http.StatusInternalServerError,
			c.logger,
		)
		return
	}

	writeObject(
		w,
		struct {
			JobNames []string `json:"jobNames"`
		}{
			JobNames: res,
		},
		http.StatusOK,
		c.logger,
	)
}

func (c *CronJobController) getJob(
	w http.ResponseWriter,
	r *http.Request,
) {
	jobName, ok := mux.Vars(r)["jobName"]
	if !ok {
		errorResponse(
			w,
			fmt.Errorf("could not find job"),
			http.StatusNotFound,
			c.logger,
		)
	}
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second*2,
	)
	defer cancel()
	res, err := c.app.GetCronJobConfig(ctx, jobName)
	if err != nil {
		errorResponse(
			w,
			err,
			http.StatusInternalServerError,
			c.logger,
		)
		return
	}

	writeObject(
		w,
		res,
		http.StatusOK,
		c.logger,
	)
}
