package config

import "github.com/zinic/forculus/eventserver/event"

type configuration struct {
	Zoneminder       Zoneminder                 `toml:"zoneminder"`
	StorageProviders map[string]StorageProvider `toml:"storage"`
	Uploaders        map[string]uploader        `toml:"uploader"`
	SMTPServers      map[string]SMTPServer      `toml:"smtp_server"`
	EmailAlerts      map[string]emailAlert      `toml:"email_alert"`
}

type StorageProvider struct {
	Provider   StorageProviderType `toml:"provider"`
	Properties map[string]string   `toml:"properties"`
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
