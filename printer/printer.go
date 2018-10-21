package printer

import (
	"github.com/fatih/color"
	"github.com/llugin/mubi-parser/movie"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
)

// PrintTable pretty-prints collected data as a table
func PrintTable(movies []movie.Data, noColor bool, maxLen int) {
	color.NoColor = noColor
	columnsNo := 8

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 4, ' ', 0)
	colors := []*color.Color{color.New(color.FgWhite), color.New(color.FgGreen)}

	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))
	colors[0].Fprintln(w, "Days\tTitle\tDirector\tMUBI\tIMDB\tMins\tYear\tCountry\tGenre")
	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))

	var c *color.Color
	var sb strings.Builder
	sb.WriteString(strings.Repeat("%v\t", columnsNo))
	sb.WriteString("%v\n")
	for i, m := range movies {
		mc := truncate(m, maxLen)
		c = colors[i%2]
		c.Fprintf(w, sb.String(),
			mc.DaysToWatch, mc.Title, mc.Director, mubiRatingRepr(&mc),
			imdbRatingRepr(&mc), mc.Mins, mc.Year, mc.Country, mc.Genre)
	}
	colors[0].Fprintln(w, strings.Repeat("\t", columnsNo))

	w.Flush()
}

func truncate(m movie.Data, maxLen int) movie.Data {
	if maxLen <= 0 {
		return m
	}

	truncator := func(s string) string {
		if utf8.RuneCountInString(s) <= maxLen {
			return s
		}
		runes := 0
		out := s
		for i := range s {
			if runes >= maxLen {
				out = s[:i]
				break
			}
			runes++
		}
		return out
	}

	trunc := m
	r := reflect.ValueOf(&trunc).Elem()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		if f.Kind() != reflect.String {
			continue
		}
		f.SetString(truncator(f.String()))
	}

	return trunc
}

func mubiRatingRepr(d *movie.Data) string {
	var sb strings.Builder
	sb.WriteString(strconv.FormatFloat(d.MubiRating, 'f', 1, 32))
	sb.WriteString(" (")
	sb.WriteString(d.MubiRatingsNumber)
	sb.WriteString(")")
	return sb.String()
}

func imdbRatingRepr(d *movie.Data) string {
	if d.ImdbRating == 0.0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(strconv.FormatFloat(d.ImdbRating, 'f', 1, 32))
	sb.WriteString(" (")
	sb.WriteString(d.ImdbRatingsNumber)
	sb.WriteString(")")
	return sb.String()
}
