package apiserver

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/acorn-io/acorn-dns/pkg/backend"
	"github.com/acorn-io/acorn-dns/pkg/model"
	"github.com/acorn-io/acorn-dns/pkg/version"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type handler struct {
	backend backend.Backend
}

func newHandler(b backend.Backend) *handler {
	return &handler{
		backend: b,
	}
}

func (h *handler) root(w http.ResponseWriter, r *http.Request) {
	v := version.Get()
	if err := json.NewEncoder(w).Encode(v); err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"success": false}`))
	}
	w.WriteHeader(200)
}

func (h *handler) getDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainName := vars["domain"]

	d, err := h.backend.GetDomain(domainName)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	writeSuccess(w, model.DomainResponse{Name: d.Domain}, "")
}

func (h *handler) createDomain(w http.ResponseWriter, r *http.Request) {
	domain, err := h.backend.CreateDomain()
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}
	writeSuccess(w, domain, "")
}

func (h *handler) renew(w http.ResponseWriter, r *http.Request) {
	var input model.RenewRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	vars := mux.Vars(r)
	domainName := vars["domain"]
	domainID := domainIDFromContext(r.Context())

	outOfSync, err := h.backend.Renew(domainName, domainID, input.Records)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	resp := model.RenewResponse{
		Name:             domainName,
		OutOfSyncRecords: outOfSync,
	}
	writeSuccess(w, resp, "")
}

func (h *handler) createRecord(w http.ResponseWriter, r *http.Request) {

	var input model.RecordRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	if err := validateRecord(input); err != nil {
		handleError(w, http.StatusUnprocessableEntity, err)
		return
	}

	vars := mux.Vars(r)
	domain := vars["domain"]
	domainID := domainIDFromContext(r.Context())

	record, err := h.backend.CreateRecord(domain, domainID, input)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	writeSuccess(w, record, "")
}

func (h *handler) deleteRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domain := vars["domain"]
	record := vars["record"]
	domainID := domainIDFromContext(r.Context())

	err := h.backend.DeleteRecord(record, domain, domainID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}
}

func validateRecord(input model.RecordRequest) error {
	if err := model.IsValidRecordType(input.Type); err != nil {
		return err
	}

	if input.Name == "" {
		return fmt.Errorf("record name must be provided")
	}

	if len(input.Values) == 0 {
		return fmt.Errorf("must supply at least one value")
	}

	// This could be overkill if we assume that k8s's ingress logic is validating this for us
	switch input.Type {
	case model.RecordTypeA:
		for _, v := range input.Values {
			ip := net.ParseIP(v)
			if ip == nil || strings.Contains(v, ":") {
				return fmt.Errorf("value %v is not a valid IPv4 address", v)
			}
		}
	case model.RecordTypeAAAA:
		for _, v := range input.Values {
			ip := net.ParseIP(v)
			if ip == nil || !strings.Contains(v, ":") {
				return fmt.Errorf("value %v is not a valid IPv6 address", v)
			}
		}
	case model.RecordTypeCname:
		if len(input.Values) != 1 {
			return fmt.Errorf("cname records must contain exactly one value. this contains %v values", len(input.Values))
		}
	}

	return nil
}

func (h *handler) deleteDomain(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("DELETE DOMAIN")
	// TODO  Implement
}
