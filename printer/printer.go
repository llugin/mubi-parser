package printer

import (
	"github.com/fatih/color"
	"github.com/llugin/mubi-parser/movie"
	"os"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
)

var (
	tabsNo     = len(columns) - 1
	emptyRow   = strings.Repeat("\t", tabsNo)
	valuesRow  = strings.Repeat("%v\t", tabsNo) + "%v\n"
	headersRow = strings.Join(getHeaders(), "\t")
	colors     = []*color.Color{color.New(color.FgWhite), color.New(color.FgGreen)}
)

// PrintTable pretty-prints collected data as a table
func PrintTable(movies []movie.Data, noColor bool, maxLen int) {
	color.NoColor = noColor

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 4, ' ', 0)

	colors[0].Fprintln(w, emptyRow)
	colors[0].Fprintln(w, headersRow)
	colors[0].Fprintln(w, emptyRow)

	for i, m := range movies {
		colors[i%2].Fprintf(w, valuesRow, getValues(&m, maxLen)...)
	}
	colors[0].Fprintln(w, emptyRow)

	w.Flush()
}

func getHeaders() []string {
	headers := []string{}
	for _, c := range columns {
		headers = append(headers, c.Header())
	}
	return headers
}

func getValues(md *movie.Data, maxLen int) []interface{} {
	values := []interface{}{}
	for _, c := range columns {
		values = append(values, truncate(c.Value(md), maxLen))
	}
	return values
}

func truncate(value interface{}, maxLen int) interface{} {
	if maxLen <= 0 {
		return value
	}

	if s, ok := value.(string); ok {
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

	return value
}
