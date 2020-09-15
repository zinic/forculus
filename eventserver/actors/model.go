package actors

import "github.com/zinic/forculus/zoneminder/zmapi"

type MonitorEventUploadedPayload struct {
	Source        zmapi.MonitorEvent
	StorageTarget string
	StorageKey    string
}

type MonitorEventRecordedPayload struct {
	Source    zmapi.MonitorEvent
	AccessURL string
}
