package scrapper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-errors/errors"
	colly "github.com/gocolly/colly/v2"
	model "github.com/jeffyfung/flight-info-agg/models"
	"github.com/jeffyfung/flight-info-agg/pkg/collection"
	"github.com/jeffyfung/flight-info-agg/pkg/database/mongoDB"
	"github.com/jeffyfung/flight-info-agg/pkg/languages"
	"github.com/jeffyfung/flight-info-agg/pkg/tags"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	flydayURL   string = "https://flyday.hk/category/%e6%a9%9f%e7%a5%a8%e5%84%aa%e6%83%a0-tickets-promotions/"
	flyAgainURL string = "https://flyagain.la/"
)

type result struct {
	posts []model.Post
	error error
}

func Scrap() ([]model.Post, error) {

	lastScrapDate, err := getLastScrapDate()
	if err != nil {
		return nil, errors.New("Cannot get lastScrapDate" + err.Error())
	}

	ch := make(chan result)
	go scrapFlyday(ch, lastScrapDate)
	go scrapFlyAgain(ch, lastScrapDate)

	posts := []model.Post{}
	for i := 0; i < 2; i++ {
		scrappedPosts, ok := <-ch
		if !ok {
			return nil, errors.New("Cannot get posts from channel" + err.Error())
		}
		posts = append(posts, scrappedPosts.posts...)
	}

	if len(posts) > 0 {
		_, err = mongoDB.InsertBulkToCollection[model.Post]("posts", posts)
		if err != nil {
			return nil, errors.New("Cannot insert to posts table: " + err.Error())
		}
		fmt.Printf("Logged %v new posts\n", len(posts))
	}

	err = updateLastScrapDate()
	if err != nil {
		return nil, errors.New("Cannot update system info: " + err.Error())
	}

	return posts, nil
}

func scrapFlyday(ch chan result, lastScrapDate time.Time) {
	log.Println("Start scrapping flyday")
	c := colly.NewCollector(
		colly.AllowedDomains("flyday.hk"),
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

	c.OnHTML("article.item", func(h *colly.HTMLElement) {
		locations := []string{}
		airlines := []string{}

		div := h.DOM
		title := div.Find(".penci-entry-title > a").Text()
		locations = append(locations, extractDestinations(title)...)
		airlines = append(airlines, extractAirlines(title)...)

		URL := div.Find(".penci-entry-title > a").AttrOr("href", "https://flyday.hk/")
		dateStr := div.Find("time.published").AttrOr("datetime", "https://flyday.hk/")

		pubDate, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			log.Println("Cannot parse dateStr", pubDate)
			ch <- result{nil, err}
			return
		}
		if pubDate.Before(lastScrapDate) {
			return
		}

		summary := div.Find(".item-content > p").Text()
		airlines = append(airlines, extractAirlines(summary)...)

		div.Find(".cat > a").Each(func(_ int, s *goquery.Selection) {
			category := s.Text()
			locations = append(locations, extractDestinations(category)...)
		})

		locations = collection.RemoveListDuplicates[string](locations)
		airlines = collection.RemoveListDuplicates[string](airlines)

		posts = append(posts, model.Post{
			Title:     title,
			Summary:   summary,
			Locations: locations,
			Airlines:  airlines,
			URL:       URL,
			PubDate:   pubDate,
			CreatedAt: time.Now().UTC(),
			Source:    model.DataSourceFlyday,
		})
	})

	c.Visit(flydayURL)

	ch <- result{posts, nil}
}

func scrapFlyAgain(ch chan result, lastScrapDate time.Time) {
	log.Println("Start scrapping flyAgain")
	c := colly.NewCollector(
		colly.AllowedDomains("flyagain.la"),
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

	c.OnHTML("div.blogpostcategory", func(h *colly.HTMLElement) {
		locations := []string{}
		airlines := []string{}

		div := h.DOM
		title := div.Find("h2.title > a").Text()
		locations = append(locations, extractDestinations(title)...)
		airlines = append(airlines, extractAirlines(title)...)

		URL := div.AttrOr("this_url", "https://flyagain.la/")
		dateStr := strings.TrimSpace(div.Find("a.post-meta-time").Text())

		pubDate, err := time.Parse("January 2, 2006", dateStr)
		if err != nil {
			log.Println("Cannot parse dateStr", pubDate)
			ch <- result{nil, err}
			return
		}

		if pubDate.Before(lastScrapDate) {
			return
		}

		summary := ""
		div.Find("div.blogcontent > p").Each(func(_ int, s *goquery.Selection) {
			fieldName := s.Find("span").Text()
			if strings.Contains(fieldName, "航點") {
				locations = append(locations, extractDestinations(s.Text())...)
			} else if strings.Contains(fieldName, "航空公司") {
				airlines = append(airlines, extractAirlines(s.Text())...)
			}
			if strings.Contains(fieldName, "結論") {
				summary = strings.TrimSpace(s.Text())
				airlines = append(airlines, extractAirlines(summary)...)
			}
		})

		locations = append(locations, extractDestinations(div.Find("div.post-meta > div > a").Text())...)
		locations = collection.RemoveListDuplicates[string](locations)
		airlines = collection.RemoveListDuplicates[string](airlines)

		posts = append(posts, model.Post{
			Title:     title,
			Summary:   summary,
			Locations: locations,
			Airlines:  airlines,
			URL:       URL,
			PubDate:   pubDate,
			CreatedAt: time.Now().UTC(),
			Source:    model.DataSourceFlyAgain,
		})
	})

	c.Visit(flyAgainURL)

	ch <- result{posts, nil}
}

func extractDestinations(s string) []string {
	destMap := map[string]struct{}{}

	// directly check if the string contains the destination
	for _, destItem := range tags.Destinations {
		if strings.Contains(s, destItem[languages.TC]) {
			destMap[destItem[languages.TC]] = struct{}{}
		}
	}

	// check if the string contains the destination alias and convert to the destination
	for key := range tags.AliasToDestMap {
		if strings.Contains(s, key) {
			for _, destName := range tags.AliasToDestMap[key] {
				destMap[destName] = struct{}{}
			}
		}
	}

	output := make([]string, 0, len(destMap))
	for key := range destMap {
		output = append(output, key)
	}
	return output
}

func extractAirlines(s string) []string {
	airlineMap := map[string]struct{}{}

	// directly check if the string contains the airline
	for _, airlineItem := range tags.Airlines {
		if strings.Contains(s, airlineItem[languages.TC]) {
			airlineMap[airlineItem[languages.TC]] = struct{}{}
		}
	}

	// check if the string contains the airline alias and convert to the airline
	for key := range tags.AliasToAirlineMap {
		if strings.Contains(s, key) {
			for _, airlineName := range tags.AliasToAirlineMap[key] {
				airlineMap[airlineName] = struct{}{}
			}
		}
	}

	output := make([]string, 0, len(airlineMap))
	for key := range airlineMap {
		output = append(output, key)
	}
	return output

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
