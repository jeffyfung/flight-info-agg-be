package scrapper

import (
	"log"
	"regexp"
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

type result struct {
	posts []model.Post
	error error
}

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

	lastScrapDate, err := getLastScrapDate()
	if err != nil {
		log.Println("Cannot get lastScrapDate", err.Error())
		return err
	}

	ch := make(chan result)
	go scrapFlyday(ch, lastScrapDate)
	go scrapFlyAgain(ch, lastScrapDate)

	posts := []model.Post{}
	for i := 0; i < 2; i++ {
		scrappedPosts, ok := <-ch
		if !ok {
			log.Fatal("Cannot get flyDayPosts from channel")
		}
		posts = append(posts, scrappedPosts.posts...)
	}

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
		div := h.DOM
		title := div.Find(".penci-entry-title > a").Text()
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
		div := h.DOM
		title := div.Find("h2.title > a").Text()
		URL := div.AttrOr("this_url", "https://flyagain.la/")
		dateStr := strings.TrimSpace(div.Find("a.post-meta-time").Text())

		pubDate, err := time.Parse("January 02, 2006", dateStr)
		if err != nil {
			log.Println("Cannot parse dateStr", pubDate)
			ch <- result{nil, err}
			return
		}

		// TODO: can be an issue
		if pubDate.Before(lastScrapDate) {
			return
		}

		tags := []string{}
		summary := ""
		div.Find("div.blogcontent > p").Each(func(_ int, s *goquery.Selection) {
			fieldName := s.Find("span").Text()
			re, _ := regexp.Compile("航點.")
			if re.MatchString(fieldName) {
				re, _ = regexp.Compile(`\p{Han}+`)
				tags = append(tags, re.FindAllString(s.Text(), -1)...)
			}
			if strings.Contains(fieldName, "結論") {
				summary = strings.TrimSpace(s.Text())
			}
		})

		tags = append(tags, div.Find("div.post-meta > div > a").Text())

		posts = append(posts, model.Post{
			Title:     title,
			Summary:   summary,
			Tags:      tags,
			URL:       URL,
			PubDate:   pubDate,
			CreatedAt: time.Now().UTC(),
			Source:    model.DataSourceFlyAgain,
		})
	})

	c.Visit(flyAgainURL)

	ch <- result{posts, nil}
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
