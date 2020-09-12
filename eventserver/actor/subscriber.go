package actor

import (
	"sync"

	"github.com/zinic/forculus/eventserver/event"
	"github.com/zinic/forculus/log"
)

const (
	defaultHandlerEventBuffer = 15
)

type SubscriberLogic func(eventC chan event.Event, exitC chan struct{})

func NewSubscriber(logic SubscriberLogic) Subscriber {
	return &subscriberInstance{
		eventC: make(chan event.Event, defaultHandlerEventBuffer),
		exitC:  make(chan struct{}),
		logic:  logic,
	}
}

type subscriberInstance struct {
	name   string
	eventC chan event.Event
	exitC  chan struct{}
	logic  SubscriberLogic
}

func (s *subscriberInstance) Handle(nextEvent event.Event) {
	select {
	case s.eventC <- nextEvent:
	default:
		log.Errorf("Failed to publish event %s to handler %s. Handler is processing events too slowly.", nextEvent.Type, s.name)
	}
}

func (s *subscriberInstance) Start(waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)

	go func() {
		s.logic(s.eventC, s.exitC)
		waitGroup.Done()
	}()
}

func (s *subscriberInstance) Stop() {
	close(s.exitC)
}
