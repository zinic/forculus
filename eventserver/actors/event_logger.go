package actors

import (
	"github.com/zinic/forculus/eventserver"
	"github.com/zinic/forculus/log"

	"github.com/zinic/forculus/zoneminder/zmapi"
)

func EventLogger(eventC <-chan eventserver.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			switch nextEvent.Type {
			case eventserver.MonitorAlerted:
				alertedMonitor := nextEvent.Payload.(zmapi.AlertedMonitor)
				log.Infof("Monitor %s has entered alert status %s", alertedMonitor.Monitor.Name(), alertedMonitor.AlarmStatus)

			case eventserver.MonitorAlertStatusChanged:
				alertedMonitor := nextEvent.Payload.(zmapi.AlertedMonitor)
				log.Infof("Monitor %s alert status has changed to %s", alertedMonitor.Monitor.Name(), alertedMonitor.AlarmStatus)

			case eventserver.MonitorExitingAlert:
				alertedMonitor := nextEvent.Payload.(zmapi.AlertedMonitor)
				log.Infof("Monitor %s has exited alert status", alertedMonitor.Monitor.Name())

			case eventserver.MonitorNewEvent:
				monitorEvent := nextEvent.Payload.(zmapi.MonitorEvent)
				log.Infof("New monitor event %s has been created", monitorEvent.Name)
			}

		case <-exitC:
			return
		}
	}
}

func RegisterEventLogger(reactor eventserver.SubscriptionManager) {
	reactor.Register(EventLogger, eventserver.All)
}
