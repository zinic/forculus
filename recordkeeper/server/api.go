package server

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/storage"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/zinic/forculus/recordkeeper/model"

	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/recordkeeper/db"
)

func NewHandler(cfg config.RecordKeeperConfig) (Handler, error) {
	if database, err := db.NewDatabase(cfg.DatabasePath); err != nil {
		return Handler{}, nil
	} else {
		return Handler{
			cfg:      cfg,
			database: database,
		}, nil
	}
}

type Handler struct {
	cfg              config.RecordKeeperConfig
	database         *db.Database
	storageProviders map[string]storage.Provider
}

func (s *Handler) Close() error {
	return s.database.Close()
}

const (
	eventIDVarKey       = "event_id"
	EventAccessTokenKey = "access_token"
)

func (s *Handler) GetEvent(resp ResponseWrapper, req *http.Request) {
	vars := mux.Vars(req)
	rawEventRecordID := vars[eventIDVarKey]

	if eventRecordID, err := strconv.ParseUint(rawEventRecordID, 10, 64); err != nil {
		resp.Errorf(http.StatusBadRequest, "malformed event ID %s", rawEventRecordID)
	} else if accessTokenValues, hasToken := req.URL.Query()[EventAccessTokenKey]; !hasToken {
		resp.Error(http.StatusUnauthorized, "no access token specified")
	} else if len(accessTokenValues) > 1 {
		resp.Error(http.StatusBadRequest, "too many access tokens specified")
	} else if eventRecord, err := s.database.GetEventRecord(eventRecordID); err != nil {
		if err == db.ErrEventNotFound {
			resp.Errorf(http.StatusNotFound, "event ID %s not found", rawEventRecordID)
		} else {
			resp.Error(http.StatusInternalServerError, "database error")
		}
	} else if eventRecord.AccessToken != accessTokenValues[0] {
		resp.WriteHeader(http.StatusUnauthorized)
	} else if storageProvider, hasProvider := s.storageProviders[eventRecord.StorageTarget]; !hasProvider {
		resp.Errorf(http.StatusInternalServerError, "storage provider %s is not configured", eventRecord.StorageTarget)
	} else if eventInput, err := storageProvider.Read(eventRecord.StorageKey); err != nil {
		resp.Error(http.StatusInternalServerError, "storage provider error")
	} else {
		defer eventInput.Close()

		resp.WriteHeader(http.StatusOK)

		if _, err := io.Copy(resp, eventInput); err != nil {
			log.Errorf("Failed to copy storage provider stream to response writer: %v", err)
		}
	}
}

func (s *Handler) PostEvent(resp ResponseWrapper, req *http.Request) {
	if content, err := ioutil.ReadAll(req.Body); err != nil {
		resp.Error(http.StatusBadRequest, "failed to read request body")
	} else {
		var putEventReq model.EventRecord

		if err := json.Unmarshal(content, &putEventReq); err != nil {
			resp.Errorf(http.StatusBadRequest, "bad JSON input: %v", err)
		} else if recordID, err := s.database.WriteEventRecord(putEventReq); err != nil {
			resp.Errorf(http.StatusInternalServerError, "database error: %v", err)
		} else {
			putEventReq.ID = recordID

			if output, err := json.Marshal(&putEventReq); err != nil {
				resp.Errorf(http.StatusInternalServerError, "response marshaling error: %v", err)
			} else {
				resp.WriteHeader(http.StatusOK)
				resp.Write(output)
			}
		}
	}
}
