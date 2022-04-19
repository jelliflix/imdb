package torrent

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

var trackersTPB = []string{
	"udp://tracker.coppersurfer.tk:6969/announce",
	"udp://9.rarbg.to:2920/announce",
	"udp://tracker.opentrackr.org:1337",
	"udp://tracker.internetwarriors.net:1337/announce",
	"udp://tracker.leechers-paradise.org:6969/announce",
	"udp://tracker.coppersurfer.tk:6969/announce",
	"udp://tracker.pirateparty.gr:6969/announce",
	"udp://tracker.cyberia.is:6969/announce",
}

type TPBOptions struct {
	BaseURL        string
	SocksProxyAddr string
	Timeout        time.Duration
	CacheAge       time.Duration
}

var DefaultTPBOpts = TPBOptions{
	BaseURL:  "https://apibay.org",
	Timeout:  5 * time.Second,
	CacheAge: 24 * time.Hour,
}

var _ MagnetFinder = (*tpb)(nil)

type tpb struct {
	baseURL    string
	httpClient *http.Client
	cache      Cache
	cacheAge   time.Duration
	metaGetter MetaGetter
	logger     *zap.Logger
}

func NewTPB(opts TPBOptions, cache Cache, metaGetter MetaGetter, logger *zap.Logger) *tpb {
	return &tpb{
		baseURL: opts.BaseURL,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		cache:      cache,
		cacheAge:   opts.CacheAge,
		metaGetter: metaGetter,
		logger:     logger,
	}
}

func (c *tpb) FindMovie(ctx context.Context, imdbID string) ([]Result, error) {
	meta, err := c.metaGetter.GetMovie(ctx, imdbID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get movie title via Cinemeta for IMDb ID %v: %v", imdbID, err)
	}
	escapedQuery := imdbID
	return c.find(ctx, imdbID, meta.Title, escapedQuery, false)
}

func (c *tpb) FindEpisode(ctx context.Context, imdbID string, season, episode int) ([]Result, error) {
	id := imdbID + ":" + strconv.Itoa(season) + ":" + strconv.Itoa(episode)
	meta, err := c.metaGetter.GetEpisode(ctx, imdbID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get TV show title via Cinemeta for ID %v: %v", id, err)
	}
	query, err := createSeriesSearch(ctx, c.metaGetter, imdbID, season, episode)
	if err != nil {
		return nil, err
	}
	queryEscaped := url.QueryEscape(query)
	queryEscaped += "&cat=208"
	return c.find(ctx, id, meta.Title, queryEscaped, true)
}

func (c *tpb) find(ctx context.Context, id, title, escapedQuery string, fuzzy bool) ([]Result, error) {
	cacheKey := id + "-TPB"
	torrentList, created, found, err := c.cache.Get(cacheKey)
	if found && time.Since(created) <= (c.cacheAge) {
		return torrentList, nil
	}

	reqUrl := c.baseURL + "/q.php?q=" + escapedQuery
	res, err := c.httpClient.Get(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("couldn't GET %v: %v", reqUrl, err)
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

	torrents := gjson.ParseBytes(resBody).Array()
	if len(torrents) == 0 {
		return nil, nil
	}

	var results []Result
	for _, torrent := range torrents {
		torrentName := torrent.Get("name").String()
		quality := ""
		if strings.Contains(torrentName, "720p") {
			quality = "720p"
		} else if strings.Contains(torrentName, "1080p") {
			quality = "1080p"
		} else if strings.Contains(torrentName, "2160p") {
			quality = "2160p"
		} else {
			continue
		}
		if strings.Contains(torrentName, "10bit") {
			quality += " 10bit"
		}
		if strings.Contains(torrentName, "HDCAM") {
			quality += " (⚠️cam)"
		} else if strings.Contains(torrentName, "HDTS") || strings.Contains(torrentName, "HD-TS") {
			quality += " (⚠️telesync)"
		}
		infoHash := torrent.Get("info_hash").String()
		if infoHash == "" {
			continue
		} else if len(infoHash) != 40 {
			continue
		}
		infoHash = strings.ToLower(infoHash)
		magnetURL := createMagnetURL(ctx, infoHash, title, trackersTPB)
		size := int(torrent.Get("size").Int())
		seeders := int(torrent.Get("seeders").Int())
		result := Result{
			Name:      torrentName,
			Title:     title,
			Quality:   quality,
			InfoHash:  infoHash,
			MagnetURL: magnetURL,
			Fuzzy:     fuzzy,
			Size:      size,
			Seeders:   seeders,
		}
		results = append(results, result)
	}

	if err := c.cache.Set(cacheKey, results); err != nil {
		c.logger.Error("couldn't cache torrents", zap.Error(err), zap.String("cache", "torrent"))
	}

	return results, nil
}
