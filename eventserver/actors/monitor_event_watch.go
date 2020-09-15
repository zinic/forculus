package actors

import (
	"time"

	"github.com/zinic/forculus/eventserver"

	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/zoneminder/zmapi"
)

func RegisterMonitorEventWatch(reactor eventserver.SubscriptionManager, client zmapi.Client) {
	watcher := &MonitorEventWatch{
		watchedMonitors: make(map[string]time.Time),
		seenEvents:      make(map[string]time.Time),
		dispatcher:      reactor,
		client:          client,
	}
	watcher.loadMostRecentEvents()

	reactor.Register(watcher.Logic, eventserver.MonitorExitingAlert)
}

type MonitorEventWatch struct {
	watchedMonitors map[string]time.Time
	seenEvents      map[string]time.Time
	dispatcher      eventserver.EventDispatch
	client          zmapi.Client
}

func (s *MonitorEventWatch) loadMostRecentEvents() {
	const searchWindowDuration = time.Minute * 30

	var (
		end   = time.Now()
		start = end.Add(-searchWindowDuration)
	)

	for {
		log.Infof("Loading most recent events")

		if monitorEvents, err := s.client.ListEventsBetween(start, end); err != nil {
			log.Errorf("Failed to load most recent events: %v", err)
			time.Sleep(time.Second * 5)
		} else {
			for _, monitorEvent := range monitorEvents {
				if endTime, err := monitorEvent.ParseEndTime(); err != nil {
					log.Fatalf("Possible API incompatibility! Failed to parse event time %s for event %s: %v", monitorEvent.EndTime, monitorEvent.ID, err)
				} else {
					s.seenEvents[monitorEvent.ID] = endTime
				}
			}

			break
		}
	}

	log.Infof("Most recent events loaded")
}

func (s *MonitorEventWatch) cleanupSeenEvents() {
	const (
		seenEventTTL = time.Minute * 30
	)

	now := time.Now()

	for eventID, endTime := range s.seenEvents {
		if endTime.Sub(now) >= seenEventTTL {
			delete(s.seenEvents, eventID)
		}
	}
}

func (s *MonitorEventWatch) Logic(eventC <-chan eventserver.Event, exitC chan struct{}) {
	const (
		scanInterval = time.Second * 2
		searchWindow = -time.Second * 30
	)

	loopTicker := time.NewTicker(scanInterval)
	defer loopTicker.Stop()

	for {
		select {
		case nextEvent := <-eventC:
			alertedMonitor := nextEvent.Payload.(zmapi.AlertedMonitor)
			s.watchedMonitors[alertedMonitor.Monitor.Details.ID] = time.Now().Add(searchWindow)

		case <-loopTicker.C:
			now := time.Now()

			for monitorID, watchStart := range s.watchedMonitors {
				if monitorEvents, err := s.client.ListMonitorEvents(monitorID, watchStart, now); err != nil {
					log.Errorf("Failed to list monitor events for monitor %s: %v", monitorID, err)
				} else {
					for _, monitorEvent := range monitorEvents {
						if _, seen := s.seenEvents[monitorEvent.ID]; !seen {
							delete(s.watchedMonitors, monitorID)

							s.dispatcher.Send(eventserver.Event{
								Type:    eventserver.MonitorNewEvent,
								Payload: monitorEvent,
							})

							s.seenEvents[monitorEvent.ID] = now
						}
					}
				}
			}

			s.cleanupSeenEvents()

		case <-exitC:
			return
		}
	}
}
