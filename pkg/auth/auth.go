package auth

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jeffyfung/flight-info-agg/config"
	"github.com/jeffyfung/flight-info-agg/pkg/encoding"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

const (
	MaxAge  = 60 * 60 * 24 // * 30
	session = "session"    // key for gorilla store
)

var Store *sessions.CookieStore

func NewAuth() {
	config := config.Cfg

	Store = sessions.NewCookieStore([]byte(config.Server.Secret))

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   MaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	gothic.Store = Store

	var domain string
	if config.Server.Domain != "" {
		domain = config.Server.Domain
	} else {
		domain = "http://localhost:" + config.Server.Port
	}

	goth.UseProviders(
		google.New(config.Server.GoogleClientID, config.Server.GoogleClientSecret, fmt.Sprintf("%v/auth/callback?provider=google", domain)),
		github.New(config.Server.GithubClientID, config.Server.GithubClientSecret, fmt.Sprintf("%v/auth/callback?provider=github", domain), "user:email"),
	)
}

func AddUserToSession(c echo.Context, user goth.User) error {
	session, err := Store.Get(c.Request(), session)
	if err != nil {
		return fmt.Errorf("problem getting session from store: %v", err.Error())
	}

	// Remove the raw data to reduce the size
	user.RawData = map[string]interface{}{}

	session.Values["user"] = user
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("problem saving session data: %v", err.Error())
	}

	userInfo := struct {
		Email     string `json:"email"`
		Provider  string `json:"provider"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}{
		Email:     user.Email,
		Provider:  user.Provider,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
	}

	userInfoStr, err := encoding.StructToBase64(userInfo)
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     "user_info",
		Value:    userInfoStr,
		Path:     "/",
		MaxAge:   MaxAge,
		HttpOnly: false,
		Secure:   true,
	})

	return nil
}

func RemoveUserFromSession(c echo.Context) error {
	session, err := Store.Get(c.Request(), session)
	if err != nil {
		return fmt.Errorf("problem getting session from store: %v", err.Error())
	}

	session.Options.MaxAge = -1
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("problem saving session data: %v", err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name:     "user_info",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   true,
	})

	return nil
}

func User(c echo.Context) (*goth.User, error) {
	session, _ := Store.Get(c.Request(), session)
	user, ok := session.Values["user"]
	if !ok {
		return nil, fmt.Errorf("User not found in session")
	}
	u := user.(goth.User)

	return &u, nil
}
