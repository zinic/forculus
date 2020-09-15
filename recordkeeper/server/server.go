package server

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/zinic/forculus/config"
)

func newMux(handler Handler, users map[string]config.AuthorizationConfig) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/event", authFilter(users, methodFilter(handler.PostEvent, http.MethodPost)))
	router.HandleFunc("/event/{event_id}", methodFilter(handler.GetEvent, http.MethodGet))

	return router
}
func NewServer(cfg config.RecordKeeperConfig, handler Handler) http.Server {
	return http.Server{
		Addr:    cfg.BindAddress,
		Handler: newMux(handler, cfg.Users),
	}
}
