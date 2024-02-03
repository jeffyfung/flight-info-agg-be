package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jeffyfung/flight-info-agg/config"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/jeffyfung/flight-info-agg/pkg/auth"
	"github.com/jeffyfung/flight-info-agg/pkg/database/mongoDB"
	"github.com/jeffyfung/flight-info-agg/pkg/notification/telegram"
	"github.com/jeffyfung/flight-info-agg/pkg/tags"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	QueryPostRequest struct {
		From      time.Time `json:"from"`
		To        time.Time `json:"to"`
		Locations []string  `json:"locations"`
		Airlines  []string  `json:"airlines"`
	}

	UserQueryPostRequest struct {
		QueryPostRequest
		LoadUserSettings bool `json:"load_user_settings"`
	}
)

func HealthCheckHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func AuthProviderHandler(c echo.Context) error {
	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request()); err == nil {
		return c.JSON(http.StatusOK, gothUser)
	} else {
		gothic.BeginAuthHandler(c.Response(), c.Request())
		return nil
	}
}

func AuthCallbackHandler(c echo.Context) error {
	gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if gothUser.Name == "" && gothUser.NickName != "" {
		gothUser.Name = gothUser.NickName
	}

	auth.AddUserToSession(c, gothUser)

	// when user logs in, if the user is not in the database, create a new user with the information from provider
	// the callback should return whether the user is new or not
	dbUser, err := mongoDB.GetById[model.User]("users", gothUser.Provider+"__"+gothUser.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			t := time.Now().UTC()
			user := model.User{
				ID:           gothUser.Provider + "__" + gothUser.Email,
				Email:        gothUser.Email,
				Name:         gothUser.Name,
				Provider:     gothUser.Provider,
				AvatarURL:    gothUser.AvatarURL,
				LastUpdated:  &t,
				Notification: model.NotificationOff,
				TelegramUID:  uuid.New().String(),
			}
			mongoDB.InsertToCollection[model.User]("users", user)
			return c.Redirect(http.StatusFound, config.Cfg.UIOrigin+"/profile?new=1")
		} else {
			fmt.Println(err.(*errors.Error).ErrorStack())
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	} else {
		// update last login time
		t := time.Now().UTC()
		mongoDB.UpdateById("users", dbUser.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "last_login", Value: &t}}}})
		return c.Redirect(http.StatusFound, config.Cfg.UIOrigin)
	}

}

func ProviderLogoutHandler(c echo.Context) error {
	gothic.Logout(c.Response(), c.Request())
	auth.RemoveUserFromSession(c)
	return c.JSON(http.StatusOK, struct{}{})
}

func UserProfileHandler(c echo.Context) error {
	gothUser := c.Get("gothUser").(goth.User)
	userID := gothUser.Provider + "__" + gothUser.Email
	user, err := mongoDB.GetById[model.User]("users", userID)
	if err != nil {
		fmt.Println(err.(*errors.Error).ErrorStack())
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	selectedLocs := tags.EnrichLocationsWithLabels(user.SelectedLocations)
	selectedAirlines := tags.EnrichAirlinesWithLabels(user.SelectedAirlines)

	var wrappedUser = struct {
		model.User
		SelectedLocations []tags.DestWithLabel     `json:"selected_locations"`
		SelectedAirlines  []tags.AirlinesWithLabel `json:"selected_airlines"`
	}{
		user,
		selectedLocs,
		selectedAirlines,
	}

	return c.JSON(http.StatusOK, model.Response{Payload: wrappedUser})
}

func UpdateUserProfileHandler(c echo.Context) error {
	gothUser := c.Get("gothUser").(goth.User)
	userID := gothUser.Provider + "__" + gothUser.Email

	var req model.User
	err := c.Bind(&req)
	if err != nil {
		fmt.Printf("Unable to bind request: %v\n", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "last_updated", Value: time.Now().UTC()},
		{Key: "selected_locations", Value: req.SelectedLocations},
		{Key: "selected_airlines", Value: req.SelectedAirlines},
		{Key: "notification", Value: req.Notification},
	}}}
	_, err = mongoDB.UpdateById("users", userID, update)
	if err != nil {
		fmt.Println(err.(*errors.Error).ErrorStack())
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else {
		return c.JSON(http.StatusOK, struct{}{})
	}
}

