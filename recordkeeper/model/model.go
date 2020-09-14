package model

type EventRecord struct {
	ID            uint64            `json:"id,omitempty"`
	StorageTarget string            `json:"storage_target"`
	StorageKey    string            `json:"storage_key"`
	AccessToken   string            `json:"access_token"`
	Tags          map[string]string `json:"tags"`
}
