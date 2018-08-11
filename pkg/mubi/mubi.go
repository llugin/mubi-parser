package mubi

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/llugin/mubi-parser/pkg/movie"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// MaxMovies is a number of currently available movies
	MaxMovies  = 30
	showingURL = "https://mubi.com/showing"
	baseURL    = "https://mubi.com"

	// mubi goquery selection queries
	selMovie          = ".full-width-tile--now-showing, .showing-page-hero-tile"
	selTitle          = ".full-width-tile__title, .showing-page-hero-tile__title"
	selDirector       = "[itemprop=name]"
	selCountryAndYear = ".now-showing-tile-director-year__year-country"
	selGenre          = ".film-show__genres"
	selAltTitle       = ".film-show__titles__title-alt"
	selLink           = ".full-width-tile__link, .showing-page-hero-tile__link"
	selRating         = ".average-rating__overall"
	selRatingsNumber  = ".average-rating__total"
	selMins           = "[itemprop=duration]"
)

// ReceiveMoviesWithBasicData returns a buffered channel with
// movies with basic data available to collect from mubi main page
func ReceiveMoviesWithBasicData() (<-chan movie.Data, error) {
	moviesChan := make(chan movie.Data, MaxMovies)

	s, err := getSelectionFromWebPage()
	if err != nil {
		close(moviesChan)
		return moviesChan, err
	}

	go func() {
		s.Each(func(i int, s *goquery.Selection) {
			movie, err := queryBasicData(s)
			if err != nil {
				fmt.Println(err)
			} else {
				moviesChan <- movie
			}
		})
		close(moviesChan)
	}()
	return moviesChan, nil
}

//ReceiveMoviesDetails returns channel with movies with detailed data
func ReceiveMoviesDetails(in <-chan movie.Data) <-chan movie.Data {
	out := make(chan movie.Data, MaxMovies)
	go func() {
		for md := range in {
			time.Sleep(time.Second * 3)

			url := baseURL + md.MubiLink
			resp, err := http.Get(url)
			if err != nil {
				out <- md
				continue
			}
			defer resp.Body.Close()

			document, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				out <- md
				continue
			}

			md.MubiRating = strings.TrimSpace(document.Find(selRating).Text())
			md.Genre = strings.TrimSpace(document.Find(selGenre).Text())
			md.AltTitle = strings.TrimSpace(document.Find(selAltTitle).Text())
			md.Mins = strings.TrimSpace(document.Find(selMins).Text())
			raw := document.Find(selRatingsNumber).Text()
			md.MubiRatingsNumber = strings.TrimSpace(strings.Trim(raw, "Ratings\n"))

			out <- md
		}
		close(out)
	}()
	return out
}

func queryBasicData(s *goquery.Selection) (movie.Data, error) {
	var md movie.Data
	var err error

	md.Title = s.Find(selTitle).Text()
	md.Director = s.Find(selDirector).Text()

	countryAndYear := strings.Split(s.Find(selCountryAndYear).Text(), ", ")
	md.Country = countryAndYear[0]
	md.AbbrevCountry()
	year, err := strconv.Atoi(countryAndYear[1])
	if err == nil {
		md.Year = year
	}

	link, exists := s.Find(selLink).Attr("href")
	md.MubiLink = link
	if !exists {
		err = fmt.Errorf("%v: link for movie details could not be found", md.Title)
	}
	return md, err
}

func getSelectionFromWebPage() (*goquery.Selection, error) {
	resp, err := http.Get(showingURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc.Find(selMovie), nil
}
