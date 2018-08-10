package main

import (
	"flag"
	"fmt"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/mubi"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)

	flagFromFile := flag.Bool("cached", false, "Read data from mubi.json file")
	flagNoColor := flag.Bool("no-color", false, "Disable color output")
	flag.Parse()

	var movies []movie.Data
	var err error

	start := time.Now()
	if *flagFromFile {
		movies, err = movie.ReadFromCached()
	} else {
		movies, err = mubi.GetMovies()
	}
	if err != nil {
		log.Fatal(err)
	}

	movie.PrintFormatted(movies, *flagNoColor)

	err = movie.WriteToCache(movies)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total time: %0.f s\n", time.Since(start).Seconds())
}
