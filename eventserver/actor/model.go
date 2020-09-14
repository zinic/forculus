package actor

import (
	"github.com/zinic/forculus/eventserver/event"
	"github.com/zinic/forculus/eventserver/service"
)

type Subscriber interface {
	Handle(e event.Event)
	service.Service
}

type Dispatch interface {
	Dispatch(event event.Event)
}

type Reactor interface {
	Register(subscriber Subscriber, eventInterests ...event.Type)
	Dispatch
}
