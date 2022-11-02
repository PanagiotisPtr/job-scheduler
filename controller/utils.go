package controller

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func errorResponse(
	w http.ResponseWriter,
	err error,
	code int,
	logger *zap.Logger,
) {
	w.WriteHeader(code)
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

func writeObject(
	w http.ResponseWriter,
	obj interface{},
	code int,
	logger *zap.Logger,
) {
	w.WriteHeader(code)
	b, err := json.Marshal(obj)
	if err != nil {
		errorResponse(
			w,
			err,
			http.StatusInternalServerError,
			logger,
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
