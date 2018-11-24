package parser

import (
	"github.com/llugin/mubi-parser/debug"
	"github.com/llugin/mubi-parser/imdb"
	"github.com/llugin/mubi-parser/movie"
	"github.com/llugin/mubi-parser/mubi"
	"sync"
)

// GetMovies reads movie data from the web
func GetMovies(refresh bool) ([]movie.Data, error) {

	done := make(chan struct{})
	defer close(done)

	cacheSuccess := func() ([]movie.Data, bool) {
		movies, err := movie.ReadFromJSON()
		if err != nil {
			debug.Log().Printf("Could not read cached data json: %s\n", err)
			return nil, false
		}
		if movie.FromToday(movies) {
			return movies, true
		}
		return nil, false
	}

	if !refresh {
		if movies, ok := cacheSuccess(); ok {
			return movies, nil
		}
	}

	out, err := mubi.SendMoviesWithBasicData(done)
	if err != nil {
		return nil, err
	}

	out, cached := sendCachedDetails(refresh, done, out)
	out = mubi.SendMoviesDetails(done, out)
	out = imdb.SendRatings(done, out)

	var movies []movie.Data
	for m := range merge(done, out, cached) {
		movies = append(movies, m)
	}
	debug.Log().Printf("OMDB API called %v times\n", imdb.APICount)
	return movies, nil
}

func sendCachedDetails(refresh bool, done <-chan struct{}, in <-chan movie.Data) (<-chan movie.Data, chan movie.Data) {
	cached := make(chan movie.Data, mubi.MaxMovies)
	if refresh {
		// do nothing
		close(cached)
		return in, cached
	}

	vals, err := movie.ReadFromJSON()
	if err != nil {
		debug.Log().Printf("%v. Could not read cached data, reading from web", err)
		return in, cached
	}

	new := make(chan movie.Data, mubi.MaxMovies)
	go func() {
		defer close(new)
		defer close(cached)
		for md := range in {
			if val, found := movie.Find(md, vals); found == true {
				// Update days to watch value
				val.DaysToWatch = md.DaysToWatch
				select {
				case cached <- val:
				case <-done:
					return
				}
			} else {
				debug.Log().Printf("Movie: %s not found in cached data\n", md.Title)
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
