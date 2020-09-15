package server

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/zinic/forculus/config"
)

const (
	SHA256AuthorizationMethod = "sha256"
	AuthorizationHeaderKey    = "Authorization"
)

func methodFilter(handler HandlerFunc, accepts ...string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handlesMethod := false
		for _, acceptedMethod := range accepts {
			if request.Method == acceptedMethod {
				handlesMethod = true
				break
			}
		}

		if handlesMethod {
			wrapper := &responseWrapper{
				ResponseWriter: writer,
			}

			handler(wrapper, request)
		} else {
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func authFilter(users map[string]config.AuthorizationConfig, handler http.HandlerFunc) http.HandlerFunc {
	var (
		userHashes = make(map[string]string)
		hasher     = sha256.New()
	)

	for username, authConfig := range users {
		hasher.Write([]byte(fmt.Sprintf("%s:%s", username, authConfig.Password)))
		userHashes[fmt.Sprintf("%x", hasher.Sum(nil))] = username
		hasher.Reset()
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		if values, hasHeader := request.Header[AuthorizationHeaderKey]; !hasHeader {
			writer.WriteHeader(http.StatusUnauthorized)
		} else if len(values) > 1 {
			writer.WriteHeader(http.StatusBadRequest)
		} else if authHeaderParts := strings.Fields(values[0]); len(authHeaderParts) != 2 {
			writer.WriteHeader(http.StatusBadRequest)
		} else if hashMethod := authHeaderParts[0]; hashMethod != SHA256AuthorizationMethod {
			writer.WriteHeader(http.StatusBadRequest)
		} else if _, authValid := userHashes[authHeaderParts[1]]; !authValid {
			writer.WriteHeader(http.StatusBadRequest)
		} else {
			handler(writer, request)
		}
	}
}
