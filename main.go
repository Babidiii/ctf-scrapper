package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/google/uuid"
)

type Ctf struct {
	Id      uuid.UUID `json:"id"`
	Place   int       `json:"place"`
	Name    string    `json:"name"`
	Points  float64   `json:"points"`
	Ratings float64   `json:"ratings"`
}

func NewCtf(place int, name string, points float64, ratings float64) *Ctf {
	return &Ctf{
		Id:      uuid.New(),
		Place:   place,
		Name:    name,
		Points:  points,
		Ratings: ratings,
	}
}

type Season struct {
	Year         *int     `json:"year,omitempty"`
	Place        *int     `json:"place,omitempty"`
	Points       *float64 `json:"points,omitempty"`
	CountryPlace *int     `json:"countryPlace,omitempty"`
	Ctfs         []*Ctf   `json:"ctfs,omitempty"`
}

type Stats struct {
	Seasons []*Season `json:"seasons"`
}

func parseStringToFloat64Pt(s string) *float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func parseStringToIntPt(s string) *int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &i
}

// CreatedDate time.Time `json:"createdDate"`
func main() {
	// team_id := 155626
	// string_teamId := strconv.Itoa(team_id)
	if len(os.Args[1:]) != 1 {
		log.Fatalln("Need an argument")
	}
	string_teamId := os.Args[1]

	log.Println("Collector initialization")
	collector := colly.NewCollector(
		colly.CacheDir("./cache"),
	)
	extensions.RandomUserAgent(collector) // Use random agent
	stats := &Stats{}

	collector.OnHTML(".container:nth-child(3) > div:nth-child(5) > .tab-pane", func(e *colly.HTMLElement) {
		year := strings.SplitN(e.Attr("id"), "_", 2)[1]
		place := e.ChildText("p:first-of-type > b:first-of-type")
		points := e.ChildText("p:first-of-type > b:last-of-type")
		countryPlace := e.ChildText("p:last-of-type > b:first-of-type")

		season_points := parseStringToFloat64Pt(points)
		season_year := parseStringToIntPt(year)
		season_place := parseStringToIntPt(place)
		season_countryPlace := parseStringToIntPt(countryPlace)

		season := &Season{
			Year:         season_year,
			Place:        season_place,
			Points:       season_points,
			CountryPlace: season_countryPlace,
			Ctfs:         make([]*Ctf, 0, 10),
		}

		e.ForEach("tr", func(ind int, item *colly.HTMLElement) {
			fall := item.ChildTexts("td")
			if len(fall) > 0 {
				pts := parseStringToFloat64Pt(fall[3])
				rgs := parseStringToFloat64Pt(fall[4])
				plc := parseStringToIntPt(fall[1])
				season.Ctfs = append(season.Ctfs, (NewCtf(*plc, fall[2], *pts, *rgs)))
			}
		})
		stats.Seasons = append(stats.Seasons, season)
	})

	// log on response
	collector.OnResponse(func(r *colly.Response) {
		log.Println("Response received", r.StatusCode)
	})

	// Before making a request print "Visiting ..."
	collector.OnRequest(func(r *colly.Request) {
		log.Println("Scrapping", r.URL.String())
	})

	collector.Visit("https://ctftime.org/team/" + string_teamId)

	// Waiting that collector job is finished
	collector.Wait()

	filename := "./output/" + string_teamId + ".json"
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	data, err := json.MarshalIndent(stats, "", " ")
	if err != nil {
		log.Fatalln(err)
	}

	_, err = file.Write(data)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Json file", filename, "generated")
}
