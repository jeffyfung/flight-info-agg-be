package main

import (
	"log"
	"time"

	"github.com/jeffyfung/flight-info-agg/api/handlers"
	"github.com/jeffyfung/flight-info-agg/api/middlewares"
	"github.com/jeffyfung/flight-info-agg/config"
	"github.com/jeffyfung/flight-info-agg/pkg/database/mongoDB"
	"github.com/jeffyfung/flight-info-agg/pkg/scrapper"
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

	err = scrapper.RunScrapper(12 * time.Hour)
	if err != nil {
		log.Fatal("Cron job fails", err.Error())
	}

	startServer()
}

func startServer() {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	e.GET("/", handlers.HealthCheckHandler)

	e.POST("/signin", handlers.SignInHandler)
	e.POST("/createUser", handlers.CreateUserHandler)

	// group of endpoints that require sign in
	user := e.Group("/user", middlewares.JWTMiddleware)
	user.POST("/posts", handlers.QueryPostsHandler)
	user.POST("/renewJWT", handlers.RenewJWTHandler)
	user.POST("/setFeed", handlers.UpdateQueryHandler)

	e.Logger.Fatal(e.Start(":" + config.Cfg.Server.Port))
}

// TODO: scrap flyagain
// TODO: DB housekeeping - remove data from e.g. 3 months ago
// TODO: single sign on
// TODO: check password strength - guard?
