package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/jeffyfung/flight-info-agg/pkg/auth"
	"github.com/jeffyfung/flight-info-agg/pkg/database/mongoDB"
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

	auth.AddUserToSession(c, gothUser)

	// when user logs in, if the user is not in the database, create a new user with the information from provider
	// e.g. email, name, userID? - uid -> provider + userID
	// the callback should return whether the user is new or not
	dbUser, err := mongoDB.GetById[model.User]("users", gothUser.Provider+"__"+gothUser.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			t := time.Now().UTC()
			user := model.User{
				ID:          gothUser.Provider + "__" + gothUser.Email,
				Email:       gothUser.Email,
				Name:        gothUser.Name,
				Provider:    gothUser.Provider,
				AvatarURL:   gothUser.AvatarURL,
				LastUpdated: &t,
				Query: model.Query{
					Notification: model.NotificationNull,
				},
			}
			mongoDB.InsertToCollection[model.User]("users", user)
			return c.Redirect(http.StatusFound, "http://localhost:3000/profile?new=1")
		} else {
			fmt.Println(err.(*errors.Error).ErrorStack())
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	} else {
		// update last login time
		t := time.Now().UTC()
		mongoDB.UpdateById("users", dbUser.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "last_login", Value: &t}}}})
		return c.Redirect(http.StatusFound, "http://localhost:3000")
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
		model.User{
			ID:          user.ID,
			Email:       user.Email,
			Name:        user.Name,
			Provider:    user.Provider,
			AvatarURL:   user.AvatarURL,
			LastUpdated: user.LastUpdated,
			LastLogin:   user.LastLogin,
		},
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

	sort := []mongoDB.SortOption{{SortKey: "pub_date", Order: -1}}
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

	sort := []mongoDB.SortOption{{SortKey: "pub_date", Order: -1}}
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
