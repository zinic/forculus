package actors

import (
	"fmt"

	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/eventserver"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/storage"
	"github.com/zinic/forculus/zoneminder/zmapi"
)

func NewUploader(name string, dispatch eventserver.EventDispatch, zmClient zmapi.Client, storageProvider storage.Provider, cfg config.Uploader) eventserver.EventHandlerFunc {
	uploader := &EventUploader{
		name:            name,
		dispatch:        dispatch,
		zmClient:        zmClient,
		storageProvider: storageProvider,
		cfg:             cfg,
	}

	return uploader.Logic
}

type EventUploader struct {
	name            string
	dispatch        eventserver.EventDispatch
	zmClient        zmapi.Client
	storageProvider storage.Provider
	cfg             config.Uploader
}

func (s *EventUploader) Logic(eventC <-chan eventserver.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			monitorEvent := nextEvent.Payload.(zmapi.MonitorEvent)

			if s.cfg.Filter.NameRegex != nil && !s.cfg.Filter.NameRegex.MatchString(monitorEvent.Name) {
				log.Debugf("Event %s does not match the name regex filter for exporter %s", monitorEvent.Name, s.name)
				continue
			}

			if alertFrames, err := monitorEvent.ParseAlertFrames(); err != nil {
				log.Errorf("Failed to parse alert frames for event %s: %v", monitorEvent.Name, err)
				continue
			} else if s.cfg.Filter.AlertFrameThreshold > 0 && s.cfg.Filter.AlertFrameThreshold > alertFrames {
				log.Debugf("Event %s does not meet the alert frame threshold for exporter %s", monitorEvent.Name, s.name)
				continue
			}

			eventFilename := fmt.Sprintf("%s.tar.gz", monitorEvent.Name)
			log.Infof("Exporting event %s", monitorEvent.Name)

			if eventExportStream, err := s.zmClient.ExportEvent(monitorEvent); err != nil {
				log.Errorf("Failed to download the MP4 video: %v", err)
			} else {
				if err := s.storageProvider.Write(eventFilename, eventExportStream); err != nil {
					log.Errorf("Failed to upload event to storage provider: %v", err)
				}

				eventExportStream.Close()

				log.Infof("Event %s exported successfully", monitorEvent.Name)
			}

			s.dispatch.Send(eventserver.Event{
				Type: eventserver.MonitorEventUploaded,
				Payload: MonitorEventUploadedPayload{
					Source:        monitorEvent,
					StorageTarget: s.cfg.StorageTarget,
					StorageKey:    eventFilename,
				},
			})

		case <-exitC:
			return
		}
	}
}
