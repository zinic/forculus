package eventserver

import (
	"github.com/zinic/forculus/log"
	"sync"
)

const (
	defaultHandlerEventBuffer = 15
)

func NewSubscription(logic EventHandlerFunc, interests ...EventType) *subscription {
	interestsMap := make(map[EventType]struct{}, len(interests))
	for _, eventInterest := range interests {
		interestsMap[eventInterest] = struct{}{}
	}

	return &subscription{
		eventC:    make(chan Event, defaultHandlerEventBuffer),
		exitC:     make(chan struct{}),
		logic:     logic,
		interests: interestsMap,
	}
}

type subscription struct {
	name      string
	eventC    chan Event
	exitC     chan struct{}
	logic     EventHandlerFunc
	interests map[EventType]struct{}
}

func (s *subscription) Start(waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)

	go func() {
		s.logic(s.eventC, s.exitC)
		waitGroup.Done()
	}()
}

func (s *subscription) Stop() {
	close(s.exitC)
}

func (s *subscription) Send(event Event) {
	select {
	case s.eventC <- event:
	default:
		log.Errorf("Failed to publish event %s to handler %s. Handler is processing events too slowly.", event.Type, s.name)
	}
}

func (s *subscription) Accepts(eventType EventType) bool {
	if _, acceptsAllEvents := s.interests[All]; !acceptsAllEvents {
		_, acceptsEvent := s.interests[eventType]
		return acceptsEvent
	}

	return true
}
