package actor

import (
	"sync"

	"github.com/zinic/forculus/eventserver/service"

	"github.com/zinic/forculus/eventserver/event"
)

type dispatcher struct {
	subscriber     Subscriber
	eventInterests map[event.Type]struct{}
}

func (s dispatcher) Dispatch(event event.Event) {
	if s.Accepts(event.Type) {
		s.subscriber.Handle(event)
	}
}

func (s dispatcher) Accepts(eventType event.Type) bool {
	if _, acceptsAllEvents := s.eventInterests[event.All]; !acceptsAllEvents {
		_, acceptsEvent := s.eventInterests[eventType]
		return acceptsEvent
	}

	return true
}

func NewReactor(manager *service.Manager) Reactor {
	return &reactor{
		manager:      manager,
		dispatchLock: &sync.Mutex{},
	}
}

type reactor struct {
	manager       *service.Manager
	dispatchLock  *sync.Mutex
	subscriptions []dispatcher
}

func (s *reactor) Stop() {
	s.manager.Stop()
}

func (s *reactor) Register(subscriber Subscriber, eventInterests ...event.Type) {
	eventInterestMap := make(map[event.Type]struct{}, len(eventInterests))
	for _, eventInterest := range eventInterests {
		eventInterestMap[eventInterest] = struct{}{}
	}

	s.subscriptions = append(s.subscriptions, dispatcher{
		subscriber:     subscriber,
		eventInterests: eventInterestMap,
	})

	s.manager.Start(subscriber)
}

func (s *reactor) Dispatch(event event.Event) {
	s.dispatchLock.Lock()
	defer s.dispatchLock.Unlock()

	for _, subscription := range s.subscriptions {
		subscription.Dispatch(event)
	}
}
