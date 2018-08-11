package parser

import (
	"github.com/llugin/mubi-parser/pkg/imdb"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/mubi"
)

// GetMovies reads movie data from HTML body
func GetMovies() ([]movie.Data, error) {
	var movies []movie.Data

	moviesChan, err := mubi.ReceiveMoviesWithBasicData()
	if err != nil {
		return movies, err
	}
	mubiOut := mubi.ReceiveMoviesDetails(moviesChan)
	imdbOut := imdb.GetRatings(mubiOut)

	for m := range imdbOut {
		movies = append(movies, m)
	}
	return movies, nil
}
