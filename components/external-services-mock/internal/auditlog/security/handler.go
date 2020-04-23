package security

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type SecurityEventService interface {
	Save(change model.SecurityEvent) (string, error)
	Get(id string) *model.SecurityEvent
	Delete(id string)
	List() []model.SecurityEvent
}

type ConfigChangeHandler struct {
	service SecurityEventService
	logger  *log.Logger
}

func NewSecurityEventHandler(service SecurityEventService, logger *log.Logger) *ConfigChangeHandler {
	return &ConfigChangeHandler{service: service, logger: logger}
}

func (h *ConfigChangeHandler) Save(writer http.ResponseWriter, req *http.Request) {
	defer h.closeBody(req)
	var auditLog model.SecurityEvent
	err := json.NewDecoder(req.Body).Decode(&auditLog)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while decoding input"), http.StatusInternalServerError)
		return
	}

	id, err := h.service.Save(auditLog)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while saving security event log"), http.StatusInternalServerError)
		return
	}

	response := model.SuccessResponse{ID: id}
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(&response)
	if err != nil {
		err := errors.New("error while encoding response")
		httphelpers.WriteError(writer, err, http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) Get(writer http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		httphelpers.WriteError(writer, errors.New("parameter id not provided"), http.StatusBadRequest)
		return
	}

	val := h.service.Get(id)
	if val == nil {
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	err := json.NewEncoder(writer).Encode(&val)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("error while encoding response"), http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) List(writer http.ResponseWriter, req *http.Request) {
	values := h.service.List()

	err := json.NewEncoder(writer).Encode(&values)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("error while encoding response"), http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) Delete(writer http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		httphelpers.WriteError(writer, errors.New("parameter id not provided"), http.StatusBadRequest)
	}

	h.service.Delete(id)
}

func (h *ConfigChangeHandler) closeBody(req *http.Request) {
	err := req.Body.Close()
	if err != nil {
		h.logger.Error(errors.Wrap(err, "while closing body"))
	}
}