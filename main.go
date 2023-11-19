package main

import (
	"log"
	"time"

	"github.com/jeffyfung/flight-info-agg/modules/scrapper"
	"github.com/jeffyfung/flight-info-agg/utils/database/mongoDB"
	"github.com/joho/godotenv"
)

func main() {
	// TODO: do not load env if in prod env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Cannot load .env")
	}

	err = mongoDB.InitDB()
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

	time.Sleep(time.Minute * 5)

	// scrap info from 2 website
	// convert them into a standard format
	// insert it into a database
	// create an endpoint for it
}

// TODO: watch Ben Davis's video
// TODO: scrap flyagain
// TODO: logrus - https://github.com/Sirupsen/logrus
// TODO: DB housekeeping - remove data from e.g. 3 months ago
