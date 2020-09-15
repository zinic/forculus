package eventserver

import (
	"sync"

	"github.com/zinic/forculus/service"
)

func NewDispatch(manager *service.Manager) SubscriptionManager {
	return &reactor{
		manager:      manager,
		dispatchLock: &sync.Mutex{},
	}
}

type reactor struct {
	manager       *service.Manager
	dispatchLock  *sync.Mutex
	subscriptions []*subscription
}

func (s *reactor) Stop() {
	s.manager.Stop()
}

func (s *reactor) Register(handler EventHandlerFunc, interests ...EventType) {
	newSubscription := NewSubscription(handler, interests...)

	s.subscriptions = append(s.subscriptions, newSubscription)
	s.manager.Start(newSubscription)
}

func (s *reactor) Send(event Event) {
	s.dispatchLock.Lock()
	defer s.dispatchLock.Unlock()

	for _, sub := range s.subscriptions {
		if sub.Accepts(event.Type) {
			sub.Send(event)
		}
	}
}
