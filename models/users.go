package model

import (
	"time"
)

type (
	Query struct {
		SelectedLocations []string     `json:"selected_locations,omitempty" bson:"selected_locations,omitempty"`
		SelectedAirlines  []string     `json:"selected_airlines,omitempty" bson:"selected_airlines,omitempty"`
		Notification      Notification `json:"notification" bson:"notification"`
	}

	User struct {
		Query       `bson:"inline"`
		ID          string     `json:"id" bson:"_id"`
		Email       string     `json:"email" bson:"email"`
		Name        string     `json:"name" bson:"name"`
		Provider    string     `json:"provider" bson:"provider"`
		AvatarURL   string     `json:"avatar_url" bson:"avatar_url"`
		LastUpdated *time.Time `json:"last_updated" bson:"last_updated"`
		LastLogin   *time.Time `json:"last_login" bson:"last_login"`
	}

	Role int

	Notification int
)

const (
	RoleBasic Role = iota
	RoleAdmin
	// NotificationInstant does not exist because data is scrapped daily
)

const (
	NotificationNull Notification = iota
	NotificationDaily
)
