package scrapper

import (
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-co-op/gocron"
	colly "github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/jeffyfung/flight-info-agg/utils/database/mongoDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var flydayURL string = "https://flyday.hk/category/%e6%a9%9f%e7%a5%a8%e5%84%aa%e6%83%a0-tickets-promotions/"

// var flyagainURL string = ""

func RunScrapper(interval time.Duration) error {
	sche := gocron.NewScheduler(time.UTC)
	_, err := sche.Every(1).Day().At("23:00").Do(Scrap)
	if err != nil {
		return err
	}

	sche.StartAsync()
	return nil
}

func Scrap() error {
	c := colly.NewCollector(
		colly.AllowedDomains("flyday.hk"),
		colly.Debugger(&debug.LogDebugger{}),
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
	lastScrapDate, _ := getLastScrapDate()

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

	_, err := mongoDB.InsertBulkToCollection[model.Post]("posts", posts)
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
	return mongoDB.GetById[time.Time]("system", "scrapper")
}
