package mubi

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	showingURL = "https://mubi.com/showing"
	baseURL    = "https://mubi.com"

	// mubi goquery selection queries
	selMovie          = ".full-width-tile--now-showing, .showing-page-hero-tile"
	selTitle          = ".full-width-tile__title, .showing-page-hero-tile__title"
	selDirector       = "[itemprop=name]"
	selCountryAndYear = ".now-showing-tile-director-year__year-country"
	selLink           = ".full-width-tile__link, .showing-page-hero-tile__link"
	selRating         = ".average-rating__overall"
	selRatingsNumber  = ".average-rating__total"
	selMins           = "[itemprop=duration]"
)

var cacheFile = filepath.Join(os.Getenv("GOPATH"), "mubi.json")

// MovieData represent data collected by parser about movie
type MovieData struct {
	Title             string `json:"title"`
	Director          string `json:"director"`
	Country           string `json:"country"`
	Year              string `json:"year"`
	Mins              string `json:"mins"`
	MubiRating        string `json:"MUBI rating"`
	MubiRatingsNumber string `json:"MUBI ratings num"`
}

func queryMovieDetails(url string, md *MovieData) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Error loading html body", err)
	}
	defer resp.Body.Close()

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	md.MubiRating = strings.TrimSpace(document.Find(selRating).Text())
	md.Mins = strings.TrimSpace(document.Find(selMins).Text())
	raw := document.Find(selRatingsNumber).Text()
	md.MubiRatingsNumber = strings.TrimSpace(strings.Trim(raw, "Ratings\n"))
}

func queryMovies(doc *goquery.Document) []MovieData {
	var movies []MovieData

	doc.Find(selMovie).Each(func(i int, s *goquery.Selection) {
		var md MovieData

		md.Title = s.Find(selTitle).Text()
		md.Director = s.Find(selDirector).Text()

		countryAndYear := strings.Split(s.Find(selCountryAndYear).Text(), ", ")
		md.Country = countryAndYear[0]
		md.Year = countryAndYear[1]

		link, exists := s.Find(selLink).Attr("href")
		if !exists {
			log.Fatal("link for movie does not exist")
		}

		time.Sleep(time.Second * 3)

		queryMovieDetails(baseURL+link, &md)
		movies = append(movies, md)
	})

	return movies
}

// GetBodyFromWeb gets current HTML body with shown movies
// from MUBI website. Body needs to be closed by user
func GetBodyFromWeb() *io.ReadCloser {
	resp, err := http.Get(showingURL)
	if err != nil {
		log.Fatal("Error loading html body ", err)
	}
	return &resp.Body
}

// ReadFromCached reads json data from cache file
func ReadFromCached() []MovieData {
	out, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		log.Fatal("cache file could not be opened ", err)
	}
	var movies []MovieData
	if err := json.Unmarshal(out, &movies); err != nil {
		log.Fatal("Could not read data from cache ", err)
	}
	return movies
}

// ReadFromBody reads data from HTML body
func ReadFromBody(rc *io.ReadCloser) []MovieData {
	document, err := goquery.NewDocumentFromReader(*rc)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}
	return queryMovies(document)
}

// PrintFormatted pretty-prints collected data
func PrintFormatted(movies []MovieData, noColor bool) {

	color.NoColor = noColor
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 4, ' ', 0)
	colors := []*color.Color{color.New(color.FgWhite), color.New(color.FgGreen)}

	colors[0].Fprintln(w, strings.Repeat("\t", 5))
	colors[0].Fprintln(w, "Title\tDirector\tRating\tMins\tYear\tCountry")
	colors[0].Fprintln(w, strings.Repeat("\t", 5))

	var c *color.Color
	for i, m := range movies {
		c = colors[i%2]
		c.Fprintf(w, "%v\t%v\t%v (%v)\t%v\t%v\t%v\n",
			m.Title, m.Director, m.MubiRating, m.MubiRatingsNumber,
			m.Mins, m.Year, m.Country)
	}
	colors[0].Fprintln(w, strings.Repeat("\t", 5))

	w.Flush()
}

// WriteToCache writes collected data to cache file as json
func WriteToCache(movies []MovieData) {
	out, err := json.MarshalIndent(movies, "", " ")
	if err != nil {
		log.Fatal("json could not be marshalled ", err)
	}
	err = ioutil.WriteFile(cacheFile, out, 0666)
	if err != nil {
		log.Fatal("Could not write to mubi.json file ", err)
	}
}
