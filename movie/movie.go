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

const jsonFileName = "mubi.json"

// JSONFilePath is a path mubi.json json file
var JSONFilePath string

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

// SetJSONFilePath sets path to json file with stored movie data
func SetJSONFilePath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(abs)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return err
	}
	JSONFilePath = abs
	return nil
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
	out, err := json.MarshalIndent(movies, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(JSONFilePath, out, 0666)
	if err != nil {
		return err
	}
	return nil
}

// ReadFromJSON reads json data from json file
func ReadFromJSON() ([]Data, error) {
	var movies []Data
	fmt.Println(JSONFilePath)
	out, err := ioutil.ReadFile(JSONFilePath)
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

// LastFromToday checks the current date, and date of newest movie appearance
// from json - if matches, returns true. Currently not implemented
// (always returns false)
func LastFromToday() bool {
	return false
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
