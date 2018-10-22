package printer

import (
	"github.com/llugin/mubi-parser/movie"
	"strconv"
	"strings"
)

var columns = []columnRepr{
	days{}, title{}, director{}, mubi{}, imdb{}, mins{}, year{}, country{}, genre{}}

type columnRepr interface {
	Header() string
	Value(*movie.Data) interface{}
}

type days struct{}

func (d days) Header() string                   { return "Days" }
func (d days) Value(md *movie.Data) interface{} { return md.DaysToWatch }

type title struct{}

func (t title) Header() string                   { return "Title" }
func (t title) Value(md *movie.Data) interface{} { return md.Title }

type director struct{}

func (d director) Header() string                   { return "Director" }
func (d director) Value(md *movie.Data) interface{} { return md.Director }

type mubi struct{}

func (m mubi) Header() string { return "MUBI" }
func (m mubi) Value(md *movie.Data) interface{} {
	var sb strings.Builder
	sb.WriteString(strconv.FormatFloat(md.MubiRating, 'f', 1, 32))
	sb.WriteString(" (")
	sb.WriteString(md.MubiRatingsNumber)
	sb.WriteString(")")
	return sb.String()
}

type imdb struct{}

func (i imdb) Header() string { return "IMDB" }
func (i imdb) Value(md *movie.Data) interface{} {
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

type mins struct{}

func (m mins) Header() string                   { return "Mins" }
func (m mins) Value(md *movie.Data) interface{} { return md.Mins }

type year struct{}

func (y year) Header() string                   { return "Year" }
func (y year) Value(md *movie.Data) interface{} { return md.Year }

type country struct{}

func (c country) Header() string                   { return "Country" }
func (c country) Value(md *movie.Data) interface{} { return md.Country }

type genre struct{}

func (g genre) Header() string                   { return "Genre" }
func (g genre) Value(md *movie.Data) interface{} { return md.Genre }
