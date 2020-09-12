package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zinic/forculus/zoneminder/constants"
)

type AlertedMonitor struct {
	Monitor     Monitor
	AlarmStatus AlarmStatus
}

type LoginCredentials struct {
	Username string
	Password string
}

type LoginDetails struct {
	AccessToken         string  `json:"access_token"`
	AccessTokenExpires  float64 `json:"access_token_expires"`
	APIVersion          string  `json:"apiversion"`
	AppendPassword      int64   `json:"append_password"`
	Credentials         string  `json:"credentials"`
	RefreshToken        string  `json:"refresh_token"`
	RefreshTokenExpires float64 `json:"refresh_token_expires"`
	Version             string  `json:"version"`
}

type Version struct {
	APIVersion     string `json:"apiversion"`
	ServiceVersion string `json:"version"`
}

type EventList []MonitorEvent

type ListEventsResponse struct {
	Events     []EventWrapper    `json:"events"`
	Pagination PaginationDetails `json:"pagination"`
}

type MonitorEvent struct {
	AlarmFrames        string `json:"AlarmFrames"`
	Archived           string `json:"Archived"`
	AvgScore           string `json:"AvgScore"`
	Cause              string `json:"Cause"`
	DefaultVideo       string `json:"DefaultVideo"`
	DiskSpace          string `json:"DiskSpace"`
	Emailed            string `json:"Emailed"`
	EndTime            string `json:"EndTime"`
	Executed           string `json:"Executed"`
	FileSystemPath     string `json:"FileSystemPath"`
	Frames             string `json:"Frames"`
	Height             string `json:"Height"`
	ID                 string `json:"Id"`
	Length             string `json:"Length"`
	Locked             bool   `json:"Locked"`
	MaxScore           string `json:"MaxScore"`
	MaxScoreFrameID    string `json:"MaxScoreFrameId"`
	Messaged           string `json:"Messaged"`
	MonitorID          string `json:"MonitorId"`
	Name               string `json:"Name"`
	Notes              string `json:"Notes"`
	Orientation        string `json:"Orientation"`
	SaveJPEGs          string `json:"SaveJPEGs"`
	Scheme             string `json:"Scheme"`
	SecondaryStorageID string `json:"SecondaryStorageId"`
	StartTime          string `json:"StartTime"`
	StateID            string `json:"StateId"`
	StorageID          string `json:"StorageId"`
	TotScore           string `json:"TotScore"`
	Uploaded           string `json:"Uploaded"`
	Videoed            string `json:"Videoed"`
	Width              string `json:"Width"`
}

func (s MonitorEvent) ParseAlertFrames() (int, error) {
	return strconv.Atoi(s.AlarmFrames)
}

func (s MonitorEvent) String() string {
	return fmt.Sprintf("%s:%s", s.Name, s.ID)
}

func (s MonitorEvent) ParseStartTime() (time.Time, error) {
	return time.Parse(constants.ZMDateFormat, s.StartTime)
}

func (s MonitorEvent) ParseEndTime() (time.Time, error) {
	return time.Parse(constants.ZMDateFormat, s.EndTime)
}

type PaginationDetails struct {
	Count      int64                  `json:"count"`
	Current    int64                  `json:"current"`
	Limit      int64                  `json:"limit"`
	NextPage   bool                   `json:"nextPage"`
	Options    map[string]interface{} `json:"options"`
	Order      map[string]string      `json:"order"`
	Page       int64                  `json:"page"`
	PageCount  int64                  `json:"pageCount"`
	ParamType  string                 `json:"paramType"`
	PrevPage   bool                   `json:"prevPage"`
	QueryScope interface{}            `json:"queryScope"`
}

type EventWrapper struct {
	Event MonitorEvent `json:"Event"`
}

type MonitorList []Monitor

type ListMonitorsResponse struct {
	Monitors []Monitor `json:"monitors"`
}

