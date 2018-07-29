package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	showingURL = "https://mubi.com/showing"
	baseURL    = "https://mubi.com"

	// mubi goquery selection queries
	selMovie          = ".full-width-tile--now-showing"
	selTitle          = ".full-width-tile__title"
	selDirector       = "[itemprop=name]"
	selCountryAndYear = ".now-showing-tile-director-year__year-country"
	selLink           = ".full-width-tile__link"
	selRating         = ".average-rating__overall"
	selRatingsNumber  = ".average-rating__total"
	selDuration       = "[itemprop=duration]"
)

var cacheFile = filepath.Join(os.Getenv("GOPATH"), "mubi.json")

type movieData struct {
	Title             string `json:"title"`
	Director          string `json:"director"`
	Country           string `json:"country"`
	Year              string `json:"year"`
	Duration          string `json:"duration"`
	MubiRating        string `json:"MUBI rating"`
	MubiRatingsNumber string `json:"MUBI ratings num"`
}

func queryMovieDetails(url string, md *movieData) {
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
	md.Duration = strings.TrimSpace(document.Find(selDuration).Text())
	raw := document.Find(selRatingsNumber).Text()
	md.MubiRatingsNumber = strings.TrimSpace(strings.Trim(raw, "Ratings\n"))
}

func queryMovies(doc *goquery.Document) []movieData {
	var movies []movieData

	doc.Find(selMovie).Each(func(i int, s *goquery.Selection) {
		var md movieData

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

func readFromCached() []movieData {
	out, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		log.Fatal("cache file could not be opened ", err)
	}
	var movies []movieData
	if err := json.Unmarshal(out, &movies); err != nil {
		log.Fatal("Could not read data from cache ", err)
	}
	return movies
}

func readFromWebPage() []movieData {
	resp, err := http.Get(showingURL)
	if err != nil {
		log.Fatal("Error loading html body ", err)
	}
	defer resp.Body.Close()
	time.Sleep(time.Second * 3)
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}
	return queryMovies(document)
}

func main() {
	fromFile := flag.Bool("cached", false, "Read data from mubi.json file")
	flag.Parse()

	var movies []movieData

	if *fromFile {
		movies = readFromCached()
	} else {
		movies = readFromWebPage()
	}

	out, err := json.MarshalIndent(movies, "", " ")
	if err != nil {
		log.Fatal("json could not be marshalled ", err)
	}

	fmt.Println(string(out))

	err = ioutil.WriteFile(cacheFile, out, 0666)
	if err != nil {
		log.Fatal("Could not write to mubi.json file ", err)
	}
}
