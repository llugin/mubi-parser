package parser

import (
	"github.com/llugin/mubi-parser/debuglog"
	"github.com/llugin/mubi-parser/imdb"
	"github.com/llugin/mubi-parser/movie"
	"github.com/llugin/mubi-parser/mubi"
	"log"
	"sync"
)

var debug = debuglog.GetLogger()

// GetMovies reads movie data from the web
func GetMovies(refresh bool) ([]movie.Data, error) {
	var movies []movie.Data

	done := make(chan struct{})
	defer close(done)

	out, err := mubi.SendMoviesWithBasicData(done)
	if err != nil {
		return movies, err
	}

	out, cached := sendCachedDetails(refresh, done, out)
	out = mubi.SendMoviesDetails(done, out)
	out = imdb.SendRatings(done, out)

	for m := range merge(done, out, cached) {
		movies = append(movies, m)
	}
	log.Printf("OMDB API called %v times\n", imdb.APICount)
	return movies, nil
}

func sendCachedDetails(refresh bool, done <-chan struct{}, in <-chan movie.Data) (<-chan movie.Data, chan movie.Data) {
	cached := make(chan movie.Data, mubi.MaxMovies)
	if refresh {
		// do nothing
		close(cached)
		return in, cached
	}

	vals, err := movie.ReadFromCached()
	if err != nil {
		debug.Printf("%v. Could not read cached data, reading from web", err)
		return in, cached
	}

	new := make(chan movie.Data, mubi.MaxMovies)
	go func() {
		defer close(new)
		defer close(cached)
		for md := range in {
			if val, found := find(md, vals); found == true {
				// Update days to watch value
				val.DaysToWatch = md.DaysToWatch
				select {
				case cached <- val:
				case <-done:
					return
				}
			} else {
				debug.Printf("Movie: %s not found in cached data\n", md.Title)
				select {
				case new <- md:
				case <-done:
					return
				}
			}
		}
	}()
	return new, cached
}

func find(searched movie.Data, in []movie.Data) (movie.Data, bool) {
	for _, m := range in {
		if searched.Title == m.Title && searched.Director == m.Director {
			return m, true
		}
	}
	return movie.Data{}, false
}

// taken from https://blog.golang.org/pipelines
func merge(done <-chan struct{}, cs ...<-chan movie.Data) <-chan movie.Data {
	var wg sync.WaitGroup
	out := make(chan movie.Data)

	output := func(c <-chan movie.Data) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
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
