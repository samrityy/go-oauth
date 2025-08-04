package models

import "time"

type UserOAuth struct {
	ID           string
	Provider     string
	ProviderID   string
	AccessToken  string
	RefreshToken string
	TokenExpiry  time.Time
	UserID       int
}
