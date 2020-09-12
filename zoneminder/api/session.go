package api

import "time"

type LoginSession struct {
	Details     LoginDetails
	Created     time.Time
	LastRefresh time.Time
}

func (s *LoginSession) Refresh(details LoginDetails) {
	s.LastRefresh = time.Now()

	s.Details.Credentials = details.Credentials
	s.Details.AccessToken = details.AccessToken
	s.Details.AccessTokenExpires = details.AccessTokenExpires
}

func (s LoginSession) RefreshRequired() bool {
	return time.Now().Sub(s.LastRefresh).Seconds() > s.Details.AccessTokenExpires
}

func (s LoginSession) Expired() bool {
	return time.Now().Sub(s.Created).Seconds() > s.Details.RefreshTokenExpires
}