type MonitorDetails struct {
	AlarmFrameCount        string      `json:"AlarmFrameCount"`
	AlarmMaxFPS            interface{} `json:"AlarmMaxFPS"`
	AlarmRefBlendPerc      string      `json:"AlarmRefBlendPerc"`
	AnalysisFPSLimit       interface{} `json:"AnalysisFPSLimit"`
	AnalysisUpdateDelay    string      `json:"AnalysisUpdateDelay"`
	ArchivedEventDiskSpace interface{} `json:"ArchivedEventDiskSpace"`
	ArchivedEvents         interface{} `json:"ArchivedEvents"`
	AutoStopTimeout        interface{} `json:"AutoStopTimeout"`
	Brightness             string      `json:"Brightness"`
	Channel                string      `json:"Channel"`
	Colour                 string      `json:"Colour"`
	Colours                string      `json:"Colours"`
	Contrast               string      `json:"Contrast"`
	ControlAddress         interface{} `json:"ControlAddress"`
	ControlDevice          interface{} `json:"ControlDevice"`
	ControlID              interface{} `json:"ControlId"`
	Controllable           string      `json:"Controllable"`
	DayEventDiskSpace      string      `json:"DayEventDiskSpace"`
	DayEvents              string      `json:"DayEvents"`
	DecoderHWAccelDevice   interface{} `json:"DecoderHWAccelDevice"`
	DecoderHWAccelName     string      `json:"DecoderHWAccelName"`
	DefaultCodec           string      `json:"DefaultCodec"`
	DefaultRate            string      `json:"DefaultRate"`
	DefaultScale           string      `json:"DefaultScale"`
	Deinterlacing          string      `json:"Deinterlacing"`
	Device                 string      `json:"Device"`
	Enabled                string      `json:"Enabled"`
	EncoderParameters      string      `json:"EncoderParameters"`
	EventPrefix            string      `json:"EventPrefix"`
	Exif                   bool        `json:"Exif"`
	FPSReportInterval      string      `json:"FPSReportInterval"`
	Format                 string      `json:"Format"`
	FrameSkip              string      `json:"FrameSkip"`
	Function               string      `json:"Function"`
	Height                 string      `json:"Height"`
	Host                   interface{} `json:"Host"`
	HourEventDiskSpace     string      `json:"HourEventDiskSpace"`
	HourEvents             string      `json:"HourEvents"`
	Hue                    string      `json:"Hue"`
	ID                     string      `json:"Id"`
	ImageBufferCount       string      `json:"ImageBufferCount"`
	LabelFormat            string      `json:"LabelFormat"`
	LabelSize              string      `json:"LabelSize"`
	LabelX                 string      `json:"LabelX"`
	LabelY                 string      `json:"LabelY"`
	LinkedMonitors         string      `json:"LinkedMonitors"`
	MaxFPS                 interface{} `json:"MaxFPS"`
	Method                 string      `json:"Method"`
	MinSectionLength       string      `json:"MinSectionLength"`
	MonthEventDiskSpace    string      `json:"MonthEventDiskSpace"`
	MonthEvents            string      `json:"MonthEvents"`
	MotionFrameSkip        string      `json:"MotionFrameSkip"`
	Name                   string      `json:"Name"`
	Notes                  string      `json:"Notes"`
	Options                interface{} `json:"Options"`
	Orientation            string      `json:"Orientation"`
	OutputCodec            interface{} `json:"OutputCodec"`
	OutputContainer        interface{} `json:"OutputContainer"`
	Palette                string      `json:"Palette"`
	Pass                   interface{} `json:"Pass"`
	Path                   string      `json:"Path"`
	Port                   string      `json:"Port"`
	PostEventCount         string      `json:"PostEventCount"`
	PreEventCount          string      `json:"PreEventCount"`
	Protocol               interface{} `json:"Protocol"`
	RTSPDescribe           bool        `json:"RTSPDescribe"`
	RecordAudio            string      `json:"RecordAudio"`
	RefBlendPerc           string      `json:"RefBlendPerc"`
	Refresh                interface{} `json:"Refresh"`
	ReturnDelay            interface{} `json:"ReturnDelay"`
	ReturnLocation         string      `json:"ReturnLocation"`
	SaveJPEGs              string      `json:"SaveJPEGs"`
	SectionLength          string      `json:"SectionLength"`
	Sequence               string      `json:"Sequence"`
	ServerID               string      `json:"ServerId"`
	SignalCheckColour      string      `json:"SignalCheckColour"`
	SignalCheckPoints      string      `json:"SignalCheckPoints"`
	StorageID              string      `json:"StorageId"`
	StreamReplayBuffer     string      `json:"StreamReplayBuffer"`
	SubPath                string      `json:"SubPath"`
	TotalEventDiskSpace    string      `json:"TotalEventDiskSpace"`
	TotalEvents            string      `json:"TotalEvents"`
	TrackDelay             interface{} `json:"TrackDelay"`
	TrackMotion            string      `json:"TrackMotion"`
	Triggers               string      `json:"Triggers"`
	Type                   string      `json:"Type"`
	User                   interface{} `json:"User"`
	V4LCapturesPerFrame    string      `json:"V4LCapturesPerFrame"`
	V4LMultiBuffer         interface{} `json:"V4LMultiBuffer"`
	VideoWriter            string      `json:"VideoWriter"`
	WarmupCount            string      `json:"WarmupCount"`
	WebColour              string      `json:"WebColour"`
	WeekEventDiskSpace     string      `json:"WeekEventDiskSpace"`
	WeekEvents             string      `json:"WeekEvents"`
	Width                  string      `json:"Width"`
	ZoneCount              string      `json:"ZoneCount"`
}

