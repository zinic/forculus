package actors

import (
	"math/rand"

	"github.com/zinic/forculus/apitools"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/eventserver"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/recordkeeper/rkapi"
	"github.com/zinic/forculus/recordkeeper/rkdb"
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

func NewRecordKeeper(dispatch eventserver.EventDispatch, cfg config.RecordKeeperClient) eventserver.EventHandlerFunc {
	var (
		endpoint    = apitools.NewEndpoint(cfg.Scheme, cfg.Host, cfg.Port, "")
		credentials = rkapi.Credentials{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	)

	rk := &RecordKeeper{
		cfg:      cfg,
		dispatch: dispatch,
		endpoint: endpoint,
		client:   rkapi.NewClient(credentials, endpoint),
	}

	return rk.Logic
}

type RecordKeeper struct {
	cfg      config.RecordKeeperClient
	dispatch eventserver.EventDispatch
	endpoint apitools.Endpoint
	client   rkapi.Client
}

func (s *RecordKeeper) Logic(eventC <-chan eventserver.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			var (
				eventUploadedPayload = nextEvent.Payload.(MonitorEventUploadedPayload)
				createRecordReq      = rkdb.CreateEventRecord{
					StorageTarget: eventUploadedPayload.StorageTarget,
					StorageKey:    eventUploadedPayload.StorageKey,
					AccessToken:   newAccessToken(),
				}
			)

			if newRecordID, err := s.client.CreateEventRecord(createRecordReq); err != nil {
				log.Errorf("Failed to create new event record via the record keeper API: %v")
			} else {
				s.dispatch.Send(eventserver.Event{
					Type: eventserver.MonitorEventRecorded,
					Payload: MonitorEventRecordedPayload{
						Source:    eventUploadedPayload.Source,
						AccessURL: s.client.FormatEventURL(newRecordID, createRecordReq.AccessToken),
					},
				})
			}

		case <-exitC:
			return
		}
	}

}
