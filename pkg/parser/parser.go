package parser

import (
	"github.com/llugin/mubi-parser/pkg/imdb"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/mubi"
	"log"
	"sync"
)

// GetMovies reads movie data from the web
func GetMovies(refresh bool) ([]movie.Data, error) {
	var movies []movie.Data

	out, err := mubi.ReceiveMoviesWithBasicData()
	if err != nil {
		return movies, err
	}

	cached := make(chan movie.Data, mubi.MaxMovies)
	if !refresh {
		out, cached = channelCachedDetails(out)
	} else {
		close(cached)
	}
	out = mubi.ReceiveMoviesDetails(out)
	out = imdb.GetRatings(out)

	for m := range merge(out, cached) {
		movies = append(movies, m)
	}
	log.Printf("OMDB API called %v times\n", imdb.APICount)
	return movies, nil
}

func channelCachedDetails(in <-chan movie.Data) (<-chan movie.Data, chan movie.Data) {
	vals, err := movie.ReadFromCached()

	cached := make(chan movie.Data, mubi.MaxMovies)
	if err != nil {
		log.Printf("%v. Could not read cached data, reading from web", err)
		close(cached)
		return in, cached
	}

	new := make(chan movie.Data, mubi.MaxMovies)
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

// taken from https://blog.golang.org/pipelines
func merge(cs ...<-chan movie.Data) <-chan movie.Data {
	var wg sync.WaitGroup
	out := make(chan movie.Data)

	output := func(c <-chan movie.Data) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
