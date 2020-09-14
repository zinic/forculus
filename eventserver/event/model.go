package event

type Type string

const (
	All                       Type = "all events"
	MonitorAlerted            Type = "monitor alerted"
	MonitorAlertStatusChanged Type = "monitor alert status changed"
	MonitorExitingAlert       Type = "monitor exiting alert"
	NewMonitorEvent           Type = "new monitor event"
	EventUploaded             Type = "new event uploaded"
	EmailNotice               Type = "email notification"
)

type Event struct {
	Type    Type
	Payload interface{}
}
