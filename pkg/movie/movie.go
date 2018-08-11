package movie

import (
	"encoding/json"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"path/filepath"
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
	Title             string `json:"title"`
	Director          string `json:"director"`
	Country           string `json:"country"`
	Year              string `json:"year"`
	Mins              string `json:"mins"`
	MubiLink          string `json:"MUBI link"`
	MubiRating        string `json:"MUBI rating"`
	MubiRatingsNumber string `json:"MUBI ratings num"`
	ImdbRating        string `json:"IMDB rating"`
	ImdbRatingsNumber string `json:"IMDB ratings num"`
}

func init() {
	ex, err := os.Executable()
	cacheFilePath = filepath.Join(filepath.Dir(ex), cacheFileName)
	cacheErr = err
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
	if cacheErr != nil {
		return movies, cacheErr
	}

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
	columnsNo := 6

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 4, ' ', 0)
	colors := []*color.Color{color.New(color.FgWhite), color.New(color.FgGreen)}

	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))
	colors[0].Fprintln(w, "Title\tDirector\tMUBI\tIMDB\tMins\tYear\tCountry")
	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))

	var c *color.Color
	for i, m := range movies {
		c = colors[i%2]
		c.Fprintf(w, "%v\t%v\t%v (%v)\t%v (%v)\t%v\t%v\t%v\n",
			m.Title, m.Director, m.MubiRating, m.MubiRatingsNumber,
			m.ImdbRating, m.ImdbRatingsNumber, m.Mins, m.Year, m.Country)
	}
	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))

	w.Flush()
}
