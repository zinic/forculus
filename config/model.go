package config

import (
	"regexp"
	"time"

	"github.com/zinic/forculus/eventserver"
)

type StorageProviderType string

const (
	ProviderAWS StorageProviderType = "aws_s3"
)

type EventServerConfig struct {
	Zoneminder       Zoneminder
	StorageProviders map[string]StorageProvider
	Uploaders        map[string]Uploader
	RecordKeepers    map[string]RecordKeeperClient
	SMTPServers      map[string]SMTPServer
	EmailAlerts      map[string]EmailAlert
}

type Uploader struct {
	StorageTarget string
	Filter        AlertFilter
}

type EmailAlert struct {
	Filter AlertFilter
	emailAlert
}

type AlertFilter struct {
	EventTrigger        eventserver.EventType
	NameRegex           *regexp.Regexp
	AlertFrameThreshold int
	EventTimeAfter      time.Time
	EventTimeBefore     time.Time
}
