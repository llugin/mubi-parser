package parser

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/llugin/mubi-parser/imdb"
	"github.com/llugin/mubi-parser/movie"
	"github.com/llugin/mubi-parser/mubi"
)

func queryMovies(doc *goquery.Document) []movie.Data {
	var movies []movie.Data

	doc.Find(selMovie).Each(func(i int, s *goquery.Selection) {
		movie := mubi.GetBasicInfo(s)
		time.Sleep(time.Second * 3)
		imdb.GetRatings(&movie)

		mubi.GetDetails(&movie)
		movies = append(movies, movie)
	})

	return movies
}
