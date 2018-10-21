package printer

import (
	"github.com/fatih/color"
	"github.com/llugin/mubi-parser/movie"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
)

var columns = []columnRepr{daysRepr{}, titleRepr{}, directorRepr{},
	mubiRepr{}, imdbRepr{}, minsRepr{}, yearRepr{}, countryRepr{}, genreRepr{}}

// PrintTable pretty-prints collected data as a table
func PrintTable(movies []movie.Data, noColor bool, maxLen int) {
	color.NoColor = noColor
	tabsNo := len(columns) - 1

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 4, ' ', 0)
	colors := []*color.Color{color.New(color.FgWhite), color.New(color.FgGreen)}

	colors[0].Fprintln(w, strings.Repeat("\t", tabsNo))
	colors[0].Fprintln(w, strings.Join(getHeaders(), "\t"))
	colors[0].Fprintln(w, strings.Repeat("\t", tabsNo))

	var c *color.Color
	var sb strings.Builder
	sb.WriteString(strings.Repeat("%v\t", tabsNo))
	sb.WriteString("%v\n")
	for i, m := range movies {
		mc := truncate(m, maxLen)
		c = colors[i%2]
		c.Fprintf(w, sb.String(), getValues(&mc)...)
	}
	colors[0].Fprintln(w, strings.Repeat("\t", tabsNo))

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

func getValues(md *movie.Data) []interface{} {
	values := []interface{}{}
	for _, c := range columns {
		values = append(values, c.Value(md))
	}
	return values
}

func getHeaders() []string {
	headers := []string{}
	for _, c := range columns {
		headers = append(headers, c.Header())
	}
	return headers
}
