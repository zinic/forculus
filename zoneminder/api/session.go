package api

import "time"

type LoginSession struct {
	Details       LoginDetails
	LatestRefresh time.Time
}

func (s *LoginSession) Refresh(details LoginDetails) {
	s.LatestRefresh = time.Now()

	s.Details.Credentials = details.Credentials
	s.Details.AccessToken = details.AccessToken
	s.Details.AccessTokenExpires = details.AccessTokenExpires
}

func (s LoginSession) RefreshRequired() bool {
	return time.Now().Sub(s.LatestRefresh).Seconds() > s.Details.AccessTokenExpires
}

func (s LoginSession) CanRefresh() bool {
	return time.Now().Sub(s.LatestRefresh).Seconds() < s.Details.RefreshTokenExpires
}
