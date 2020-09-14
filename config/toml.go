package config

import (
	"fmt"

	"github.com/zinic/forculus/eventserver/event"
)

type RecordKeeperConfig struct {
	ExternalHostname string                         `toml:"external_hostname"`
	BindAddress      string                         `toml:"bind_address"`
	DatabasePath     string                         `toml:"db_path"`
	Users            map[string]AuthorizationConfig `toml:"user"`
	StorageProviders map[string]StorageProvider     `toml:"storage"`
}

type AuthorizationConfig struct {
	Password string `toml:"password"`
}

type eventServerConfiguration struct {
	Zoneminder       Zoneminder                    `toml:"zoneminder"`
	StorageProviders map[string]StorageProvider    `toml:"storage"`
	RecordKeepers    map[string]RecordKeeperClient `toml:"record_keeper"`
	Uploaders        map[string]uploader           `toml:"uploader"`
	SMTPServers      map[string]SMTPServer         `toml:"smtp_server"`
	EmailAlerts      map[string]emailAlert         `toml:"email_alert"`
	Emailers         map[string]Emailer            `toml:"emailer"`
}

type Zoneminder struct {
	Scheme   string `toml:"scheme"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	RootPath string `toml:"root_path"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type RecordKeeperClient struct {
	Scheme     string   `toml:"scheme"`
	Host       string   `toml:"host"`
	Port       int      `toml:"port"`
	Username   string   `toml:"username"`
	Password   string   `toml:"password"`
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

type StorageProvider struct {
	Provider   StorageProviderType `toml:"provider"`
	Properties map[string]string   `toml:"properties"`
}

type Emailer struct {
	Server     string   `toml:"server"`
	Recipients []string `toml:"recipients"`
}

type uploader struct {
	StorageTarget string      `toml:"storage_target"`
	Filter        alertFilter `toml:"filter"`
}

type emailAlert struct {
	Server          string      `toml:"server"`
	SubjectTemplate string      `toml:"subject_template"`
	Recipients      []string    `toml:"recipients"`
	Filter          alertFilter `toml:"filter"`
}

type alertFilter struct {
	EventTrigger        event.Type `toml:"event_trigger"`
	NameFilterRegex     string     `toml:"name_filter"`
	AlertFrameThreshold int        `toml:"alert_frame_threshold"`
	EventTimeAfter      string     `toml:"event_time_after"`
	EventTimeBefore     string     `toml:"event_time_before"`
}
