package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-errors/errors"
	"github.com/jeffyfung/flight-info-agg/config"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/jeffyfung/flight-info-agg/pkg/collection"
	"github.com/jeffyfung/flight-info-agg/pkg/database/mongoDB"
	"github.com/jeffyfung/flight-info-agg/pkg/notification/telegram"
	"github.com/jeffyfung/flight-info-agg/pkg/scrapper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

func main() {
	config.LoadConfig()

	err := mongoDB.InitDB()
	if err != nil {
		log.Fatal("MongDB error: ", err.(*errors.Error).ErrorStack())
	}
	defer func() {
		err = mongoDB.Disconnect()
		if err != nil {
			log.Fatal("MongDB error: ", err.(*errors.Error).ErrorStack())
		}
	}()

	posts, err := scrapper.Scrap()
	if err != nil {
		log.Fatal("Cron job fails", err.(*errors.Error).ErrorStack())
	}

	result, err := deleteOldPosts(3)
	if err != nil {
		log.Fatal("Cannot delete old posts: ", err.(*errors.Error).ErrorStack())
	}
	fmt.Printf("Delete %d old posts >3 months old\n", result.DeletedCount)

	notify(posts)

}

func deleteOldPosts(expireMonths int) (result *mongo.DeleteResult, err error) {
	filter := bson.D{{
		Key:   "pub_date",
		Value: bson.M{"$lt": time.Now().UTC().AddDate(0, -1*expireMonths, 0)},
	}}
	result, err = mongoDB.DeleteMany("posts", filter)
	if err != nil {
		return nil, errors.New(err)
	}
	return result, nil
}

func notify(posts []model.Post) error {

	notifier := telegram.NewNotifier()

	users, err := mongoDB.Find[model.User]("users", bson.D{})
	if err != nil {
		return errors.New("Cannot get users" + err.(*errors.Error).ErrorStack())
	}

	g := new(errgroup.Group)

	for _, user := range users {
		if user.Notification != model.NotificationOn {
			continue
		}

		matchedPosts := newMatchedPost(posts, user)
		if len(matchedPosts) == 0 {
			continue
		}

		g.Go(func(u model.User) func() error {
			return func() error {
				content := notifier.FormatAlertMessages(u, matchedPosts)
				return notifier.Notify(u, content)
			}
		}(user))
	}

	if err := g.Wait(); err != nil {
		log.Println("Cannot send email: " + err.(*errors.Error).ErrorStack())
	}

	return nil

}

func newMatchedPost(posts []model.Post, user model.User) []model.Post {
	matched := []model.Post{}
	for _, post := range posts {
		locationsMatched := len(user.SelectedLocations) == 0 || collection.HaveOverlap[string](post.Locations, user.SelectedLocations)
		airlinesMatched := len(user.SelectedAirlines) == 0 || collection.HaveOverlap[string](post.Airlines, user.SelectedAirlines)
		if locationsMatched && airlinesMatched {
			matched = append(matched, post)
		}
	}
	return matched
}
