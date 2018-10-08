package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/llugin/mubi-parser/movie"
	"github.com/llugin/mubi-parser/parser"
)

func main() {
	log.SetFlags(log.Lshortfile)

	flagCached := flag.Bool("cached", false, "Read only data from mubi.json file - no web connection are made")
	flagNoColor := flag.Bool("no-color", false, "Disable color output")
	flagRefresh := flag.Bool("refresh", false, "Refresh all data, not only new movies")
	flagWatch := flag.Int("watch", -1, "Watch picked movie identified by 'Days' value")
	flagMaxLen := flag.Int("max-len", 32, "Max output table column length. Value equal or less than zero stands for unlimited length.")
	sv := sortValue{movie.SortByDays, false}
	flag.Var(&sv, "sort", "Sort by: [mubi|imdb|days|mins|year], default: days. Add '-' at argument end to reverse order")

	flag.Parse()

	start := time.Now()

	var err error
	var movies []movie.Data
	justWatch := *flagWatch != -1

	if *flagCached || justWatch {
		movies, err = movie.ReadFromCached()
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
		movie.PrintFormatted(movies, *flagNoColor, *flagMaxLen)

		if err = movie.WriteToCache(movies); err != nil {
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
