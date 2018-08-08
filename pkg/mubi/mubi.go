package mubi

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/llugin/mubi-parser/pkg/movie"
	"io"
	"net/http"
	"strings"
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

func queryMovies(doc *goquery.Document) []movie.Data {
	var movies []movie.Data

	doc.Find(selMovie).Each(func(i int, s *goquery.Selection) {
		movie, err := queryBasic(s)
		time.Sleep(time.Second * 3)
		if err != nil {
			fmt.Println(err)
		} else {
			queryMovieDetails(&movie)
		}
		movies = append(movies, movie)
	})

	return movies
}

func queryBasic(s *goquery.Selection) (movie.Data, error) {
	var md movie.Data
	var err error

	md.Title = s.Find(selTitle).Text()
	md.Director = s.Find(selDirector).Text()

	countryAndYear := strings.Split(s.Find(selCountryAndYear).Text(), ", ")
	md.Country = countryAndYear[0]
	md.Year = countryAndYear[1]

	link, exists := s.Find(selLink).Attr("href")
	md.MubiLink = link
	if !exists {
		err = fmt.Errorf("%v: link for movie details could not be found", md.Title)
	}
	return md, err
}
func queryMovieDetails(md *movie.Data) error {
	url := baseURL + md.MubiLink
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	md.MubiRating = strings.TrimSpace(document.Find(selRating).Text())
	md.Mins = strings.TrimSpace(document.Find(selMins).Text())
	raw := document.Find(selRatingsNumber).Text()
	md.MubiRatingsNumber = strings.TrimSpace(strings.Trim(raw, "Ratings\n"))

	return nil
}

// GetBodyFromWeb gets current HTML body with shown movies
// from MUBI website. Body needs to be closed by user
func GetBodyFromWeb() (*io.ReadCloser, error) {
	resp, err := http.Get(showingURL)
	return &resp.Body, err
}

// GetMovies reads data from HTML body
func GetMovies() ([]movie.Data, error) {
	resp, err := http.Get(showingURL)
	if err != nil {
		return nil, err

	}
	defer resp.Body.Close()

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	movies := queryMovies(document)
	return movies, nil
}
