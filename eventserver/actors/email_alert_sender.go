package actors

import (
	"fmt"
	"github.com/zinic/forculus/eventserver"

	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/email"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/zoneminder/zmapi"
)

func RegisterEventEmailSender(reactor eventserver.SubscriptionManager, name string, alert config.EmailAlert, server config.SMTPServer) {
	emailSender := &EventEmailSender{
		name:   name,
		alert:  alert,
		server: server,
	}

	reactor.Register(emailSender.Logic, eventserver.All)
}

type EventEmailSender struct {
	name   string
	alert  config.EmailAlert
	server config.SMTPServer
}

func (s *EventEmailSender) handleEvent(nextEvent eventserver.Event) {
	if s.alert.Filter.EventTrigger != "" && nextEvent.Type != s.alert.Filter.EventTrigger {
		return
	}

	switch nextEvent.Type {
	case eventserver.MonitorAlerted:
		alertedMonitor := nextEvent.Payload.(zmapi.AlertedMonitor)
		if alertFrames, err := alertedMonitor.Monitor.Details.ParseAlertFrameCount(); err != nil {
			log.Errorf("Failed to parse alert frame count for monitor %s: %v", alertedMonitor.Monitor.Details.Name, err)
			return
		} else if s.alert.Filter.AlertFrameThreshold > 0 && s.alert.Filter.AlertFrameThreshold > alertFrames {
			return
		}

		if s.alert.Filter.NameRegex != nil && !s.alert.Filter.NameRegex.MatchString(alertedMonitor.Monitor.Details.Name) {
			log.Debugf("Alerted monitor %s did not match alert %s regex %s",
				alertedMonitor.Monitor.Details.Name, s.name, s.alert.Filter.NameRegex)
			return
		}

		emailTemplate := email.Email{
			Subject:    s.alert.SubjectTemplate,
			Body:       fmt.Sprintf("Monitor %s has become alerted.", alertedMonitor.Monitor.Name()),
			Recipients: s.alert.Recipients,
		}

		if err := email.Send(emailTemplate, s.server); err != nil {
			log.Errorf("Failed sending email for alert %s: %v", s.name, err)
		} else {
			log.Infof("Email alert %s triggered", s.name)
		}

	case eventserver.MonitorEventRecorded:
		eventRecordedPayload := nextEvent.Payload.(MonitorEventRecordedPayload)
		if s.alert.Filter.NameRegex != nil && !s.alert.Filter.NameRegex.MatchString(eventRecordedPayload.Source.Name) {
			log.Debugf("Monitor event %s did not match alert %s regex %s",
				eventRecordedPayload.Source.Name, s.name, s.alert.Filter.NameRegex)
			return
		}

		if alertFrames, err := eventRecordedPayload.Source.ParseAlertFrames(); err != nil {
			log.Errorf("Failed to parse alert frame count for monitor event ass %s: %v", eventRecordedPayload.Source.Name, err)
		} else if s.alert.Filter.AlertFrameThreshold > 0 && s.alert.Filter.AlertFrameThreshold > alertFrames {
			return
		}

		body := fmt.Sprintf("A new monitor event %s (%s) has become available.", eventRecordedPayload.Source.Name, eventRecordedPayload.AccessURL)
		emailTemplate := email.Email{
			Subject:    s.alert.SubjectTemplate,
			Body:       body,
			Recipients: s.alert.Recipients,
		}

		if err := email.Send(emailTemplate, s.server); err != nil {
			log.Errorf("Failed sending email for alert %s: %v", s.name, err)
		} else {
			log.Infof("Email alert %s triggered", s.name)
		}
	}

}

func (s *EventEmailSender) Logic(eventC <-chan eventserver.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			s.handleEvent(nextEvent)

		case <-exitC:
			return
		}
	}
}
