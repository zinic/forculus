package logic

import (
	"fmt"

	"github.com/zinic/forculus/eventserver/actor"

	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/eventserver/email"
	"github.com/zinic/forculus/eventserver/event"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/zoneminder/zmapi"
)

func RegisterEventEmailSender(reactor actor.Reactor, name string, alert config.EmailAlert, server config.SMTPServer) {
	emailSender := EventEmailSender{
		name:   name,
		alert:  alert,
		server: server,
	}

	reactor.Register(actor.NewSubscriber(emailSender.Logic), event.All)
}

type EventEmailSender struct {
	name   string
	alert  config.EmailAlert
	server config.SMTPServer
}

func (s *EventEmailSender) handleEvent(nextEvent event.Event) {
	if s.alert.Filter.EventTrigger != "" && nextEvent.Type != s.alert.Filter.EventTrigger {
		return
	}

	switch nextEvent.Type {
	case event.MonitorAlerted:
		alertedMonitor := nextEvent.Payload.(zmapi.AlertedMonitor)
		if alertFrames, err := alertedMonitor.Monitor.Details.ParseAlertFrameCount(); err != nil {
			log.Errorf("Failed to parse alert frame count for monitor %s: %v", alertedMonitor.Monitor.Details.Name, err)
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

	case event.NewMonitorEvent:
		monitorEvent := nextEvent.Payload.(zmapi.MonitorEvent)
		if s.alert.Filter.NameRegex != nil && !s.alert.Filter.NameRegex.MatchString(monitorEvent.Name) {
			log.Debugf("Monitor event %s did not match alert %s regex %s",
				monitorEvent.Name, s.name, s.alert.Filter.NameRegex)
			return
		}

		if alertFrames, err := monitorEvent.ParseAlertFrames(); err != nil {
			log.Errorf("Failed to parse alert frame count for monitor event %s: %v", monitorEvent.Name, err)
		} else if s.alert.Filter.AlertFrameThreshold > 0 && s.alert.Filter.AlertFrameThreshold > alertFrames {
			return
		}

		emailTemplate := email.Email{
			Subject:    s.alert.SubjectTemplate,
			Body:       fmt.Sprintf("A new monitor event %s has become available.", monitorEvent.Name),
			Recipients: s.alert.Recipients,
		}

		if err := email.Send(emailTemplate, s.server); err != nil {
			log.Errorf("Failed sending email for alert %s: %v", s.name, err)
		} else {
			log.Infof("Email alert %s triggered", s.name)
		}
	}

}

func (s *EventEmailSender) Logic(eventC chan event.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			s.handleEvent(nextEvent)

		case <-exitC:
			return
		}
	}
}
