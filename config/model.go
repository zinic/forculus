package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zinic/forculus/event"
)

type Configuration struct {
	Zoneminder  Zoneminder
	S3Uploaders map[string]S3Uploader
	SMTPServer  map[string]SMTPServer
	EmailAlerts map[string]EmailAlert
}

type S3Uploader struct {
	Filter AlertFilter
	s3Uploader
}

func (s S3Uploader) Enabled() bool {
	return s.AccessKeyID != "" && s.SecretAccessKey != "" && s.Region != "" && s.Bucket != ""
}

type EmailAlert struct {
	Filter AlertFilter
	emailAlert
}

func (s EmailAlert) FormatRecipients() string {
	return strings.Join(s.Recipients, ",")
}

type configuration struct {
	Zoneminder  Zoneminder            `toml:"zoneminder"`
	S3Uploaders map[string]s3Uploader `toml:"s3_uploads"`
	SMTPServers map[string]SMTPServer `toml:"smtp_servers"`
	EmailAlerts map[string]emailAlert `toml:"email_alerts"`
}

type Zoneminder struct {
	Scheme   string `toml:"scheme"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	RootPath string `toml:"root_path"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type s3Uploader struct {
	AccessKeyID     string      `toml:"access_key_id"`
	SecretAccessKey string      `toml:"secret_access_key"`
	Bucket          string      `toml:"bucket"`
	Region          string      `toml:"region"`
	Filter          alertFilter `toml:"filter"`
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

type emailAlert struct {
	Name            string      `toml:"name"`
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

type AlertFilter struct {
	EventTrigger        event.Type
	NameRegex           *regexp.Regexp
	AlertFrameThreshold int
	EventTimeAfter      time.Time
	EventTimeBefore     time.Time
}
