package zmapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zinic/forculus/apitools"

	"github.com/zinic/forculus/zoneminder/constants"
	"golang.org/x/net/html"
)

type Client interface {
	Login() error
	LoginSession() LoginSession
	RefreshLogin() error
	Monitors() (MonitorList, error)
	ExportEvent(event MonitorEvent) (io.ReadCloser, error)
	DownloadMP4(event MonitorEvent) (io.ReadCloser, error)
	AlarmStatus(monitor Monitor) (AlarmStatus, error)
	ListEvents() (EventList, error)
	ListEventsBetween(start, end time.Time) (EventList, error)
	ListMonitorEvents(monitorID string, start, end time.Time) (EventList, error)
	Version() (Version, error)
	AlertedMonitors() (map[string]AlertedMonitor, []error)
}

func NewClient(endpoint apitools.Endpoint, credentials LoginCredentials) Client {
	return &client{
		credentials: credentials,
		httpClient:  apitools.NewHTTPClientWrapper(endpoint),
	}
}

type client struct {
	credentials  LoginCredentials
	loginSession *LoginSession
	httpClient   *apitools.HTTPClientWrapper
}

func (s *client) doGET(body io.Reader, query url.Values, header http.Header, path ...string) (*http.Response, error) {
	queryCopy := apitools.CopyURLValues(query)

	if _, hasToken := queryCopy["token"]; !hasToken && s.loginSession != nil {
		queryCopy.Set("token", s.loginSession.Details.AccessToken)
	}

	return s.httpClient.GET(body, queryCopy, header, path...)
}

func (s *client) doPOST(body io.Reader, query url.Values, header http.Header, path ...string) (*http.Response, error) {
	queryCopy := apitools.CopyURLValues(query)

	if _, hasToken := queryCopy["token"]; !hasToken && s.loginSession != nil {
		queryCopy.Set("token", s.loginSession.Details.AccessToken)
	}

	return s.httpClient.POST(body, queryCopy, header, path...)
}

func (s *client) checkLogin() error {
	if s.loginSession == nil {
		return s.Login()
	}

	if s.loginSession.Expired() {
		return s.Login()
	}

	if s.loginSession.RefreshRequired() {
		return s.RefreshLogin()
	}

	return nil
}

func (s *client) exportEventCSRF(event MonitorEvent) (string, error) {
	formQuery := url.Values{
		"view": []string{"export"},
		"eid":  []string{event.ID},
	}

	if formResp, err := s.doGET(nil, formQuery, nil, "index.php"); err != nil {
		return "", err
	} else {
		defer formResp.Body.Close()

		if document, err := html.Parse(formResp.Body); err != nil {
			return "", err
		} else if csrfToken, hasToken := ExtractCSRFToken(document); !hasToken {
			return "", fmt.Errorf("failed to find CSRF token in HTML form")
		} else {
			return csrfToken, nil
		}
	}
}

func (s *client) LoginSession() LoginSession {
	return *s.loginSession
}

func (s *client) ExportEvent(event MonitorEvent) (io.ReadCloser, error) {
	if err := s.checkLogin(); err != nil {
		return nil, err
	}

	var (
		connKey = strconv.Itoa(rand.Int() % 1000000)
		form    = url.Values{
			"view":           []string{"event"},
			"request":        []string{"event"},
			"action":         []string{"export"},
			"connkey":        []string{connKey},
			"eids[]":         []string{event.ID},
			"exportDetail":   []string{"1"},
			"exportFrames":   []string{"1"},
			"exportImages":   []string{"1"},
			"exportVideo":    []string{"1"},
			"exportMisc":     []string{"1"},
			"exportFormat":   []string{"tar"},
			"exportCompress": []string{"1"},
			"exportFile":     nil,
			"generated":      nil,
		}

		header = http.Header{
			"Content-Type": []string{"application/x-www-form-urlencoded"},
		}

		exportEventResult ExportEventResult
	)

	// Gross...
	if csrfToken, err := s.exportEventCSRF(event); err != nil {
		return nil, err
	} else {
		form[constants.CSRFMagicName] = []string{csrfToken, csrfToken}
	}

	if exportResp, err := s.doPOST(strings.NewReader(form.Encode()), nil, header, "index.php"); err != nil {
		return nil, err
	} else {
		defer exportResp.Body.Close()

		if input, err := ioutil.ReadAll(exportResp.Body); err != nil {
			return nil, err
		} else if err := json.Unmarshal(input, &exportEventResult); err != nil {
			return nil, fmt.Errorf("failed parsing %s: %w", input, err)
		} else if len(exportEventResult.ExportFile) == 0 {
			return nil, fmt.Errorf("no files exported: %s", input)
		}

		if unescapedQuery, err := url.QueryUnescape(exportEventResult.ExportFile[1:]); err != nil {
			return nil, err
		} else if exportValues, err := url.ParseQuery(unescapedQuery); err != nil {
			return nil, fmt.Errorf("failed to parse query %s: %w", exportEventResult.ExportFile, err)
		} else if fileResp, err := s.doGET(nil, exportValues, nil, "index.php"); err != nil {
			return nil, err
		} else {
			return fileResp.Body, nil
		}
	}
}

