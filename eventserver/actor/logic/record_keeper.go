package logic

import (
	"github.com/zinic/forculus/apitools"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/eventserver/actor"
	"github.com/zinic/forculus/eventserver/email"
	"github.com/zinic/forculus/eventserver/event"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/recordkeeper/model"
	"github.com/zinic/forculus/recordkeeper/rkapi"
	"math/rand"
	"strings"
)

const (
	accessTokenLength = 17
	charset           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func newAccessToken() string {
	buf := make([]byte, accessTokenLength)

	for idx := 0; idx < accessTokenLength; idx++ {
		buf[idx] = charset[rand.Intn(len(charset))]
	}

	return string(buf)
}

func NewRecordKeeper(dispatch actor.Dispatch, cfg config.RecordKeeperClient) RecordKeeper {
	var (
		endpoint    = apitools.NewEndpoint(cfg.Scheme, cfg.Host, cfg.Port, "")
		credentials = rkapi.Credentials{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	)

	return RecordKeeper{
		cfg:      cfg,
		dispatch: dispatch,
		endpoint: endpoint,
		client:   rkapi.NewClient(credentials, endpoint),
	}
}

type RecordKeeper struct {
	cfg      config.RecordKeeperClient
	dispatch actor.Dispatch
	endpoint apitools.Endpoint
	client   rkapi.Client
}

func (s *RecordKeeper) Logic(eventC chan event.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			eventRecord := nextEvent.Payload.(model.EventRecord)
			eventRecord.AccessToken = newAccessToken()

			if newRecord, err := s.client.CreateEventRecord(eventRecord); err != nil {
				log.Errorf("Failed to create new event record via the record keeper API: %v")
			} else {
				bodyWriter := strings.Builder{}
				bodyWriter.WriteString("New even available at record keeper: ")
				bodyWriter.WriteString(s.client.FormatEventURL(newRecord))
				bodyWriter.WriteRune('\n')

				s.dispatch.Dispatch(event.Event{
					Type: event.EmailNotice,
					Payload: email.Email{
						Subject:    "New Event in Record Keeper",
						Body:       bodyWriter.String(),
					},
				})
			}

		case <-exitC:
			return
		}
	}

}
