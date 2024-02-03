package telegram

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
	"github.com/go-resty/resty/v2"
	"github.com/jeffyfung/flight-info-agg/config"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/jeffyfung/flight-info-agg/pkg/collection"
	"github.com/jeffyfung/flight-info-agg/pkg/notification"
)

type (
	User struct {
		ID           int    `json:"id"`
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		LanguageCode string `json:"language_code"`
	}

	Chat struct {
		ID int64 `json:"id"`
	}

	Message struct {
		MessageID       int    `json:"message_id"`
		MessageThreadID int    `json:"message_thread_id"`
		From            User   `json:"from"`
		Date            int    `json:"date"`
		Text            string `json:"text"`
		Chat            Chat   `json:"chat"`
		// TODO: more fields
	}

	WebhookRequest struct {
		UpdateID int     `json:"update_id"`
		Message  Message `json:"message"`
	}

	TelegramNotifier struct{}

	SetUpApiResponse struct {
		OK          bool   `json:"ok"`
		Result      bool   `json:"result"`
		Description string `json:"description"`
	}
)

func NewNotifier() notification.Notifier {
	return &TelegramNotifier{}
}

func (t *TelegramNotifier) SetUp() error {
	// set server url depending on whether it's prod
	var callbackURL string
	if !config.Cfg.Prod {
		callbackURL = config.Cfg.NgrokURL + "/webhook/telegram"
	} else {
		callbackURL = config.Cfg.Server.Domain + "/webhook/telegram"
	}

	// set webhook
	resp, err := resty.New().R().
		SetQueryParams(map[string]string{
			"url": callbackURL,
		}).
		SetResult(&SetUpApiResponse{}).
		Get(fmt.Sprintf("https://api.telegram.org/bot%v/setWebhook", config.Cfg.Telegram.BotToken))
	if err != nil {
		return errors.New("Cannot set Telegram webhook" + err.Error())
	}

	respBody := resp.Result().(*SetUpApiResponse)
	if !respBody.OK {
		return errors.New("Cannot set Telegram webhook" + respBody.Description)
	}
	return nil
}

func (t *TelegramNotifier) Notify(user model.User, text string) error {
	if user.TelegramChatID == 0 {
		return errors.New("Cannot send Telegram message: no Telegram chat ID found")
	}
	return t.NotifyChat(user.TelegramChatID, text)
}

func (t *TelegramNotifier) NotifyChat(chatID int64, text string) error {
	_, err := resty.New().R().
		SetBody(map[string]any{
			"chat_id": chatID,
			"text":    text,
		}).
		Post(fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", config.Cfg.Telegram.BotToken))

	if err != nil {
		return errors.New("Cannot send Telegram message" + err.Error())
	}
	return nil
}

func (t *TelegramNotifier) FormatAlertMessages(user model.User, posts []model.Post) string {
	p := collection.Map(posts, func(post model.Post) string {
		return post.Title + "\n" + post.URL
	})

	var selectedLocs string
	if len(user.SelectedLocations) == 0 {
		selectedLocs = "All"
	} else {
		selectedLocs = strings.Join(user.SelectedLocations, ", ")
	}

	var selectedAirlines string
	if len(user.SelectedAirlines) == 0 {
		selectedAirlines = "All"
	} else {
		selectedAirlines = strings.Join(user.SelectedAirlines, ", ")
	}

	formatted := fmt.Sprintf(`
New posts based on your search criteria:
Destinations: %v
Airlines: %v

Posts:

%v
`, selectedLocs, selectedAirlines, strings.Join(p, "\n\n"))

	return formatted
}
