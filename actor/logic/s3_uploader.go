package logic

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/zinic/forculus/actor"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/zinic/forculus/aws"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/event"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/zoneminder/api"
)

func RegisterEventS3Uploader(reactor actor.Reactor, zmClient api.Client, cfg config.S3Uploader) {
	eventDownloader := EventS3Uploader{
		cfg:      cfg,
		zmClient: zmClient,
		s3Client: aws.NewS3Client(cfg),
	}

	reactor.Register(actor.NewSubscriber(eventDownloader.Logic), event.NewMonitorEvent)
}

type EventS3Uploader struct {
	s3Client *s3.Client
	zmClient api.Client
	cfg      config.S3Uploader
}

func (s *EventS3Uploader) writeS3File(filename string) error {
	log.Infof("Uploading exported event %s to S3", filename)

	if fin, err := os.Open(filename); err != nil {
		return err
	} else {
		defer fin.Close()

		putRequest := s.s3Client.PutObjectRequest(&s3.PutObjectInput{
			Body:   fin,
			Bucket: &s.cfg.Bucket,
			Key:    &filename,
		})

		_, err := putRequest.Send(context.Background())
		return err
	}
}

func (s *EventS3Uploader) writeLocalFile(filename string, source io.ReadCloser) error {
	defer source.Close()

	log.Infof("Writing exported event %s to file", filename)

	if fout, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		return fmt.Errorf("failed to open event file %s: %w", filename, err)
	} else {
		defer fout.Close()

		if _, err := io.Copy(fout, source); err != nil {
			return fmt.Errorf("failed to copy stream to file: %w", err)
		}
	}

	return nil
}

func (s *EventS3Uploader) Logic(eventC chan event.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			monitorEvent := nextEvent.Payload.(api.MonitorEvent)

			if s.cfg.Filter.NameRegex != nil && !s.cfg.Filter.NameRegex.MatchString(monitorEvent.Name) {
				continue
			}

			if alertFrames, err := monitorEvent.ParseAlertFrames(); err != nil {
				log.Errorf("Failed to parse alert frames for event %s: %v", monitorEvent.Name, err)
				continue
			} else if s.cfg.Filter.AlertFrameThreshold > 0 && s.cfg.Filter.AlertFrameThreshold > alertFrames {
				continue
			}

			eventFilename := fmt.Sprintf("%s.tar.gz", monitorEvent.Name)
			log.Infof("Exporting event %s", monitorEvent.Name)

			if eventExportStream, err := s.zmClient.ExportEvent(monitorEvent); err != nil {
				log.Errorf("Failed to download the MP4 video: %v", err)
			} else if err := s.writeLocalFile(eventFilename, eventExportStream); err != nil {
				log.Errorf("Error during local write: %v", err)
			} else if err := s.writeS3File(eventFilename); err != nil {
				log.Errorf("Error during S3 upload: %v", err)
			} else if err := os.Remove(eventFilename); err != nil {
				log.Errorf("Failed to remove local event copy %s: %v", eventFilename, err)
			} else {
				log.Infof("Exported event %s uploaded to S3 successfully", eventFilename)
			}

		case <-exitC:
			return
		}
	}
}
