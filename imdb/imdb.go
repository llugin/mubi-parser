package imdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/llugin/mubi-parser/debugging"
	"github.com/llugin/mubi-parser/movie"
	"github.com/llugin/mubi-parser/mubi"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	urlFormat = "http://www.omdbapi.com/?t=%s&y=%v&type=movie&apikey=%s"
)

var (
	//APICount - OMDB API call counter
	APICount int
	//APIKey - OMDB api key
	APIKey string
	//Sleep - sleep time between API calls in miliseconds
	Sleep = 200
)

type apiResp struct {
	ImdbRating string `json:"imdbRating"`
	ImdbVotes  string `json:"imdbVotes"`
	Response   string `json:"Response"`
	Director   string `json:"Director"`
	Error      string `json:"Error"`
}

//SendRatings get movie ratings of imdb movies
func SendRatings(done <-chan struct{}, in <-chan movie.Data) <-chan movie.Data {
	out := make(chan movie.Data, mubi.MaxMovies)

	go func() {
		defer close(out)
		for m := range in {
			if APIKey != "" {
				time.Sleep(time.Duration(Sleep) * time.Millisecond)
				obtainMovieRating(&m)
			} else {
				debugging.Log().Println("no OMDB Api Key")
			}
			select {
			case out <- m:
			case <-done:
				return
			}
		}
	}()

	return out
}

func obtainMovieRating(m *movie.Data) {
	var ar apiResp
	var err error

	if ar, err = getAPIResp(m.Title, m.Director, m.Year); err == nil {
		goto Found
	} else {
		debugging.Log().Println(err)
	}
	// Try alternative title
	if m.AltTitle != "" {
		if ar, err = getAPIResp(m.AltTitle, m.Director, m.Year); err == nil {
			goto Found
		} else {
			debugging.Log().Println(err)
		}
	}

	// Try with approximate years (+1/-1 year)
	if ar, err = getAPIResp(m.Title, m.Director, m.Year-1); err == nil {
		goto Found
	} else {
		debugging.Log().Println(err)
	}

	if ar, err = getAPIResp(m.Title, m.Director, m.Year+1); err == nil {
		goto Found
	} else {
		debugging.Log().Println(err)
	}

	// Try with normalized director name
	if ar, err = getAPIResp(m.Title, normalizeName(m.Director), m.Year); err == nil {
		goto Found
	} else {
		debugging.Log().Println(err)
	}

Found:
	if f, err := strconv.ParseFloat(ar.ImdbRating, 32); err == nil {
		m.ImdbRating = f
	} else {
		debugging.Log().Printf("Could not parse imdb rating '%s' as a float\n", ar.ImdbRating)
		m.ImdbRating = 0.0
	}

	m.ImdbRatingsNumber = ar.ImdbVotes
}

func getAPIResp(title, director string, year int) (apiResp, error) {
	APICount++
	var ar apiResp
	var err error

	url := fmt.Sprintf(urlFormat, strings.Replace(title, " ", "+", -1), year, APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return ar, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ar, err
	}
	err = json.Unmarshal(body, &ar)
	if err != nil && ar.Response != "True" {
		err = fmt.Errorf(ar.Error)
	}
	if ar.Director != director {
		err = fmt.Errorf("Wrong director")
	}
	return ar, err
}

func normalizeName(in string) string {
	isMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}

	in = strings.Replace(in, "ł", "l", -1)
	in = strings.Replace(in, "Ł", "L", -1)
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, in)
	return result
}
