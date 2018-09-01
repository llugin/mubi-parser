package parser

import (
	"github.com/llugin/mubi-parser/pkg/imdb"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/mubi"
	"log"
)

// GetMovies reads movie data from the web
func GetMovies(refresh bool) ([]movie.Data, error) {
	var movies []movie.Data

	out, err := mubi.ReceiveMoviesWithBasicData()
	if err != nil {
		return movies, err
	}

	if !refresh {
		cached, err := movie.ReadFromCached()
		if err != nil {
			log.Printf("%v. Could not read cached data, reading from web", err)
		} else {
			var found <-chan movie.Data
			out, found = channelCachedDetails(cached, out)
			for m := range found {
				movies = append(movies, m)
			}
		}
	}
	out = mubi.ReceiveMoviesDetails(out)
	out = imdb.GetRatings(out)

	for m := range out {
		movies = append(movies, m)
	}
	log.Printf("OMDB API called %v times\n", imdb.APICount)
	return movies, nil
}

func channelCachedDetails(vals []movie.Data, in <-chan movie.Data) (<-chan movie.Data, <-chan movie.Data) {
	new := make(chan movie.Data, mubi.MaxMovies)
	cached := make(chan movie.Data, mubi.MaxMovies)
	go func() {
		for md := range in {
			if val, found := find(md, vals); found == true {
				cached <- val
			} else {
				new <- md
			}
		}
		close(cached)
		close(new)
	}()
	return new, cached
}

func find(searched movie.Data, in []movie.Data) (movie.Data, bool) {
	for _, m := range in {
		if searched.Title == m.Title && searched.Director == m.Director {
			return m, true
		}
	}
	var empty movie.Data
	return empty, false
}
