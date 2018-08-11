package parser

import (
	"github.com/llugin/mubi-parser/pkg/imdb"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/mubi"
	"log"
)

// GetMovies reads movie data from HTML body
func GetMovies() ([]movie.Data, error) {
	var movies []movie.Data

	moviesChan, err := mubi.ReceiveMoviesWithBasicData()
	if err != nil {
		return movies, err
	}
	out := mubi.ReceiveMoviesDetails(moviesChan)
	out = imdb.GetRatings(out)

	for m := range out {
		movies = append(movies, m)
	}
	log.Printf("OMDB API called %v times\n", imdb.APICount)
	return movies, nil
}
