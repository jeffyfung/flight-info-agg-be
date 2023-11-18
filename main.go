package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/jeffyfung/flight-info-agg/utils/database/mongoDB"
	"github.com/joho/godotenv"
)

type Post struct {
	Title   string   `bson:"title"`
	Summary string   `bson:"summary"`
	Tags    []string `bson:"tags"`
	URL     string   `bson:"url"`
	PubDate string   `bson:"pub_date"`
}

var FlydayURL string = "https://flyday.hk/category/%e6%a9%9f%e7%a5%a8%e5%84%aa%e6%83%a0-tickets-promotions/"
var FlyagainURL string = ""

func main() {
	godotenv.Load()

	err := mongoDB.InitDB()
	if err != nil {
		log.Fatal("MongDB error", err.Error())
	}

	defer func() {
		err = mongoDB.Disconnect()
		if err != nil {
			log.Fatal("MongDB error", err.Error())
		}
	}()

	err = scrap()
	if err != nil {
		log.Fatal("Scrapping error", err.Error())
	}
	// runScrapper(12 * time.Hour)
	// scrap info from 2 website
	// convert them into a standard format
	// insert it into a database
	// create an endpoint for it
}

// TODO: move to package "scrapper"
func runScrapper(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for ; ; <-ticker.C {
		scrap()
	}
}

func scrap() error {
	c := colly.NewCollector(colly.AllowedDomains("flyday.hk"))

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Response code", r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Error", err.Error())
	})

	posts := []Post{}
	c.OnHTML("article.item", func(h *colly.HTMLElement) {
		div := h.DOM
		title := div.Find(".penci-entry-title > a").Text()
		URL := div.Find(".penci-entry-title > a").AttrOr("href", "")
		pubDate := div.Find("time.published").AttrOr("datetime", "")

		summary := div.Find(".item-content > p").Text()

		tags := []string{}
		div.Find(".cat > a").Each(func(_ int, s *goquery.Selection) {
			tag := s.Text()
			if strings.Contains(tag, "優惠") && tag != "機票優惠" {
				tags = append(tags, strings.ReplaceAll(tag, "優惠", ""))
			}
		})

		// TODO: goroutine
		// go func() {

		// }()
		posts = append(posts, Post{
			Title:   title,
			Summary: summary,
			Tags:    tags,
			URL:     URL,
			PubDate: pubDate,
		})
	})

	c.Visit("https://flyday.hk/category/%e6%a9%9f%e7%a5%a8%e5%84%aa%e6%83%a0-tickets-promotions/")

	_, err := mongoDB.InsertBulkToCollection[Post]("posts", posts)
	if err != nil {
		log.Println("Cannot insert to posts table", err)
		return err
	}
	return nil
}

// TODO: scrap multiple pages
// TODO: DB housekeeping - remove data from e.g. 3 months ago
// TODO: what happen if you insert sth that already exists?
