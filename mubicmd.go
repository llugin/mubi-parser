package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/llugin/mubi-parser/debugging"
	"github.com/llugin/mubi-parser/imdb"
	"github.com/llugin/mubi-parser/movie"
	"github.com/llugin/mubi-parser/mubi"
	"github.com/llugin/mubi-parser/parser"
	"github.com/llugin/mubi-parser/printer"
)

const jsonFileName = "mubi.json"

type config struct {
	OMDBKey  string `json:"OMDBKey"`
	DataPath string `json:"DataPath"`
	LogPath  string `json:"LogPath"`
}

func readConfig() (config, error) {
	ex, err := os.Executable()
	if err != nil {
		return config{}, err
	}
	cwd := filepath.Dir(ex)
	configPath := filepath.Join(cwd, "mubiconf.json")
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		// Set default values
		return config{"", cwd, cwd}, nil
	}

	out, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config{}, err
	}
	var c config
	err = json.Unmarshal(out, &c)
	return c, err
}

func main() {
	log.SetFlags(log.Lshortfile)

	flagCached := flag.Bool("cached", false, "Read only data from mubi.json file - no web connection are made")
	flagStderrLog := flag.Bool("stderr-debug", false, "Print debug info to stderr")
	flagMubiSleep := flag.Int("mubi-sleep", 3, "Sleep between mubi HTTP requests in seconds")
	flagImdbSleep := flag.Int("imdb-sleep", 200, "Sleep between OMDB API calls in milliseconds")
	flagNoColor := flag.Bool("no-color", false, "Disable color output")
	flagRefresh := flag.Bool("refresh", false, "Refresh all data, not only new movies")
	flagWatch := flag.Int("watch", -1, "Watch picked movie identified by 'Days' value")
	flagMaxLen := flag.Int("max-len", 32, "Max output table column length. Value equal or less than zero stands for unlimited length.")
	sv := sortValue{movie.SortByDays, false}
	flag.Var(&sv, "sort", "Sort by: [mubi|imdb|days|mins|year], default: days. Add '-' at argument end to reverse order")

	flag.Parse()
	conf, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	movie.JSONPath = conf.DataPath
	imdb.APIKey = conf.OMDBKey
	debugging.InitLogger(conf.LogPath, *flagStderrLog)

	imdb.Sleep = *flagImdbSleep
	mubi.Sleep = *flagMubiSleep

	start := time.Now()

	var movies []movie.Data
	justWatch := *flagWatch != -1

	if *flagCached || justWatch {
		movies, err = movie.ReadFromJSON()
	} else {
		movies, err = parser.GetMovies(*flagRefresh)
	}
	if err != nil {
		log.Fatal(err)
	}

	if justWatch {
		if err := watch(movies, *flagWatch); err != nil {
			log.Fatal(err)
		}
	} else {
		sv.sort(movies)
		printer.PrintTable(movies, *flagNoColor, *flagMaxLen)

		if err = movie.WriteToJSON(movies); err != nil {
			log.Fatal(err)
		}

		log.Printf("Total time: %0.f s\n", time.Since(start).Seconds())
	}
}

func watch(movies []movie.Data, day int) error {
	m, err := movie.FindByDay(day, movies)
	if err != nil {
		return err
	}
	return m.Watch()
}

type sortValue struct {
	sortingFunc func([]movie.Data)
	reversed    bool
}

func (s *sortValue) String() string {
	return ""
}

func (s *sortValue) sort(m []movie.Data) {
	s.sortingFunc(m)
	if s.reversed {
		for i, j := 0, len(m)-1; i < j; i, j = i+1, j-1 {
			m[i], m[j] = m[j], m[i]
		}
	}
}

func (s *sortValue) Set(val string) error {
	if val[len(val)-1:] == "-" {
		s.reversed = true
		val = val[:len(val)-1]
	} else {
		s.reversed = false
	}

	switch val {
	case "days":
		s.sortingFunc = movie.SortByDays
	case "mubi":
		s.sortingFunc = movie.SortByMubi
	case "imdb":
		s.sortingFunc = movie.SortByImdb
	case "mins":
		s.sortingFunc = movie.SortByMins
	case "year":
		s.sortingFunc = movie.SortByYear
	default:
		return fmt.Errorf("Undefined sort parameter")
	}

	return nil
}
