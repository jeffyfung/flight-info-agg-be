package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jeffyfung/flight-info-agg/api/middlewares"
	"github.com/jeffyfung/flight-info-agg/config"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/jeffyfung/flight-info-agg/pkg/database/mongoDB"
	"github.com/jeffyfung/flight-info-agg/pkg/password"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	QueryPostPayload struct {
		From time.Time `json:"from" `
		To   time.Time `json:"to"`
		Tags []string  `json:"tags"`
	}

	SignInPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	UpdateQueryPayload struct {
		Tags         []string           `json:"tags,omitempty"`
		Notification model.Notification `json:"notification"`
	}
)

func HealthCheckHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// find posts by tags and date
// Question: when is this used? when website gets feed
func QueryPostsHandler(c echo.Context) error {
	var filter QueryPostPayload
	err := c.Bind(&filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	dbFilter := bson.M{}
	if filter.Tags != nil {
		dbFilter["tags"] = bson.M{"$in": filter.Tags}
	}
	if !filter.From.IsZero() || !filter.To.IsZero() {
		timeFilter := bson.M{}
		if !filter.From.IsZero() {
			timeFilter["$gte"] = filter.From
		}
		if !filter.To.IsZero() {
			timeFilter["$lt"] = filter.To
		}
		dbFilter["pub_date"] = timeFilter
	}

	posts, err := mongoDB.Find[model.Post]("posts", dbFilter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, model.Response{Payload: posts})
}

// Update the criterion of a user's feed
func UpdateQueryHandler(c echo.Context) error {
	var query UpdateQueryPayload
	err := c.Bind(&query)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	lastUpdated := time.Now().UTC()
	newQuery := model.Query{
		Tags:         query.Tags,
		Notification: query.Notification,
		LastUpdated:  &lastUpdated,
	}
	user := c.Get("user").(model.UserPublicInfo)
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "query", Value: newQuery}}}}
	_, err = mongoDB.UpdateById("users", user.Email, update)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "MongoDB: "+err.Error())
	}

	return c.JSON(http.StatusOK, struct{}{})
}

func CreateUserHandler(c echo.Context) error {
	var user model.User
	err := c.Bind(&user)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// check if email is taken
	u, err := mongoDB.GetById[model.User]("users", user.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// expected behaviour
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database Error: %v", err.Error()))
		}
	} else if u.Email != "" {
		// db already contains the email
		return echo.NewHTTPError(http.StatusBadRequest, "Email already taken")
	}

	user.Role = model.RoleBasic
	user.Password, err = password.HashPassword(user.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Password does not match required format or length")
	}
	t := time.Now().UTC()
	user.Query.LastUpdated = &t

	_, err = mongoDB.InsertToCollection[model.User]("users", user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database Error: %v", err.Error()))
	}

	return c.JSON(http.StatusCreated, model.Response{Payload: struct {
		Email string `json:"_id" bson:"_id"`
		Name  string `json:"name,omitempty"`
	}{
		Email: user.Email,
		Name:  user.Name,
	}})
}

func SignInHandler(c echo.Context) error {
	var payload SignInPayload
	err := c.Bind(&payload)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	user, err := mongoDB.GetById[model.User]("users", payload.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return echo.NewHTTPError(http.StatusUnauthorized, "User does not exist or password is invalid")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Database Error: %v", err.Error()))
	}
	if match := password.CheckPasswordHash(payload.Password, user.Password); !match {
		return echo.NewHTTPError(http.StatusUnauthorized, "User does not exist or password is invalid")
	}

	expirationTime := time.Now().Add(time.Duration(config.Cfg.Server.JwtExpiry) * time.Second)
	claims := &middlewares.Claims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(config.Cfg.Server.JwtSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot sign JWT: %v", err.Error()))
	}

	return c.JSON(http.StatusOK, model.Response{
		Payload: struct {
			JWT string `json:"jwt"`
		}{
			JWT: tokenStr,
		},
	})
}

// this is called after user logs in
// the first call should be at X - 30 seconds after log in
// i.e. 30 seconds buffer
func RenewJWTHandler(c echo.Context) error {
	authHeader, ok := c.Request().Header["Authorization"]
	if !ok || authHeader[0] == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "JWT not found in header")
	}

	claims := &middlewares.Claims{}
	jwtTokenStr := authHeader[0]
	_, err := jwt.ParseWithClaims(jwtTokenStr, claims, func(token *jwt.Token) (any, error) {
		return config.Cfg.Server.JwtSecret, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if time.Until(claims.ExpiresAt.Time) > 30*time.Second {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot refresh token until 30s before expiry")
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tknStr, err := token.SignedString(config.Cfg.Server.JwtSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, model.Response{
		Payload: struct {
			JWT string `json:"jwt"`
		}{
			JWT: tknStr,
		},
	})

}

// how to create an admin user?
// TODO: sign-in - use SSO?
