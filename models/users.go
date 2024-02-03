package model

import (
	"time"
)

type (
	Query struct {
		SelectedLocations []string `json:"selected_locations,omitempty" bson:"selected_locations,omitempty"`
		SelectedAirlines  []string `json:"selected_airlines,omitempty" bson:"selected_airlines,omitempty"`
	}

	User struct {
		Query          `bson:"inline"`
		ID             string       `json:"id" bson:"_id"`
		Email          string       `json:"email" bson:"email"`
		Name           string       `json:"name" bson:"name"`
		Provider       string       `json:"provider" bson:"provider"`
		AvatarURL      string       `json:"avatar_url" bson:"avatar_url"`
		LastUpdated    *time.Time   `json:"last_updated" bson:"last_updated"`
		LastLogin      *time.Time   `json:"last_login" bson:"last_login"`
		Notification   Notification `json:"notification" bson:"notification"`
		TelegramUID    string       `json:"telegram_uid" bson:"telegram_uid"`
		TelegramChatID int64        `json:"telegram_chat_id" bson:"telegram_chat_id,omitempty"`
	}

	Role int

	Notification int
)

const (
	RoleBasic Role = iota
	RoleAdmin
)

const (
	NotificationOff Notification = iota
	NotificationOn
)
