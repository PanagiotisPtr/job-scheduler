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

type KubernetesController struct {
	logger *zap.Logger
	app    *app.App
}

func ProvideKubernetesController(
	logger *zap.Logger,
	r *mux.Router,
	app *app.App,
) (*KubernetesController, error) {
	c := &KubernetesController{
		logger: logger,
		app:    app,
	}

	r.HandleFunc("/cluster/jobs", c.listRunningJobs)
	r.HandleFunc("/cluster/jobs/{jobName}/start", c.startJob)
	r.HandleFunc("/cluster/jobs/{jobName}/stop", c.stopJob)

	return c, nil
}

func (c *KubernetesController) listRunningJobs(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second*2,
	)
	defer cancel()
	res, err := c.app.ListRunningJobs(ctx)
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

func (c *KubernetesController) startJob(
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
	err := c.app.StartJob(
		ctx,
		jobName,
	)
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
			Success bool `json:"success"`
		}{
			Success: true,
		},
		http.StatusOK,
		c.logger,
	)
}

func (c *KubernetesController) stopJob(
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
	err := c.app.StopJob(
		ctx,
		jobName,
	)
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
			Success bool `json:"success"`
		}{
			Success: true,
		},
		http.StatusOK,
		c.logger,
	)
}
