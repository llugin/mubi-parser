package movie

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
)

const cacheFileName = "mubi.json"

// CacheFilePath is a path mubi.json cache file
var CacheFilePath string

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
}

func init() {
	log.SetFlags(log.Lshortfile)

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	CacheFilePath = filepath.Join(filepath.Dir(ex), cacheFileName)
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

// WriteToCache writes collected data to cache file as json
func WriteToCache(movies []Data) error {
	out, err := json.MarshalIndent(movies, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(CacheFilePath, out, 0666)
	if err != nil {
		return err
	}
	return nil
}

// ReadFromCached reads json data from cache file
func ReadFromCached() ([]Data, error) {
	var movies []Data

	out, err := ioutil.ReadFile(CacheFilePath)
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
