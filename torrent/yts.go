package torrent

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

var trackersYTS = []string{
	"udp://open.demonii.com:1337/announce",
	"udp://tracker.openbittorrent.com:80",
	"udp://tracker.coppersurfer.tk:6969",
	"udp://glotorrents.pw:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://torrent.gresille.org:80/announce",
	"udp://p4p.arenabg.com:1337",
	"udp://tracker.leechers-paradise.org:6969",
}

type YTSOptions struct {
	BaseURL  string
	Timeout  time.Duration
	CacheAge time.Duration
}

var DefaultYTSOpts = YTSOptions{
	BaseURL:  "https://yts.mx",
	Timeout:  5 * time.Second,
	CacheAge: 24 * time.Hour,
}

var _ MagnetFinder = (*yts)(nil)

type yts struct {
	baseURL    string
	httpClient *http.Client
	cache      Cache
	cacheAge   time.Duration
	logger     *zap.Logger
}

func NewYTS(opts YTSOptions, cache Cache, logger *zap.Logger) *yts {
	return &yts{
		baseURL: opts.BaseURL,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		cache:    cache,
		cacheAge: opts.CacheAge,
		logger:   logger,
	}
}

func (c *yts) FindMovie(ctx context.Context, imdbID string) ([]Result, error) {
	cacheKey := imdbID + "-YTS"
	torrentList, created, found, err := c.cache.Get(cacheKey)
	if err != nil {
		c.logger.Error("couldn't get torrent results from cache", zap.Error(err))
	}

	if found && time.Since(created) <= (c.cacheAge) {
		return torrentList, nil
	}

	url := c.baseURL + "/api/v2/list_movies.json?query_term=" + imdbID
	res, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't GET %v: %v", url, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad GET response: %v", res.StatusCode)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't read response body: %v", err)
	}

	torrents := gjson.GetBytes(resBody, "data.movies.0.torrents").Array()
	if len(torrents) == 0 {
		return nil, nil
	}
	title := gjson.GetBytes(resBody, "data.movies.0.title").String()
	var results []Result
	for _, torrent := range torrents {
		quality := torrent.Get("quality").String()
		if quality == "720p" || quality == "1080p" || quality == "2160p" {
			infoHash := torrent.Get("hash").String()
			if infoHash == "" {
				continue
			} else if len(infoHash) != 40 {
				continue
			}
			infoHash = strings.ToLower(infoHash)
			magnetURL := createMagnetURL(ctx, infoHash, title, trackersYTS)
			ripType := torrent.Get("type").String()
			if ripType != "" {
				quality += " (" + ripType + ")"
			}
			size := int(torrent.Get("size_bytes").Int())
			seeders := int(torrent.Get("seeds").Int())

			result := Result{
				Name:      title + " [" + quality + "] [YTS]",
				Title:     title,
				Quality:   quality,
				InfoHash:  infoHash,
				MagnetURL: magnetURL,
				Size:      size,
				Seeders:   seeders,
			}
			results = append(results, result)
		}
	}

	if err := c.cache.Set(cacheKey, results); err != nil {
		c.logger.Error("couldn't cache torrents", zap.Error(err), zap.String("cache", "torrent"))
	}

	return results, nil
}

func (c *yts) FindEpisode(_ context.Context, _ string, _, _ int) ([]Result, error) {
	return nil, nil
}
