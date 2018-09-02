package movie

import (
	"encoding/json"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

const cacheFileName = "mubi.json"

var (
	cacheFilePath string
	cacheErr      error
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
}

func init() {
	log.SetFlags(log.Lshortfile)

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	cacheFilePath = filepath.Join(filepath.Dir(ex), cacheFileName)
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

// WriteToCache writes collected data to cache file as json
func WriteToCache(movies []Data) error {
	if cacheErr != nil {
		return cacheErr
	}
	out, err := json.MarshalIndent(movies, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cacheFilePath, out, 0666)
	if err != nil {
		return err
	}
	return nil
}

// ReadFromCached reads json data from cache file
func ReadFromCached() ([]Data, error) {
	var movies []Data

	out, err := ioutil.ReadFile(cacheFilePath)
	if err != nil {
		return movies, err
	}
	if err := json.Unmarshal(out, &movies); err != nil {
		return movies, err
	}
	return movies, nil
}

// PrintFormatted pretty-prints collected data
func PrintFormatted(movies []Data, noColor bool) {
	color.NoColor = noColor
	columnsNo := 8

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 4, ' ', 0)
	colors := []*color.Color{color.New(color.FgWhite), color.New(color.FgGreen)}

	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))
	colors[0].Fprintln(w, "Days\tTitle\tDirector\tMUBI\tIMDB\tMins\tYear\tCountry\tGenre")
	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))

	var c *color.Color
	var sb strings.Builder
	sb.WriteString(strings.Repeat("%v\t", columnsNo))
	sb.WriteString("%v\n")
	for i, m := range movies {
		c = colors[i%2]
		c.Fprintf(w, sb.String(),
			m.DaysToWatch, m.Title, m.Director, m.mubiRatingRepr(),
			m.imdbRatingRepr(), m.Mins, m.Year, m.Country, m.Genre)
	}
	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))

	w.Flush()
}

func (d *Data) mubiRatingRepr() string {
	var sb strings.Builder
	sb.WriteString(strconv.FormatFloat(d.MubiRating, 'f', 1, 32))
	sb.WriteString(" (")
	sb.WriteString(d.MubiRatingsNumber)
	sb.WriteString(")")
	return sb.String()
}

func (d *Data) imdbRatingRepr() string {
	if d.ImdbRating == 0.0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(strconv.FormatFloat(d.ImdbRating, 'f', 1, 32))
	sb.WriteString(" (")
	sb.WriteString(d.ImdbRatingsNumber)
	sb.WriteString(")")
	return sb.String()
}
