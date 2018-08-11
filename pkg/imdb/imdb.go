package imdb

import (
	"encoding/json"
	"fmt"
	"github.com/llugin/mubi-parser/pkg/movie"
	"github.com/llugin/mubi-parser/pkg/mubi"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	keyFile   = "omdb_api"
	urlFormat = "http://www.omdbapi.com/?t=%s&y=%s&apikey=%s"
)

var (
	keyFilePath string
	key         string
)

type apiResp struct {
	ImdbRating string `json:"imdbRating"`
	ImdbVotes  string `json:"imdbVotes"`
	Response   string `json:"Response"`
	Error      string `json:"Error"`
}

func init() {
	var err error
	key, err = getKey()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

//GetRatings get movie ratings of imdb movies
func GetRatings(in <-chan movie.Data) <-chan movie.Data {
	out := make(chan movie.Data, mubi.MaxMovies)
	go func() {
		for m := range in {
			time.Sleep(time.Millisecond * 200)
			obtainMovieRating(&m)
			out <- m
		}
		close(out)
	}()
	return out
}

func obtainMovieRating(m *movie.Data) {
	ar, err := getAPIResp(m)
	if err != nil || ar.Response != "True" {
		return
	}

	m.ImdbRating = ar.ImdbRating
	m.ImdbRatingsNumber = ar.ImdbVotes
}

func getAPIResp(m *movie.Data) (apiResp, error) {
	var ar apiResp

	url := fmt.Sprintf(urlFormat, strings.Replace(m.Title, " ", "+", -1), m.Year, key)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return ar, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return ar, err
	}
	err = json.Unmarshal(body, &ar)
	if err != nil {
		log.Println(err)
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
