package apiserver

import (
	"encoding/json"
	"net/http"

	"github.com/acorn-io/acorn-dns/pkg/model"
	"github.com/sirupsen/logrus"
)

func writeErrorResponse(w http.ResponseWriter, httpStatus int, message string, data interface{}) {
	o := model.ErrorResponse{
		Status:  httpStatus,
		Message: message,
		Data:    data,
	}
	res, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_, _ = w.Write(res)
}

func handleError(w http.ResponseWriter, httpStatus int, err error) {
	logrus.Errorf("Error during request: %v", err)
	writeErrorResponse(w, httpStatus, err.Error(), nil)
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	res, err := json.Marshal(data)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(res)
}
