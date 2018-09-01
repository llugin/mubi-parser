package main

import (
	"flag"
	"fmt"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/parser"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)

	flagFromFile := flag.Bool("cached", false, "Read only data from mubi.json file - no web connection are made")
	flagNoColor := flag.Bool("no-color", false, "Disable color output")
	flagRefresh := flag.Bool("refresh", false, "Refresh all data, not only new movies")
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
	movie.Sort(movies)
	movie.PrintFormatted(movies, *flagNoColor)

	err = movie.WriteToCache(movies)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total time: %0.f s\n", time.Since(start).Seconds())
}
