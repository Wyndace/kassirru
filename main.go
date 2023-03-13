package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type responseType struct {
	HTML        string `json:"html"`
	MoreResults string `json:"more_results"`
}

type CardType struct {
	Title     string `json:"eventName"`
	Datetime  any    `json:"date"`
	Place     string `json:"venueName"`
	PriceMin  int64  `json:"minPrice"`
	PriceMax  int64  `json:"maxPrice"`
	Link      string `json:"link"`
	PlaceLink string `json:"placeLink"`
	ImageLink string `json:"image"`
}

func main() {
	c := colly.NewCollector()
	var afisha []CardType
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		fmt.Println(r.URL)
	})
	c.OnResponse(func(r *colly.Response) {
		var response responseType
		err := json.Unmarshal(r.Body, &response)
		if err != nil {
			fmt.Println(err, r.Body)
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(response.HTML))
		if err != nil {
			fmt.Println(err)
		}
		doc.Find(".event-card").Each(func(_ int, sel *goquery.Selection) {
			var card CardType
			card.Link, _ = sel.Find(".title a").First().Attr("href")
			card.PlaceLink, _ = sel.Find(".venue a").First().Attr("href")
			raw, _ := sel.Find("[data-ec-item]").First().Attr("data-ec-item")
			err := json.Unmarshal([]byte(raw), &card)
			if err != nil {
				fmt.Println(err)
				card.Title = strings.TrimSpace(sel.Find(".title").Text())
				_, card.Datetime, _ = strings.Cut(card.Link, "_")
				PriceMinRaw, PriceMaxRaw, _ := strings.Cut(sel.Find("cost").Text(), " — ")
				card.PriceMin, _ = strconv.ParseInt(strings.TrimSpace(PriceMinRaw), 10, 64)
				card.PriceMax, _ = strconv.ParseInt(strings.TrimSpace(PriceMaxRaw), 10, 64)
				card.ImageLink, _ = sel.Find(".poster img").Attr("data-src")
			}
			afisha = append(afisha, card)
			c.Visit(r.Request.AbsoluteURL(response.MoreResults))
		})
	})
	c.Visit("https://msk.kassir.ru/bilety-na-koncert?p=1")
	afishaByte, _ := json.Marshal(afisha)
	ioutil.WriteFile("result.json", afishaByte, os.ModePerm)
}
