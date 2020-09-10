package logic

import (
	"fmt"

	"github.com/zinic/forculus/actor"

	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/email"
	"github.com/zinic/forculus/event"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/zoneminder/api"
)

func RegisterEmailer(reactor actor.Reactor, alert config.EmailAlert, server config.SMTPServer) {
	emailSender := EventEmailSender{
		alert:  alert,
		server: server,
	}

	reactor.Register(actor.NewSubscriber(emailSender.Logic), event.All)
}

type EventEmailSender struct {
	alert  config.EmailAlert
	server config.SMTPServer
}

func (s *EventEmailSender) handleEvent(nextEvent event.Event) {
	if s.alert.Filter.EventTrigger != "" && nextEvent.Type != s.alert.Filter.EventTrigger {
		return
	}

	switch nextEvent.Type {
	case event.MonitorAlerted:
		alertedMonitor := nextEvent.Payload.(api.AlertedMonitor)
		if alertFrames, err := alertedMonitor.Monitor.Details.ParseAlertFrameCount(); err != nil {
			log.Errorf("Failed to parse alert frame count for monitor %s: %v", alertedMonitor.Monitor.Details.Name, err)
		} else if s.alert.Filter.AlertFrameThreshold > 0 && s.alert.Filter.AlertFrameThreshold > alertFrames {
			return
		}

		if s.alert.Filter.NameRegex != nil && !s.alert.Filter.NameRegex.MatchString(alertedMonitor.Monitor.Details.Name) {
			log.Debugf("Alerted monitor %s did not match alert %s regex %s",
				alertedMonitor.Monitor.Details.Name, s.alert.Name, s.alert.Filter.NameRegex)
			return
		}

		body := fmt.Sprintf("Monitor %s has become alerted.", alertedMonitor.Monitor.Name())
		if err := email.Send(body, s.alert, s.server); err != nil {
			log.Errorf("Failed sending email for alert %s: %v", s.alert.Name, err)
		} else {
			log.Infof("Email alert %s triggered", s.alert.Name)
		}

	case event.NewMonitorEvent:
		monitorEvent := nextEvent.Payload.(api.MonitorEvent)
		if s.alert.Filter.NameRegex != nil && !s.alert.Filter.NameRegex.MatchString(monitorEvent.Name) {
			log.Debugf("Monitor event %s did not match alert %s regex %s",
				monitorEvent.Name, s.alert.Name, s.alert.Filter.NameRegex)
			return
		}

		if alertFrames, err := monitorEvent.ParseAlertFrames(); err != nil {
			log.Errorf("Failed to parse alert frame count for monitor event %s: %v", monitorEvent.Name, err)
		} else if s.alert.Filter.AlertFrameThreshold > 0 && s.alert.Filter.AlertFrameThreshold > alertFrames {
			return
		}

		body := fmt.Sprintf("A new monitor event %s has become available.", monitorEvent.Name)
		if err := email.Send(body, s.alert, s.server); err != nil {
			log.Errorf("Failed sending email for alert %s: %v", s.alert.Name, err)
		} else {
			log.Infof("Email alert %s triggered", s.alert.Name)
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
