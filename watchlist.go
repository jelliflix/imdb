package imdb

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type IMDB struct {
	id   string
	opts Options
}

type Options struct {
	URL string

	Timeout time.Duration
}

func NewIMDB(opts Options, id string) (*IMDB, error) {
	imdb := &IMDB{opts: opts}
	err := imdb.setID(id)

	return imdb, err
}

var DefaultOptions = Options{
	Timeout: 10 * time.Second,
	URL:     "https://www.imdb.com/",
}

func (i *IMDB) request(endpoint string) (body []byte, err error) {
	URL, err := url.Parse(i.opts.URL)
	if err != nil {
		return
	}

	URL.Path += endpoint

	c := &http.Client{Timeout: i.opts.Timeout}
	resp, err := c.Get(URL.String())
	if err != nil {
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return body, fmt.Errorf("not found")
	} else if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("got http error %q", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func (i *IMDB) setID(id string) (err error) {
	i.id = id

	switch {
	case strings.Contains(id, "ls"):
		i.id = id
	case strings.Contains(id, "ur"):
		ids, err := i.extractWatchList()
		if len(ids) > 0 {
			i.id = ids[0]
		} else {
			return fmt.Errorf("user not found %s", id)
		}

		return err
	case id == "":
		return fmt.Errorf("missing watchlist/user id")
	default:
		i.id = fmt.Sprintf("ls%v", id)
	}

	return
}

func (i *IMDB) extractWatchList() (ids []string, err error) {
	resp, err := i.request(fmt.Sprintf("user/%s/watchlist", i.id))
	if err != nil {
		return
	}

	return regexp.MustCompile(`ls\d{9}`).FindAllString(string(resp), -1), err
}

func (i *IMDB) ExportWatchList() (watchlist []Item, err error) {
	resp, err := i.request(fmt.Sprintf("list/%s/export", i.id))
	if err != nil {
		return
	}

	r := csv.NewReader(bytes.NewReader(resp))
	data, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	return parseWatchList(data), nil
}

type Item struct {
	ID   string
	Name string
	Type string
}

func parseWatchList(data [][]string) (watchList []Item) {
	for i, line := range data {
		if i > 0 {
			var rec Item
			for j, field := range line {
				if j == 1 {
					rec.ID = field
				} else if j == 5 {
					rec.Name = field
				} else if j == 7 {
					rec.Type = field
				}
			}
			watchList = append(watchList, rec)
		}
	}

	return
}