func (s MonitorDetails) ParseAlertFrameCount() (int, error) {
	return strconv.Atoi(s.AlarmFrameCount)
}

type ExportEventResult struct {
	Result     string `json:"result"`
	ExportFile string `json:"exportFile"`
}

type MonitorStatus struct {
	AnalysisFPS      string `json:"AnalysisFPS"`
	CaptureBandwidth string `json:"CaptureBandwidth"`
	CaptureFPS       string `json:"CaptureFPS"`
	MonitorID        string `json:"MonitorId"`
	State            string `json:"Status"`
}

type Monitor struct {
	Details MonitorDetails `json:"Monitor"`
	Status  MonitorStatus  `json:"Monitor_Status"`
}

func (s Monitor) Name() string {
	return fmt.Sprintf("%s(id:%s)", s.Details.Name, s.Details.ID)
}

type MonitorAlarmStatus struct {
	Status string `json:"status"`
}

type AlarmStatus string

func (s AlarmStatus) String() string {
	switch s {
	case AlarmStatusIdle:
		return "Idle"

	case AlarmStatusPreAlarm:
		return "Pre-Alarm"

	case AlarmStatusAlert:
		return "Alerted"

	case AlarmStatusAlarm:
		return "Alarmed"

	case AlarmStatusTape:
		return "Alarm Taping"

	default:
		return "Invalid"
	}
}

const (
	AlarmStatusInvalid  AlarmStatus = ""
	AlarmStatusIdle     AlarmStatus = "0"
	AlarmStatusPreAlarm AlarmStatus = "1"
	AlarmStatusAlert    AlarmStatus = "2"
	AlarmStatusAlarm    AlarmStatus = "3"
	AlarmStatusTape     AlarmStatus = "4"
)

func ParseAlarmStatus(raw string) AlarmStatus {
	switch raw {
	case string(AlarmStatusIdle):
		return AlarmStatusIdle

	case string(AlarmStatusPreAlarm):
		return AlarmStatusPreAlarm

	case string(AlarmStatusAlert):
		return AlarmStatusAlert

	case string(AlarmStatusAlarm):
		return AlarmStatusAlarm

	case string(AlarmStatusTape):
		return AlarmStatusTape

	default:
		return AlarmStatusInvalid
	}
}
