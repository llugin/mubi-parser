package main

import (
	"flag"
	"github.com/llugin/mubi-parser/pkg/mubi"
)

func main() {
	flagFromFile := flag.Bool("cached", false, "Read data from mubi.json file")
	flagNoColor := flag.Bool("no-color", false, "Disable color output")
	flag.Parse()

	var movies []mubi.MovieData

	if *flagFromFile {
		movies = mubi.ReadFromCached()
	} else {
		body := mubi.GetBodyFromWeb()
		defer (*body).Close()
		movies = mubi.ReadFromBody(body)
	}

	mubi.WriteToCache(movies)
	mubi.PrintFormatted(movies, *flagNoColor)
}
