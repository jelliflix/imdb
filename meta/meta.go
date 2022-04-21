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
	SeriesID string
	Episode  int
	Season   int
	Year     int

	Title string
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
		SeriesID string `json:"seriesID,required"`
		Episode  string `json:"Episode,required"`
		Season   string `json:"Season,required"`
		Year     string `json:"Year,required"`

		Title string `json:"Title,required"`
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	episode, _ := strconv.ParseInt(v.Episode, 0, 64)
	season, _ := strconv.ParseInt(v.Season, 0, 64)
	year, _ := strconv.ParseInt(strings.ReplaceAll(strings.Split(v.Year, "–")[0], "–", ""), 0, 64)

	// Sometimes api only contains one `t` on series id.
	series := []rune(v.SeriesID)
	if len(series) > 0 && string(series[1]) != "t" {
		m.SeriesID = "t" + v.SeriesID
	} else {
		m.SeriesID = v.SeriesID
	}

	m.Episode = int(episode)
	m.Season = int(season)
	m.Year = int(year)

	m.Title = v.Title

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

func (o *OMDB) GetMovie(_ context.Context, id string) (Meta, error) {
	meta, err := o.reqMeta("movie", id)
	return meta, err
}

func (o *OMDB) GetEpisode(_ context.Context, id string) (Meta, error) {
	meta, err := o.reqMeta("episode", id)
	return meta, err
}

func (o *OMDB) GetSeriesByEpisode(ctx context.Context, id string) (Meta, error) {
	episode, err := o.GetEpisode(ctx, id)
	meta, err := o.reqMeta("series", episode.SeriesID)
	return meta, err
}
