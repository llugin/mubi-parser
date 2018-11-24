package movie

import (
	"encoding/json"
	"fmt"
	"github.com/llugin/mubi-parser/debug"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

const (
	jsonFileName = "mubi.json"

	// time layout for data values
	layout = "2006-1-2"
)

// JSONFilePath is a path mubi.json json file
var (
	JSONPath = ""
)

// Data represent movie data collected by parser
type Data struct {
	Title             string  `json:"title"`
	Director          string  `json:"director"`
	Country           string  `json:"country"`
	Year              int     `json:"year,string"`
	Genre             string  `json:"genre"`
	Mins              int     `json:"mins,string"`
	AltTitle          string  `json:"alt title"`
	MubiLink          string  `json:"MUBI link"`
	MubiRating        float64 `json:"MUBI rating,string"`
	MubiRatingsNumber string  `json:"MUBI ratings num"`
	ImdbRating        float64 `json:"IMDB rating,string"`
	ImdbRatingsNumber string  `json:"IMDB ratings num"`
	DaysToWatch       int     `json:"days,string"`
	DateAppeared      string  `json:"appeared"`
}

func init() {
	log.SetFlags(log.Lshortfile)
}

func jsonfile() string {
	return filepath.Join(JSONPath, jsonFileName)

}

// AbbrevCountry abbreviates names of selected countries
func (d *Data) AbbrevCountry() {
	switch d.Country {
	case "United States":
		d.Country = "USA"
	case "United Kingdom":
		d.Country = "UK"
	case "Soviet Union":
		d.Country = "USSR"
	case "South Africa":
		d.Country = "RSA"
	default:
		return
	}
}

// Watch opens movie page in default browser
func (d *Data) Watch() error {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("Watch function not supported on this OS")
	}

	return exec.Command(cmd, d.MubiLink).Run()
}

// WriteToJSON writes collected data to json file as json
func WriteToJSON(movies []Data) error {
	SortByDays(movies)
	out, err := json.MarshalIndent(movies, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(jsonfile(), out, 0666)
	if err != nil {
		return err
	}
	return nil
}

// ReadFromJSON reads json data from json file
func ReadFromJSON() ([]Data, error) {
	var movies []Data
	out, err := ioutil.ReadFile(jsonfile())
	if err != nil {
		return movies, err
	}
	if err := json.Unmarshal(out, &movies); err != nil {
		return movies, err
	}
	return movies, nil
}

// FindByDay return movie based by day
func FindByDay(day int, movies []Data) (Data, error) {
	for _, m := range movies {
		if m.DaysToWatch == day {
			return m, nil
		}
	}
	return Data{}, fmt.Errorf("Movie not found")
}

// FromToday checks the current date, and date of newest movie appearance
// from json - if matches, returns true. Currently not implemented
// (always returns false)
func FromToday(movies []Data) bool {

	today := time.Now()
	lastMovie, err := FindByDay(30, movies)
	if err != nil {
		debug.Log().Printf("Could not find movie with 30 days left, %v\n", err)
		return false
	}

	last, err := time.Parse(layout, lastMovie.DateAppeared)
	if err != nil {
		debug.Log().Printf("Could not parse movie date, %v\n", err)
		return false
	}
	return today.Year() == last.Year() && today.YearDay() == last.YearDay()
}

// SetDateAppeared sets appearance date string in recognized layout
func (d *Data) SetDateAppeared(retrieved time.Time) {
	d.DateAppeared = retrieved.AddDate(0, 0, d.DaysToWatch-30).Format(layout)
}

// ParseDateAppeared returns date parsed from string
func (d *Data) ParseDateAppeared() (time.Time, error) {
	date, err := time.Parse(layout, d.DateAppeared)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

// Find searches for movie in movie slice
func Find(searched Data, in []Data) (Data, bool) {
	for _, m := range in {
		if searched.Title == m.Title && searched.Director == m.Director {
			return m, true
		}
	}
	return Data{}, false
}

// SortByDays sorts slice of movies by days to watch
func SortByDays(movies []Data) {
	sort.Slice(movies, func(i, j int) bool {
		return movies[i].DaysToWatch > movies[j].DaysToWatch
	})
}

// SortByImdb sorts slice of movies by IMDB rating
func SortByImdb(movies []Data) {
	sort.Slice(movies, func(i, j int) bool {
		return movies[i].ImdbRating > movies[j].ImdbRating
	})
}

// SortByMubi sorts slice of movies by MUBI rating
func SortByMubi(movies []Data) {
	sort.Slice(movies, func(i, j int) bool {
		return movies[i].MubiRating > movies[j].MubiRating
	})
}

// SortByMins sorts slice of movies by duration
func SortByMins(movies []Data) {
	sort.Slice(movies, func(i, j int) bool {
		return movies[i].Mins > movies[j].Mins
	})
}

// SortByYear sorts slice of movies by year
func SortByYear(movies []Data) {
	sort.Slice(movies, func(i, j int) bool {
		return movies[i].Year > movies[j].Year
	})
}
