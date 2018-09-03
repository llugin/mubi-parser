package imdb

import (
	"encoding/json"
	"fmt"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/mubi"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	keyFile   = "omdb_apikey"
	urlFormat = "http://www.omdbapi.com/?t=%s&y=%v&type=movie&apikey=%s"
)

var (
	//APICount - OMDB API call counter
	APICount    int
	keyFilePath string
	key         string
	keyError    error
)

type apiResp struct {
	ImdbRating string `json:"imdbRating"`
	ImdbVotes  string `json:"imdbVotes"`
	Response   string `json:"Response"`
	Director   string `json:"Director"`
	Error      string `json:"Error"`
}

func init() {
	log.SetFlags(log.Lshortfile)
	key, keyError = getKey()
	if keyError != nil {
		log.Printf("%v. Could not get OMDB API key. Skipping IMDB data acquirement.", keyError)
	}
}

//SendRatings get movie ratings of imdb movies
func SendRatings(done <-chan struct{}, in <-chan movie.Data) <-chan movie.Data {
	out := make(chan movie.Data, mubi.MaxMovies)

	go func() {
		defer close(out)
		for m := range in {
			if keyError == nil {
				time.Sleep(time.Millisecond * 200)
				obtainMovieRating(&m)
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
	}
	// Try alternative title
	if m.AltTitle != "" {
		if ar, err = getAPIResp(m.AltTitle, m.Director, m.Year); err == nil {
			goto Found
		}
	}
	// Try with approximate years (+1/-1 year)
	if ar, err = getAPIResp(m.Title, m.Director, m.Year-1); err == nil {
		goto Found
	}
	if ar, err = getAPIResp(m.Title, m.Director, m.Year+1); err == nil {
		goto Found
	}
	// Try with normalized director name
	if ar, err = getAPIResp(m.Title, normalizeName(m.Director), m.Year); err == nil {
		goto Found
	}

Found:
	if f, err := strconv.ParseFloat(ar.ImdbRating, 32); err == nil {
		m.ImdbRating = f
	} else {
		m.ImdbRating = 0.0
	}

	m.ImdbRatingsNumber = ar.ImdbVotes
}

func getAPIResp(title, director string, year int) (apiResp, error) {
	APICount++
	var ar apiResp
	var err error

	url := fmt.Sprintf(urlFormat, strings.Replace(title, " ", "+", -1), year, key)
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

func getKey() (string, error) {
	var key string
	ex, err := os.Executable()
	if err != nil {
		return key, err
	}
	keyFilePath = filepath.Join(filepath.Dir(ex), keyFile)

	out, err := ioutil.ReadFile(keyFilePath)
	if err != nil {
		return key, err
	}
	return strings.TrimSpace(string(out)), nil
}

func normalizeName(in string) string {
	in = strings.Replace(in, "ł", "l", -1)
	in = strings.Replace(in, "Ł", "L", -1)
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, in)
	return result
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
