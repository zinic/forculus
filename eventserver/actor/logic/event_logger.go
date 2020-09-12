package logic

import (
	"github.com/zinic/forculus/eventserver/actor"
	"github.com/zinic/forculus/eventserver/event"

	"github.com/zinic/forculus/log"

	"github.com/zinic/forculus/zoneminder/api"
)

func EventLogger(eventC chan event.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			switch nextEvent.Type {
			case event.MonitorAlerted:
				alertedMonitor := nextEvent.Payload.(api.AlertedMonitor)
				log.Infof("Monitor %s has entered alert status %s", alertedMonitor.Monitor.Name(), alertedMonitor.AlarmStatus)

			case event.MonitorAlertStatusChanged:
				alertedMonitor := nextEvent.Payload.(api.AlertedMonitor)
				log.Infof("Monitor %s alert status has changed to %s", alertedMonitor.Monitor.Name(), alertedMonitor.AlarmStatus)

			case event.MonitorExitingAlert:
				alertedMonitor := nextEvent.Payload.(api.AlertedMonitor)
				log.Infof("Monitor %s has exited alert status", alertedMonitor.Monitor.Name())

			case event.NewMonitorEvent:
				monitorEvent := nextEvent.Payload.(api.MonitorEvent)
				log.Infof("New monitor event %s has been created", monitorEvent.Name)
			}

		case <-exitC:
			return
		}
	}
}

func RegisterEventLogger(reactor actor.Reactor) {
	reactor.Register(actor.NewSubscriber(EventLogger), event.All)
}
