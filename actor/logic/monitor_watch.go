package logic

import (
	"sync"
	"time"

	"github.com/zinic/forculus/actor"

	"github.com/zinic/forculus/event"

	"github.com/zinic/forculus/log"

	"github.com/zinic/forculus/zoneminder/api"
)

type MonitorWatch struct {
	client     api.Client
	dispatcher actor.Dispatch
	exitC      chan struct{}
}

func NewMonitorWatch(client api.Client, dispatch actor.Dispatch) actor.Service {
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
		watchedMonitors = make(map[string]api.AlertedMonitor)
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

				s.dispatcher.Dispatch(event.Event{
					Type:    event.MonitorAlerted,
					Payload: alertedMonitor,
				})
			} else if lastWatch.AlarmStatus != alertedMonitor.AlarmStatus {
				watchedMonitors[monitorID] = alertedMonitor

				s.dispatcher.Dispatch(event.Event{
					Type:    event.MonitorAlertStatusChanged,
					Payload: alertedMonitor,
				})
			}
		}

		for monitorID, watchedMonitor := range watchedMonitors {
			if _, stillAlerted := alertedMonitors[monitorID]; !stillAlerted {
				delete(watchedMonitors, monitorID)

				s.dispatcher.Dispatch(event.Event{
					Type:    event.MonitorExitingAlert,
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
