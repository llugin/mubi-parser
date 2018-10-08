package mubi

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/llugin/mubi-parser/debuglog"
	"github.com/llugin/mubi-parser/movie"
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
	selLink           = ".full-width-tile__link, .showing-page-hero-tile__link"
	selDaysToWatch    = ".showing-page-hero-tile__fotd-label, .full-width-tile__days-left"
	selDirector       = "[itemprop=name]"
	selCountryAndYear = ".now-showing-tile-director-year__year-country"
	selGenre          = ".film-show__genres"
	selAltTitle       = ".film-show__titles__title-alt"
	selRating         = ".average-rating__overall"
	selRatingsNumber  = ".average-rating__total"
	selMins           = "[itemprop=duration]"
)

var debug = debuglog.GetLogger()

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
		s.Each(func(i int, s *goquery.Selection) {
			movie, err := queryBasicData(s)
			if err != nil {
				debug.Println(err)
			} else {
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
				debug.Println(err)
				goto Send
			}
			defer resp.Body.Close()

			doc, err = goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				debug.Println(err)
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
		debug.Println(err)
		m.MubiRating = 0.0
	}
	m.Genre = strings.TrimSpace(doc.Find(selGenre).Text())
	m.AltTitle = strings.TrimSpace(doc.Find(selAltTitle).Text())
	minsStr := strings.TrimSpace(doc.Find(selMins).Text())
	if i, err := strconv.Atoi(minsStr); err == nil {
		m.Mins = i
	} else {
		debug.Println(err)
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
	} else {
		debug.Println(err)
	}

	daysToWatchStr := s.Find(selDaysToWatch).Text()
	if daysToWatch, err := parseDaysToWatch(daysToWatchStr); err == nil {
		md.DaysToWatch = daysToWatch
	} else {
		debug.Println(err)
	}

	link, exists := s.Find(selLink).Attr("href")
	md.MubiLink = baseURL + link
	if !exists {
		err = fmt.Errorf("%v: link for movie details could not be found", md.Title)
	}
	return md, err
}

func parseDaysToWatch(text string) (int, error) {
	text = strings.TrimSpace(text)
	if text == "Film of the day" {
		return MaxMovies, nil
	} else if text == "Expiring at midnight" {
		return 1, nil
	} else {
		return strconv.Atoi(strings.Split(text, " ")[0])
	}
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
