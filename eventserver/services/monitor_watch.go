package services

import (
	"github.com/zinic/forculus/eventserver"
	"github.com/zinic/forculus/service"
	"sync"
	"time"

	"github.com/zinic/forculus/log"

	"github.com/zinic/forculus/zoneminder/zmapi"
)

type MonitorWatch struct {
	client     zmapi.Client
	dispatcher eventserver.EventDispatch
	exitC      chan struct{}
}

func NewMonitorWatch(client zmapi.Client, dispatch eventserver.EventDispatch) service.Service {
	return &MonitorWatch{
		client:     client,
		dispatcher: dispatch,
		exitC:      make(chan struct{}),
	}
}

func (s *MonitorWatch) monitorWatchLoop() {
	const (
		scanInterval = time.Second * 2
	)

	var (
		loopTicker      = time.NewTicker(scanInterval)
		watchedMonitors = make(map[string]zmapi.AlertedMonitor)
	)

	defer loopTicker.Stop()

	log.Info("Beginning monitor watch")

	for done := false; !done; {
		alertedMonitors, errList := s.client.AlertedMonitors()

		// Capture the errors that may have occurred while enumerating the alert status
		// of our watched monitors
		if errList != nil {
			for _, err := range errList {
				log.Errorf("Error during alerted monitor enumeration: %v", err)
			}

			continue
		}

		for monitorID, alertedMonitor := range alertedMonitors {
			if lastWatch, watching := watchedMonitors[monitorID]; !watching {
				watchedMonitors[monitorID] = alertedMonitor

				s.dispatcher.Send(eventserver.Event{
					Type:    eventserver.MonitorAlerted,
					Payload: alertedMonitor,
				})
			} else if lastWatch.AlarmStatus != alertedMonitor.AlarmStatus {
				watchedMonitors[monitorID] = alertedMonitor

				s.dispatcher.Send(eventserver.Event{
					Type:    eventserver.MonitorAlertStatusChanged,
					Payload: alertedMonitor,
				})
			}
		}

		for monitorID, watchedMonitor := range watchedMonitors {
			if _, stillAlerted := alertedMonitors[monitorID]; !stillAlerted {
				delete(watchedMonitors, monitorID)

				s.dispatcher.Send(eventserver.Event{
					Type:    eventserver.MonitorExitingAlert,
					Payload: watchedMonitor,
				})
			}
		}

		select {
		case <-loopTicker.C:
		case <-s.exitC:
			done = true
		}
	}
}

func (s *MonitorWatch) Start(waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)

	go func() {
		s.monitorWatchLoop()
		waitGroup.Done()
	}()
}

func (s *MonitorWatch) Stop() {
	close(s.exitC)
}
