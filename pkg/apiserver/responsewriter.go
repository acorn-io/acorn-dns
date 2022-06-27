package apiserver

import (
	"encoding/json"
	"net/http"

	"github.com/acorn-io/acorn-dns/pkg/model"
	"github.com/sirupsen/logrus"
)

func writeError(w http.ResponseWriter, httpStatus int, err error) {
	logrus.Errorf("got a response error: %v", err)
	o := model.ErrorResponse{
		Status:  httpStatus,
		Message: err.Error(),
	}
	res, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_, _ = w.Write(res)
}

func writeSuccess(w http.ResponseWriter, data interface{}, msg string) {
	res, err := json.Marshal(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(res)
}
