package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jelliflix/imdb/torrent"
)

type OMDB struct {
	apiKey string
	opts   Options
}

type Options struct {
	URL string

	Timeout time.Duration
}

type Meta struct {
	Title string
	Year  int
}

func NewOMDB(opts Options, apiKey string) *OMDB {
	return &OMDB{opts: opts, apiKey: apiKey}
}

var DefaultOptions = Options{
	Timeout: 10 * time.Second,
	URL:     "https://www.omdbapi.com/",
}

func (m *Meta) UnmarshalJSON(data []byte) error {
	var v struct {
		Title string `json:"Title,required"`
		Year  string `json:"Year,required"`
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	year, _ := strconv.ParseInt(strings.ReplaceAll(strings.Split(v.Year, "–")[0], "–", ""), 0, 64)

	m.Title = v.Title
	m.Year = int(year)

	return nil
}

func (o *OMDB) request(params url.Values) (reader io.ReadCloser, err error) {
	URL, err := url.Parse(o.opts.URL)
	if err != nil {
		return
	}

	URL.RawQuery = params.Encode()

	c := &http.Client{Timeout: o.opts.Timeout}
	resp, err := c.Get(URL.String())
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return reader, fmt.Errorf("got http error %q", resp.Status)
	}

	return resp.Body, err
}

func (o *OMDB) reqMeta(kind, id string) (meta Meta, err error) {
	params := url.Values{}
	params.Add("i", id)
	params.Add("type", kind)
	params.Add("apikey", o.apiKey)

	resp, err := o.request(params)
	if err != nil {
		return
	}

	defer func() {
		_ = resp.Close()
	}()

	dec := json.NewDecoder(resp)
	err = dec.Decode(&meta)
	if err != nil {
		return
	}

	return
}

func (o *OMDB) GetMovie(_ context.Context, id string) (torrent.Meta, error) {
	meta, err := o.reqMeta("movie", id)
	return torrent.Meta(meta), err
}

func (o *OMDB) GetSeries(_ context.Context, id string) (torrent.Meta, error) {
	meta, err := o.reqMeta("series", id)
	return torrent.Meta(meta), err
}
