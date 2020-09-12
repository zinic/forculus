package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zinic/forculus/eventserver/event"
)

type StorageProviderType string

const (
	ProviderAWS StorageProviderType = "aws_s3"
)

type EventServerConfig struct {
	Zoneminder       Zoneminder
	StorageProviders map[string]StorageProvider
	Uploaders        map[string]Uploader
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

func (s EmailAlert) FormatRecipients() string {
	return strings.Join(s.Recipients, ",")
}

type Zoneminder struct {
	Scheme   string `toml:"scheme"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	RootPath string `toml:"root_path"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type SMTPServer struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	UseTLS   bool   `toml:"use_tls"`
	Sender   string `toml:"sender"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

func (s SMTPServer) FormatAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type AlertFilter struct {
	EventTrigger        event.Type
	NameRegex           *regexp.Regexp
	AlertFrameThreshold int
	EventTimeAfter      time.Time
	EventTimeBefore     time.Time
}
