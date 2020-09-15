package rkapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/zinic/forculus/recordkeeper/rkdb"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/zinic/forculus/apitools"
	"github.com/zinic/forculus/recordkeeper/server"
)

type Credentials struct {
	Username string
	Password string
}

func NewClient(credentials Credentials, endpoint apitools.Endpoint) Client {
	return &recordKeeperClient{
		credentials: credentials,
		httpClient:  apitools.NewHTTPClientWrapper(endpoint),
	}
}

type Client interface {
	CreateEventRecord(req rkdb.CreateEventRecord) (uint64, error)
	FormatEventURL(id uint64, accessToken string) string
}

type recordKeeperClient struct {
	credentials Credentials
	httpClient  *apitools.HTTPClientWrapper
}

func (s *recordKeeperClient) authHeaderValue() string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s:%s", s.credentials.Username, s.credentials.Password)))

	authHash := fmt.Sprintf("%s %x", server.SHA256AuthorizationMethod, hasher.Sum(nil))
	hasher.Reset()

	return authHash
}

func (s *recordKeeperClient) FormatEventURL(id uint64, accessToken string) string {
	query := url.Values{
		server.EventAccessTokenKey: []string{accessToken},
	}

	return s.httpClient.Endpoint.FormatQuery(query, "event", fmt.Sprintf("%d", id))
}

func (s *recordKeeperClient) CreateEventRecord(req rkdb.CreateEventRecord) (uint64, error) {
	headers := http.Header{
		server.AuthorizationHeaderKey: []string{s.authHeaderValue()},
	}

	if output, err := json.Marshal(req); err != nil {
		return 0, err
	} else if resp, err := s.httpClient.POST(bytes.NewBuffer(output), nil, headers, "event"); err != nil {
		return 0, err
	} else {
		defer resp.Body.Close()

		var newRecord rkdb.EventRecord

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return 0, fmt.Errorf("response error %d", resp.StatusCode)
		} else if input, err := ioutil.ReadAll(resp.Body); err != nil {
			return 0, fmt.Errorf("failed to read response body %v", err)
		} else if err := json.Unmarshal(input, &newRecord); err != nil {
			return 0, fmt.Errorf("failed to unmarshal response body %v", err)
		} else {
			return newRecord.ID, nil
		}
	}
}
