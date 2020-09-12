package api

type RegisterEvent struct {
	SourceEventID   string `json:"source_event_id"`
	AccessToken     string `json:"access_token"`
	EventPath       string `json:"event_path"`
	StorageEndpoint string `json:"storage_endpoint"`
}
