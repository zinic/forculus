package apitools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func CopyURLValues(values url.Values) url.Values {
	copied := make(url.Values, len(values))
	for key, valueSet := range values {
		copied[key] = valueSet
	}

	return copied
}

type Endpoint struct {
	rootURL string
}

func NewEndpoint(scheme, host string, port int, rootPath string) Endpoint {
	var rootURL = fmt.Sprintf("%s://%s:%d", scheme, host, port)

	if len(rootPath) > 0 {
		rootURL = fmt.Sprintf("%s/%s", rootURL, rootPath)
	}

	return Endpoint{
		rootURL: rootURL,
	}
}

func (s Endpoint) Format(path ...string) string {
	formattedPaths := make([]string, len(path))

	for idx, path := range path {
		formattedPaths[idx] = url.PathEscape(path)
	}

	if len(s.rootURL) > 0 {
		return fmt.Sprintf("%s/%s", s.rootURL, strings.Join(formattedPaths, "/"))
	} else if len(path) > 1 {
		return strings.Join(formattedPaths, "/")
	} else {
		return path[0]
	}
}

func (s Endpoint) FormatQuery(query url.Values, path ...string) string {
	formattedURL := s.Format(path...)

	if query != nil && len(query) > 0 {
		formattedURL = fmt.Sprintf("%s?%s", formattedURL, query.Encode())
	}

	return formattedURL
}

func NewHTTPClientWrapper(endpoint Endpoint) *HTTPClientWrapper {
	return &HTTPClientWrapper{
		Endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}

type HTTPClientWrapper struct {
	Endpoint   Endpoint
	httpClient *http.Client
}

func (s *HTTPClientWrapper) doRequest(method string, body io.Reader, query url.Values, header http.Header, path ...string) (*http.Response, error) {
	if req, err := http.NewRequest(method, s.Endpoint.FormatQuery(query, path...), body); err != nil {
		return nil, err
	} else {
		req.Header = header
		return s.httpClient.Do(req)
	}
}

func (s *HTTPClientWrapper) GET(body io.Reader, query url.Values, header http.Header, path ...string) (*http.Response, error) {
	return s.doRequest(http.MethodGet, body, query, header, path...)
}

func (s *HTTPClientWrapper) POST(body io.Reader, query url.Values, header http.Header, path ...string) (*http.Response, error) {
	return s.doRequest(http.MethodPost, body, query, header, path...)
}
