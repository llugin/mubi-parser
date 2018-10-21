package printer

import (
	"github.com/llugin/mubi-parser/movie"
	"strconv"
	"strings"
)

type columnRepr interface {
	Header() string
	Value(*movie.Data) interface{}
}

type daysRepr struct{}

func (d daysRepr) Header() string                   { return "Days" }
func (d daysRepr) Value(md *movie.Data) interface{} { return md.DaysToWatch }

type titleRepr struct{}

func (t titleRepr) Header() string                   { return "Title" }
func (t titleRepr) Value(md *movie.Data) interface{} { return md.Title }

type directorRepr struct{}

func (d directorRepr) Header() string                   { return "Director" }
func (d directorRepr) Value(md *movie.Data) interface{} { return md.Director }

type mubiRepr struct{}

func (m mubiRepr) Header() string { return "MUBI" }
func (m mubiRepr) Value(md *movie.Data) interface{} {
	var sb strings.Builder
	sb.WriteString(strconv.FormatFloat(md.MubiRating, 'f', 1, 32))
	sb.WriteString(" (")
	sb.WriteString(md.MubiRatingsNumber)
	sb.WriteString(")")
	return sb.String()
}

type imdbRepr struct{}

func (i imdbRepr) Header() string { return "IMDB" }
func (i imdbRepr) Value(md *movie.Data) interface{} {
	if md.ImdbRating == 0.0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(strconv.FormatFloat(md.ImdbRating, 'f', 1, 32))
	sb.WriteString(" (")
	sb.WriteString(md.ImdbRatingsNumber)
	sb.WriteString(")")
	return sb.String()
}

type minsRepr struct{}

func (m minsRepr) Header() string                   { return "Mins" }
func (m minsRepr) Value(md *movie.Data) interface{} { return md.Mins }

type yearRepr struct{}

func (y yearRepr) Header() string                   { return "Year" }
func (y yearRepr) Value(md *movie.Data) interface{} { return md.Year }

type countryRepr struct{}

func (c countryRepr) Header() string                   { return "Country" }
func (c countryRepr) Value(md *movie.Data) interface{} { return md.Country }

type genreRepr struct{}

func (g genreRepr) Header() string                   { return "Genre" }
func (g genreRepr) Value(md *movie.Data) interface{} { return md.Genre }
