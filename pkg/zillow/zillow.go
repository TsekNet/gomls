// Package zillow scrapes real estate listings from a zillow.com
package zillow

import (
	"fmt"
	"gomls/pkg/helpers"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const (
	base = "https://www.zillow.com"
)

var (
	houses = []helpers.House{}
	c      = colly.NewCollector(
		colly.MaxDepth(1),
		colly.Async(true),
	)
)

func init() {
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error visiting %s %s\n", r.Request.URL, err.Error())
	})
	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		if strings.Contains(r.URL.String(), `/homes/`) {
			fmt.Printf("Extracting property links from %q\n", r.URL)
		} else {
			fmt.Printf("--> Extracting property data from %q \n", r.URL)
		}

		// https://www.scrapehero.com/how-to-scrape-real-estate-listings-on-zillow-com-using-python-and-lxml/
		r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		r.Headers.Set("accept-language", "en-GB;q=0.9,en-US;q=0.8,en;q=0.7")
		r.Headers.Set("dpr", "1")
		r.Headers.Set("sec-fetch-dest", "document")
		r.Headers.Set("sec-fetch-mode", "navigate")
		r.Headers.Set("sec-fetch-site", "none")
		r.Headers.Set("sec-fetch-user", "?1")
		r.Headers.Set("upgrade-insecure-requests", "1")
		r.Headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
		r.Headers.Set("usePrimedCacheWhenDisabled", "true")
	})
}

// Query that scrapes real estate listings from a website
func Query(d helpers.Details) helpers.HouseSlice {
	status := "for_sale/"
	if d.Sold {
		status = "recently_sold/"
	}
	// Either:
	// - https://www.zillow.com/homes/for_sale/<location>_rb/
	// - https://www.zillow.com/homes/recently_sold/<location>_rb/
	url := fmt.Sprintf(`%s/homes/%s%s_rb/`, base, status, d.Location)

	// This class is unique to the div that holds all information about a house
	// This filter excludes "Similar results nearby"
	c.OnHTML("ul[class*='photo-cards_extra-attribution']", func(e *colly.HTMLElement) {
		e.ForEach("a[data-test='property-card-link']", func(i int, houseElement *colly.HTMLElement) {
			pURL := houseElement.Attr("href")
			c.Visit(pURL)
		})
	})

	// __NEXT_DATA__ is a JSON-like object that contains metadata about the property
	c.OnHTML("script#__NEXT_DATA__", func(e *colly.HTMLElement) {
		// Only grab data from the property pages, ignore search results page
		if !strings.Contains(e.Request.URL.Path, "/homedetails") {
			return
		}
		h := helpers.House{}

		dataMap := helpers.JsonToMap(e.Text)

		url = dataMap["hdpUrl"]
		fullAddress := strings.SplitN(url, "/", -1)[2]

		m := map[string]string{
			"address":      strings.ReplaceAll(fullAddress, "-", " "),
			"baths":        strings.TrimSuffix(dataMap["bathrooms"], ","),
			"beds":         strings.TrimSuffix(dataMap["bd"], ","),
			"description":  dataMap["description"],
			"img":          dataMap["url"],
			"link":         base + url,
			"listdate":     dataMap["datePostedString"],
			"listprice":    strings.TrimSuffix(dataMap["listPrice"], ","),
			"mapsurl":      fmt.Sprintf("https://maps.google.com/?q=%s", fullAddress),
			"showing":      fmt.Sprintf("%s - %s", dataMap["startTime"], dataMap["endTime"]),
			"size":         fmt.Sprintf("%s %s", strings.TrimSuffix(dataMap["livingAreaValue"], ","), dataMap["livingAreaUnitsShort"]),
			"solddate":     helpers.MSToTime(strings.TrimSuffix(dataMap["dateSold"], ",")),
			"soldprice":    strings.TrimSuffix(dataMap["lastSoldPrice"], ","),
			"status":       dataMap["keystoneHomeStatus"],
			"propertytype": dataMap["homeType"],
		}

		// Fields that need no manipulation
		h.Address = m["address"]
		h.Link = m["link"]
		h.Img = m["img"]
		h.MapsURL = m["mapsurl"]
		h.Status = m["status"]

		// Args passed via flags
		h.Beds = m["beds"]
		if be, err := strconv.Atoi(h.Beds); err != nil || be < d.Beds {
			return
		}

		h.Baths = m["baths"]
		if ba, err := strconv.Atoi(h.Baths); err != nil || ba < d.Baths {
			return
		}

		h.PropertyType = m["propertytype"]
		if d.PropertyType != "" && h.PropertyType != d.PropertyType {
			return
		}

		// The rest of the fields need some manipulation
		if !strings.Contains(m["description"], "null") {
			h.Description = m["description"]
		}

		if m["showing"] != " - " {
			h.Showing = m["showing"]
		}

		if !strings.Contains(m["size"], "null") && !strings.HasPrefix(m["size"], "0") {
			h.Size = m["size"]
		}

		// TODO: Fix the list and sold dates
		if d.Sold {
			h.SoldPrice = m["soldprice"]
			h.SoldDate = m["solddate"]

			if h.SoldPrice != "" {
				// Calculate the price difference
				sp, _ := strconv.Atoi(h.SoldPrice)
				lp, _ := strconv.Atoi(h.ListPrice)
				h.PriceDiff = fmt.Sprintf("%d", sp-lp)
			}
		}

		// Passed via Flag
		h.ListPrice = m["listprice"]
		if lp, err := strconv.Atoi(h.ListPrice); err != nil || lp < d.Price {
			return
		}
		h.ListDate = m["listdate"]

		houses = append(houses, h)
	})

	c.Visit(url)
	c.Wait()

	if len(houses) == 0 {
		fmt.Println("Property search returned no results")
	}
	return houses
}
