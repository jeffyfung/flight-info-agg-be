package notification

import (
	model "github.com/jeffyfung/flight-info-agg/models"
)

type (
	Notifier interface {
		SetUp() error
		Notify(user model.User, text string) error
		NotifyChat(chatID int64, text string) error
		FormatAlertMessages(user model.User, posts []model.Post) string
	}
)