func (s *client) DownloadMP4(event MonitorEvent) (io.ReadCloser, error) {
	if err := s.checkLogin(); err != nil {
		return nil, err
	}

	params := url.Values{
		"eid":  []string{event.ID},
		"view": []string{"view_video"},
		"mode": []string{"mp4"},
	}

	if resp, err := s.doGET(nil, params, nil, "index.php"); err != nil {
		return nil, err
	} else {
		return resp.Body, nil
	}
}

func (s *client) AlertedMonitors() (map[string]AlertedMonitor, []error) {
	var (
		errorList       []error
		changedMonitors = make(map[string]AlertedMonitor)
	)

	if monitors, err := s.Monitors(); err != nil {
		return nil, []error{err}
	} else {
		for _, monitor := range monitors {
			if alarmStatus, err := s.AlarmStatus(monitor); err != nil {
				errorList = append(errorList, fmt.Errorf("failed to fetch alarm status for monitor %s: %w", monitor.Name(), err))
			} else {
				switch alarmStatus {
				case AlarmStatusPreAlarm:
					fallthrough
				case AlarmStatusAlert:
					fallthrough
				case AlarmStatusAlarm:
					changedMonitors[monitor.Details.ID] = AlertedMonitor{
						Monitor:     monitor,
						AlarmStatus: alarmStatus,
					}
				}
			}
		}
	}

	return changedMonitors, errorList
}

func (s *client) Monitors() (MonitorList, error) {
	if err := s.checkLogin(); err != nil {
		return nil, err
	}

	var listMonitorsResponse ListMonitorsResponse
	if resp, err := s.doGET(nil, nil, nil, "api", "monitors.json"); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()

		if content, err := ioutil.ReadAll(resp.Body); err != nil {
			return nil, err
		} else if err := json.Unmarshal(content, &listMonitorsResponse); err != nil {
			return nil, err
		}
	}

	return listMonitorsResponse.Monitors, nil
}

func (s *client) AlarmStatus(monitor Monitor) (AlarmStatus, error) {
	if err := s.checkLogin(); err != nil {
		return AlarmStatusInvalid, err
	}

	var monitorAlarmStatus MonitorAlarmStatus
	if resp, err := s.doGET(nil, nil, nil, "api", "monitors", "alarm", fmt.Sprintf("id:%s", monitor.Details.ID), "command:status.json"); err != nil {
		return AlarmStatusInvalid, err
	} else {
		defer resp.Body.Close()

		if content, err := ioutil.ReadAll(resp.Body); err != nil {
			return AlarmStatusInvalid, err
		} else if err := json.Unmarshal(content, &monitorAlarmStatus); err != nil {
			return AlarmStatusInvalid, err
		}
	}

	return ParseAlarmStatus(monitorAlarmStatus.Status), nil
}

func (s *client) ListMonitorEvents(monitorID string, start, end time.Time) (EventList, error) {
	if err := s.checkLogin(); err != nil {
		return nil, err
	}

	var (
		listEventsResponse ListEventsResponse
		events             EventList

		monitorPath = fmt.Sprintf("MonitorId:%s", monitorID)
		startTime   = fmt.Sprintf("StartTime >=:%s", start.Format(constants.ZMDateFormat))
		endTime     = fmt.Sprintf("EndTime <=:%s.json", end.Format(constants.ZMDateFormat))
		query       = url.Values{}
		page        = 1
	)

	for {
		query.Set("page", strconv.Itoa(page))
		page += 1

		if resp, err := s.doGET(nil, query, nil, "api", "events", "index", monitorPath, startTime, endTime); err != nil {
			return nil, err
		} else if content, err := ioutil.ReadAll(resp.Body); err != nil {
			resp.Body.Close()
			return nil, err
		} else if err := json.Unmarshal(content, &listEventsResponse); err != nil {
			resp.Body.Close()
			return nil, err
		} else {
			resp.Body.Close()

			for _, eventWrapper := range listEventsResponse.Events {
				events = append(events, eventWrapper.Event)
			}

			if !listEventsResponse.Pagination.NextPage {
				break
			}
		}
	}

	return events, nil
}

