package scrapper

import (
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-co-op/gocron"
	colly "github.com/gocolly/colly/v2"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/jeffyfung/flight-info-agg/pkg/database/mongoDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	flydayURL   string = "https://flyday.hk/category/%e6%a9%9f%e7%a5%a8%e5%84%aa%e6%83%a0-tickets-promotions/"
	flyAgainURL string = "https://flyagain.la/"
)

func RunScrapper(interval time.Duration) error {
	go main()

	sche := gocron.NewScheduler(time.UTC)
	_, err := sche.Every(1).Day().At("23:00").Do(main)
	if err != nil {
		return err
	}

	sche.StartAsync()
	return nil
}

func main() {
	scrap()
	notify()
}

func scrap() error {
	log.Println("Start scrapping")
	c := colly.NewCollector(
		colly.AllowedDomains("flyday.hk", "flyagain.la"),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
		log.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println("Response code", r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error", err.Error())
	})

	posts := []model.Post{}
	lastScrapDate, err := getLastScrapDate()
	if err != nil {
		log.Println("Cannot get lastScrapDate", err.Error())
		return err
	}

	c.OnHTML("article.item", func(h *colly.HTMLElement) {
		div := h.DOM
		title := div.Find(".penci-entry-title > a").Text()
		URL := div.Find(".penci-entry-title > a").AttrOr("href", "")
		dateStr := div.Find("time.published").AttrOr("datetime", "")

		pubDate, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			log.Println("Cannot parse dateStr", pubDate)
			return
		}
		if pubDate.Before(lastScrapDate) {
			return
		}

		summary := div.Find(".item-content > p").Text()

		tags := []string{}
		div.Find(".cat > a").Each(func(_ int, s *goquery.Selection) {
			tag := s.Text()
			if strings.Contains(tag, "優惠") && tag != "機票優惠" {
				tags = append(tags, strings.ReplaceAll(tag, "優惠", ""))
			}
		})

		posts = append(posts, model.Post{
			Title:     title,
			Summary:   summary,
			Tags:      tags,
			URL:       URL,
			PubDate:   pubDate,
			CreatedAt: time.Now().UTC(),
		})
	})

	c.Visit(flydayURL)

	_, err = mongoDB.InsertBulkToCollection[model.Post]("posts", posts)
	if err != nil {
		log.Println("Cannot insert to posts table:", err)
		return err
	}
	log.Printf("Logged %v new posts\n", len(posts))

	err = updateLastScrapDate()
	if err != nil {
		log.Println("Cannot update system info:", err)
		return err
	}

	return nil
}

// TODO: notify users on matched query (after last notification date - i.e. yesterday)
// email?

func notify() error {
	// TODO: make concurrent
	// for each user,
	// if user notification is on
	// get yesterday's data from DB
	// find the posts that match the user's query
	// send an email
	return nil
}

func updateLastScrapDate() error {
	update := bson.D{{
		Key:   "$set",
		Value: bson.D{{Key: "last_updated", Value: time.Now().UTC()}},
	}}
	opts := options.Update().SetUpsert(true)
	_, err := mongoDB.UpdateById("system", "scrapper", update, opts)
	return err
}

func getLastScrapDate() (time.Time, error) {
	type lu = struct {
		LastUpdated time.Time `bson:"last_updated"`
	}
	output, err := mongoDB.GetById[lu]("system", "scrapper")
	return output.LastUpdated, err
}
