package main

import (
	"log"

	"github.com/jeffyfung/flight-info-agg/api/handlers"
	"github.com/jeffyfung/flight-info-agg/api/middlewares"
	"github.com/jeffyfung/flight-info-agg/config"
	"github.com/jeffyfung/flight-info-agg/pkg/auth"
	"github.com/jeffyfung/flight-info-agg/pkg/database/mongoDB"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.LoadConfig()

	err := mongoDB.InitDB()
	if err != nil {
		log.Fatal("MongDB error: ", err.Error())
	}
	defer func() {
		err = mongoDB.Disconnect()
		if err != nil {
			log.Fatal("MongDB error: ", err.Error())
		}
	}()

	auth.NewAuth()
	startServer()
}

func startServer() {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:*"},
		AllowCredentials: true,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/", handlers.HealthCheckHandler)

	// login and logout
	e.GET("/auth/callback", handlers.AuthCallbackHandler)
	e.GET("/auth", handlers.AuthProviderHandler)
	e.GET("/logout", handlers.ProviderLogoutHandler)

	e.POST("/posts", handlers.QueryPostsHandler)
	e.GET("/tags", handlers.TagsHandler)

	// group of endpoints that require sign in
	user := e.Group("/user", middlewares.UserMiddleware)

	user.POST("/profile", handlers.UpdateUserProfileHandler)
	user.GET("/profile", handlers.UserProfileHandler)
	user.POST("/posts", handlers.UserQueryPostsHandler)

	e.Logger.Fatal(e.Start(":" + config.Cfg.Server.Port))
}