func (s *client) ListEventsBetween(start, end time.Time) (EventList, error) {
	if err := s.checkLogin(); err != nil {
		return nil, err
	}

	var (
		listEventsResponse ListEventsResponse
		events             EventList

		startTime = fmt.Sprintf("StartTime >=:%s", start.Format(constants.ZMDateFormat))
		endTime   = fmt.Sprintf("EndTime <=:%s.json", end.Format(constants.ZMDateFormat))
		query     = url.Values{}
		page      = 1
	)

	for {
		query.Set("page", strconv.Itoa(page))
		page += 1

		if resp, err := s.doGET(nil, query, nil, "api", "events", "index", startTime, endTime); err != nil {
			return nil, err
		} else if content, err := ioutil.ReadAll(resp.Body); err != nil {
			resp.Body.Close()
			return nil, err
		} else if err := json.Unmarshal(content, &listEventsResponse); err != nil {
			resp.Body.Close()
			return nil, err
		} else {
			resp.Body.Close()

			for _, eventWrapper := range listEventsResponse.Events {
				events = append(events, eventWrapper.Event)
			}

			if !listEventsResponse.Pagination.NextPage {
				break
			}
		}
	}

	return events, nil
}

func (s *client) ListEvents() (EventList, error) {
	if err := s.checkLogin(); err != nil {
		return nil, err
	}

	var (
		listEventsResponse ListEventsResponse
		events             EventList

		query = url.Values{}
		page  = 1
	)

	for {
		query.Set("page", strconv.Itoa(page))
		page += 1

		if resp, err := s.doGET(nil, query, nil, "api", "events.json"); err != nil {
			return nil, err
		} else if content, err := ioutil.ReadAll(resp.Body); err != nil {
			resp.Body.Close()
			return nil, err
		} else if err := json.Unmarshal(content, &listEventsResponse); err != nil {
			resp.Body.Close()
			return nil, err
		} else {
			resp.Body.Close()

			for _, eventWrapper := range listEventsResponse.Events {
				events = append(events, eventWrapper.Event)
			}

			if !listEventsResponse.Pagination.NextPage {
				break
			}
		}
	}

	return events, nil
}

func (s *client) Version() (Version, error) {
	var version Version

	if err := s.checkLogin(); err != nil {
		return version, err
	}

	if resp, err := s.doGET(nil, nil, nil, "api", "host", "getVersion.json"); err != nil {
		return version, err
	} else {
		defer resp.Body.Close()

		if content, err := ioutil.ReadAll(resp.Body); err != nil {
			return version, err
		} else if err := json.Unmarshal(content, &version); err != nil {
			return version, err
		}
	}

	return version, nil
}

func (s *client) RefreshLogin() error {
	query := url.Values{
		"token": []string{s.loginSession.Details.RefreshToken},
	}

	if resp, err := s.doGET(nil, query, nil, "api", "host", "login.json"); err != nil {
		return err
	} else {
		defer resp.Body.Close()

		var refreshDetails LoginDetails

		if content, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else if err := json.Unmarshal(content, &refreshDetails); err != nil {
			return err
		} else {
			s.loginSession.Refresh(refreshDetails)
		}
	}

	return nil
}

func (s *client) Login() error {
	var (
		form         = make(url.Values)
		header       = make(http.Header)
		loginDetails LoginDetails
	)

	form.Set("user", s.credentials.Username)
	form.Set("pass", s.credentials.Password)

	header.Set("Content-Type", "application/x-www-form-urlencoded")

	if resp, err := s.doPOST(strings.NewReader(form.Encode()), nil, header, "api", "host", "login.json"); err != nil {
		return err
	} else {
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("request failed with response code %s", resp.Status)
		}

		now := time.Now()
		if content, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else if err := json.Unmarshal(content, &loginDetails); err != nil {
			return err
		} else {
			s.loginSession = &LoginSession{
				Details:     loginDetails,
				Created:     now,
				LastRefresh: now,
			}
		}
	}

	return nil
}
