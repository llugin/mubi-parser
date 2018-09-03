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

// SendMoviesWithBasicData returns a buffered channel with
// movies with basic data available to collect from mubi main page
func SendMoviesWithBasicData(done <-chan struct{}) (<-chan movie.Data, error) {
	moviesChan := make(chan movie.Data, MaxMovies)

	s, err := getSelectionFromWebPage()
	if err != nil {
		return moviesChan, err
	}

	go func() {
		defer close(moviesChan)
		daysToWatch := MaxMovies
		s.Each(func(i int, s *goquery.Selection) {
			movie, err := queryBasicData(s)
			if err != nil {
				fmt.Println(err)
			} else {
				movie.DaysToWatch = daysToWatch
				daysToWatch--
				select {
				case moviesChan <- movie:
				case <-done:
					return
				}
			}
		})
	}()
	return moviesChan, nil
}

//SendMoviesDetails returns channel with movies with detailed data
func SendMoviesDetails(done <-chan struct{}, in <-chan movie.Data) <-chan movie.Data {
	var doc *goquery.Document
	var resp *http.Response
	var err error

	out := make(chan movie.Data, MaxMovies)

	go func() {
		defer close(out)
		for md := range in {
			time.Sleep(time.Second * 3)

			resp, err = http.Get(md.MubiLink)
			if err != nil {
				goto Send
			}
			defer resp.Body.Close()

			doc, err = goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				goto Send
			}
			acquireDetailsFromDocument(&md, doc)

		Send:
			select {
			case out <- md:
			case <-done:
				return
			}
		}
	}()
	return out
}

func acquireDetailsFromDocument(m *movie.Data, doc *goquery.Document) {
	ratingStr := strings.TrimSpace(doc.Find(selRating).Text())
	if f, err := strconv.ParseFloat(ratingStr, 32); err == nil {
		m.MubiRating = f
	} else {
		m.MubiRating = 0.0
	}
	m.Genre = strings.TrimSpace(doc.Find(selGenre).Text())
	m.AltTitle = strings.TrimSpace(doc.Find(selAltTitle).Text())
	minsStr := strings.TrimSpace(doc.Find(selMins).Text())
	if i, err := strconv.Atoi(minsStr); err == nil {
		m.Mins = i
	}
	raw := doc.Find(selRatingsNumber).Text()
	m.MubiRatingsNumber = strings.TrimSpace(strings.Trim(raw, "Ratings\n"))

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
	md.MubiLink = baseURL + link
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
