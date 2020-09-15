package server

import (
	"fmt"
	"net/http"

	"github.com/zinic/forculus/log"
)

type ResponseWrapper interface {
	Errorf(statusCode int, format string, args ...interface{})
	Error(statusCode int, message string)
	http.ResponseWriter
}

type HandlerFunc func(ResponseWrapper, *http.Request)

type responseWrapper struct {
	http.ResponseWriter
}

func (s *responseWrapper) Error(statusCode int, message string) {
	s.WriteHeader(statusCode)

	output := []byte(fmt.Sprintf("{\"message\": \"%s\"}", message))
	if _, err := s.Write(output); err != nil {
		log.Errorf("Failed to serialize response error: %v", err)
	}
}

func (s *responseWrapper) Errorf(statusCode int, format string, args ...interface{}) {
	s.Error(statusCode, fmt.Sprintf(format, args...))
}
