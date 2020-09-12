package actor

import (
	"sync"

	"github.com/zinic/forculus/eventserver/event"
)

type subscription struct {
	subscriber     Subscriber
	eventInterests map[event.Type]struct{}
}

func (s subscription) Dispatch(event event.Event) {
	if s.Accepts(event.Type) {
		s.subscriber.Handle(event)
	}
}

func (s subscription) Accepts(eventType event.Type) bool {
	if _, acceptsAllEvents := s.eventInterests[event.All]; !acceptsAllEvents {
		_, acceptsEvent := s.eventInterests[eventType]
		return acceptsEvent
	}

	return true
}

func NewReactor() Reactor {
	return &reactor{
		subscriberLock: &sync.Mutex{},
	}
}

type reactor struct {
	subscribers    []subscription
	subscriberLock *sync.Mutex
}

func (s *reactor) Start(waitGroup *sync.WaitGroup) {
	for _, entry := range s.subscribers {
		entry.subscriber.Start(waitGroup)
	}
}

func (s *reactor) Stop() {
	for _, entry := range s.subscribers {
		entry.subscriber.Stop()
	}
}

func (s *reactor) Dispatch(event event.Event) {
	s.subscriberLock.Lock()
	defer s.subscriberLock.Unlock()

	for _, subscriber := range s.subscribers {
		subscriber.Dispatch(event)
	}
}

func (s *reactor) Register(subscriber Subscriber, eventInterests ...event.Type) {
	eventInterestMap := make(map[event.Type]struct{}, len(eventInterests))
	for _, eventInterest := range eventInterests {
		eventInterestMap[eventInterest] = struct{}{}
	}

	s.subscribers = append(s.subscribers, subscription{
		subscriber:     subscriber,
		eventInterests: eventInterestMap,
	})
}
