package api

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/gorilla/mux"
	"net/http"
)

func newMux(handler Handler) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/event", handler.PutEvent)

	return router
}

func NewHandler() (Handler, error) {
	if database, err := badger.Open(badger.DefaultOptions("rkdb")); err != nil {
		return Handler{}, nil
	} else {
		return Handler{
			database: database,
		}, nil
	}
}

type Handler struct {
	database *badger.DB
}

func (s *Handler) Close() error {
	return s.database.Close()
}

func (s *Handler) PutEvent(resp http.ResponseWriter, req *http.Request) {

}