func UserQueryPostsHandler(c echo.Context) error {
	var req UserQueryPostRequest
	err := c.Bind(&req)
	if err != nil {
		fmt.Printf("Unable to bind request: %v\n", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	filter := bson.D{}
	selectedLocations, selectedAirlines := req.Locations, req.Airlines

	if req.LoadUserSettings {
		gothUser := c.Get("gothUser").(goth.User)
		user, err := mongoDB.GetById[model.User]("users", gothUser.Provider+"__"+gothUser.Email)
		if err != nil {
			fmt.Println("Cannot find user in database")
			fmt.Println(err.(*errors.Error).ErrorStack())
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if len(user.SelectedLocations) > 0 {
			filter = append(filter, bson.E{
				Key:   "locations",
				Value: bson.M{"$in": user.SelectedLocations},
			})
		}
		if len(user.SelectedAirlines) > 0 {
			filter = append(filter, bson.E{
				Key:   "airlines",
				Value: bson.M{"$in": user.SelectedAirlines},
			})
		}
		selectedLocations, selectedAirlines = user.SelectedLocations, user.SelectedAirlines
	} else {
		if len(req.Locations) > 0 {
			filter = append(filter, bson.E{
				Key:   "locations",
				Value: bson.M{"$in": req.Locations},
			})
		}
		if len(req.Airlines) > 0 {
			filter = append(filter, bson.E{
				Key:   "airlines",
				Value: bson.M{"$in": req.Airlines},
			})
		}
	}

	sort := mongoDB.SortOption{SortKey: "pub_date", Order: -1}
	posts, err := mongoDB.Find[model.Post]("posts", filter, sort)
	if err != nil {
		fmt.Println("Cannot find posts in database")
		fmt.Println(err.(*errors.Error).ErrorStack())
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	selectedLocationsWithLabel := tags.EnrichLocationsWithLabels(selectedLocations)
	selectedAirlinesWithLabel := tags.EnrichAirlinesWithLabels(selectedAirlines)

	return c.JSON(http.StatusOK, model.Response{
		Payload: struct {
			Posts             []model.Post             `json:"posts"`
			SelectedLocations []tags.DestWithLabel     `json:"selected_locations"`
			SelectedAirlines  []tags.AirlinesWithLabel `json:"selected_airlines"`
		}{
			Posts:             posts,
			SelectedLocations: selectedLocationsWithLabel,
			SelectedAirlines:  selectedAirlinesWithLabel,
		},
	})
}

func QueryPostsHandler(c echo.Context) error {
	var req QueryPostRequest
	err := c.Bind(&req)
	if err != nil {
		fmt.Printf("Unable to bind request: %v\n", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	filter := bson.D{}
	if len(req.Locations) > 0 {
		filter = append(filter, bson.E{
			Key:   "locations",
			Value: bson.M{"$in": req.Locations},
		})
	}
	if len(req.Airlines) > 0 {
		filter = append(filter, bson.E{
			Key:   "airlines",
			Value: bson.M{"$in": req.Airlines},
		})
	}

	sort := mongoDB.SortOption{SortKey: "pub_date", Order: -1}
	posts, err := mongoDB.Find[model.Post]("posts", filter, sort)
	if err != nil {
		fmt.Println("Cannot find posts in database")
		fmt.Println(err.(*errors.Error).ErrorStack())
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, model.Response{
		Payload: struct {
			Posts []model.Post `json:"posts"`
		}{
			Posts: posts,
		},
	})
}

func TagsHandler(c echo.Context) error {
	dests := tags.DestinationsWithLabels()
	airlines := tags.AirlinesWithLabels()
	return c.JSON(http.StatusOK, model.Response{Payload: struct {
		Locations []tags.DestWithLabel     `json:"locations"`
		Airlines  []tags.AirlinesWithLabel `json:"airlines"`
	}{
		Locations: dests,
		Airlines:  airlines,
	}})
}

func TelegramWebhookHandler(c echo.Context) error {
	var req telegram.WebhookRequest
	err := c.Bind(&req)
	if err != nil {
		fmt.Printf("Unable to bind request: %v\n", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	chatID := req.Message.Chat.ID
	var message string

	inputText := req.Message.Text
	if strings.HasPrefix(inputText, "/start") {
		telegramUID := strings.TrimPrefix(inputText, "/start ")
		filter := bson.D{{Key: "telegram_uid", Value: telegramUID}}
		users, err := mongoDB.Find[model.User]("users", filter)
		if err != nil {
			fmt.Println(err.(*errors.Error).ErrorStack())
			message = "Internal server error. Please try again later. If this problem persists, please contact the team"
		} else if len(users) == 0 {
			fmt.Printf("Cannot find user with telegram UID: %v\n", telegramUID)
			message = "If you are trying to set up notifications, please sign in at our website (852-flight-deals.up.railway.app) and use the profile page to redirect to this bot"
		} else if users[0].TelegramChatID == 0 {
			updatedUser := users[0]
			updatedUser.TelegramChatID = req.Message.Chat.ID
			updatedUser.Notification = model.NotificationOn
			mongoDB.ReplaceByID[model.User]("users", updatedUser.ID, updatedUser)
			message = "Welcome to 852 Flight Deals! You have successfully set up notifications. You will now receive news about flight deals and discounts daily. Modify your notification settings (e.g. search filter) using the website: 852-flight-deals.up.railway.app"
		} else {
			message = "You have already set up notifications. Modify your notification settings (e.g. search filter) using the website: 852-flight-deals.up.railway.app"
		}
	} else {
		message = "This bot sends new posts relating to flight deals and discounts around Hong Kong. Modify your notification settings using the website: \n852-flight-deals.up.railway.app"
	}

	telegram.NewNotifier().NotifyChat(chatID, message)
	return c.JSON(http.StatusOK, struct{}{})
}
