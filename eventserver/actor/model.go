package actor

import (
	"sync"

	"github.com/zinic/forculus/eventserver/event"
)

type Service interface {
	Start(waitGroup *sync.WaitGroup)
	Stop()
}

type Subscriber interface {
	Handle(e event.Event)
	Service
}

type Dispatch interface {
	Dispatch(event event.Event)
}

type Reactor interface {
	Register(subscriber Subscriber, eventInterests ...event.Type)
	Service
	Dispatch
}
