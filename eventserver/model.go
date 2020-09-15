package eventserver

type EventType string

const (
	All                       EventType = "all events"
	MonitorAlerted            EventType = "monitor alerted"
	MonitorAlertStatusChanged EventType = "monitor alert status changed"
	MonitorExitingAlert       EventType = "monitor exiting alert"
	MonitorNewEvent           EventType = "new monitor event"
	MonitorEventUploaded      EventType = "new event uploaded"
	MonitorEventRecorded      EventType = "event saved in recordkeeper"
)

type Event struct {
	Type    EventType
	Payload interface{}
}

type EventHandlerFunc func(eventC <-chan Event, exitC chan struct{})

type EventDispatch interface {
	Send(event Event)
}

type SubscriptionManager interface {
	Register(handler EventHandlerFunc, eventInterests ...EventType)
	EventDispatch
}
