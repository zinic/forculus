package api

import "net/http"

func NewServer(addr string, handler Handler) http.Server {
	return http.Server{
		Addr:    addr,
		Handler: newMux(handler),
	}
}
