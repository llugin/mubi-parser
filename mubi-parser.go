package main

import (
	"flag"
	"fmt"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/parser"
	"log"
	"time"
)

type sortValue struct {
	sortingFunc func([]movie.Data)
}

func (s *sortValue) String() string {
	return ""
}

func (s *sortValue) Set(val string) error {
	switch val {
	case "days":
		s.sortingFunc = movie.SortByDays
		return nil
	case "mubi":
		s.sortingFunc = movie.SortByMubi
		return nil
	case "imdb":
		s.sortingFunc = movie.SortByImdb
		return nil
	default:
		return fmt.Errorf("invalid sort value. Use [mubi|imdb|days]")
	}
}

func main() {

	log.SetFlags(log.Lshortfile)

	flagFromFile := flag.Bool("cached", false, "Read only data from mubi.json file - no web connection are made")
	flagNoColor := flag.Bool("no-color", false, "Disable color output")
	flagRefresh := flag.Bool("refresh", false, "Refresh all data, not only new movies")

	sv := sortValue{movie.SortByDays}
	flag.Var(&sv, "sort", "Sort by: [mubi|imdb|days], default: days")
	flag.Parse()

	var movies []movie.Data
	var err error

	start := time.Now()

	switch {
	case *flagFromFile:
		movies, err = movie.ReadFromCached()
	default:
		movies, err = parser.GetMovies(*flagRefresh)
	}

	if err != nil {
		log.Fatal(err)
	}
	sv.sortingFunc(movies)
	movie.PrintFormatted(movies, *flagNoColor)

	err = movie.WriteToCache(movies)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total time: %0.f s\n", time.Since(start).Seconds())
}
