package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
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

type movieData struct {
	title             string
	director          string
	country           string
	year              string
	duration          string
	mubiRating        string
	mubiRatingsNumber string
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

	md.mubiRating = strings.TrimSpace(document.Find(selRating).Text())
	md.duration = strings.TrimSpace(document.Find(selDuration).Text())
	raw := document.Find(selRatingsNumber).Text()
	md.mubiRatingsNumber = strings.TrimSpace(strings.Trim(raw, "Ratings\n"))
}

func queryMovies(doc *goquery.Document) []movieData {
	var movies []movieData

	doc.Find(selMovie).Each(func(i int, s *goquery.Selection) {
		var md movieData

		md.title = s.Find(selTitle).Text()
		md.director = s.Find(selDirector).Text()

		countryAndYear := strings.Split(s.Find(selCountryAndYear).Text(), ", ")
		md.country = countryAndYear[0]
		md.year = countryAndYear[1]

		link, exists := s.Find(selLink).Attr("href")
		if !exists {
			log.Fatal("link for movie does not exist")
		}

		time.Sleep(time.Second * 3)

		queryMovieDetails(baseURL+link, &md)
		fmt.Printf("%+v\n", md)
		movies = append(movies, md)
	})

	return movies
}

func main() {
	resp, err := http.Get(showingURL)
	if err != nil {
		log.Fatal("Error loading html body", err)
	}
	defer resp.Body.Close()

	time.Sleep(time.Second * 3)

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	queryMovies(document)
}
