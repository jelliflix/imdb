package torrent

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

var magnet2InfoHashRegex = regexp.MustCompile(`btih:.+?&`)

type findFunc func(context.Context, MagnetFinder) ([]Result, error)

type MetaGetter interface {
	GetMovie(ctx context.Context, imdbID string) (Meta, error)
	GetEpisode(ctx context.Context, imdbID string) (Meta, error)
}

type MagnetFinder interface {
	FindMovie(ctx context.Context, imdbID string) ([]Result, error)
	FindEpisode(ctx context.Context, imdbID string, season, episode int) ([]Result, error)
}

type Meta struct {
	Episode int
	Season  int
	Year    int

	Title string
}

type Torrent struct {
	logger  *zap.Logger
	timeout time.Duration
	clients []MagnetFinder
}

func NewTorrent(clients []MagnetFinder, timeout time.Duration, logger *zap.Logger) *Torrent {
	return &Torrent{
		clients: clients,
		timeout: timeout,
		logger:  logger,
	}
}

func (t *Torrent) FindMovie(ctx context.Context, imdbID string) ([]Result, error) {
	find := func(ctx context.Context, siteClient MagnetFinder) ([]Result, error) {
		return siteClient.FindMovie(ctx, imdbID)
	}
	return t.find(ctx, find)
}

func (t *Torrent) FindEpisode(ctx context.Context, imdbID string, season, episode int) ([]Result, error) {
	find := func(ctx context.Context, siteClient MagnetFinder) ([]Result, error) {
		return siteClient.FindEpisode(ctx, imdbID, season, episode)
	}
	return t.find(ctx, find)
}

func (t *Torrent) find(ctx context.Context, find findFunc) ([]Result, error) {
	clients := len(t.clients)
	errChan := make(chan error, clients)
	resChan := make(chan []Result, clients)

	for _, client := range t.clients {
		timer := time.NewTimer(t.timeout)
		go func(finder MagnetFinder, timer *time.Timer) {
			defer timer.Stop()

			siteResChan := make(chan []Result)
			siteErrChan := make(chan error)
			go func() {
				results, err := find(ctx, finder)
				if err != nil {
					siteErrChan <- err
				} else {
					siteResChan <- results
				}
			}()
			select {
			case res := <-siteResChan:
				resChan <- res
			case err := <-siteErrChan:
				errChan <- err
			case <-timer.C:
				resChan <- nil
			}
		}(client, timer)
	}

	var combinedResults []Result
	var errs []error
	dupRemovalRequired := false
	for i := 0; i < clients; i++ {
		select {
		case results := <-resChan:
			if !dupRemovalRequired && len(combinedResults) > 0 && len(results) > 0 {
				dupRemovalRequired = true
			}
			combinedResults = append(combinedResults, results...)
		case err := <-errChan:
			errs = append(errs, err)
		}
	}

	returnErrors := len(errs) == clients

	if returnErrors {
		errsMsg := "couldn't find torrents on any site: "
		for i := 1; i <= clients; i++ {
			errsMsg += fmt.Sprintf("%v.: %v; ", i, errs[i-1])
		}
		errsMsg = strings.TrimSuffix(errsMsg, "; ")
		return nil, fmt.Errorf(errsMsg)
	}

	var noDupResults []Result
	if dupRemovalRequired {
		infoHashes := map[string]struct{}{}
		for _, result := range combinedResults {
			if _, ok := infoHashes[result.InfoHash]; !ok {
				noDupResults = append(noDupResults, result)
				infoHashes[result.InfoHash] = struct{}{}
			}
		}
	} else {
		noDupResults = combinedResults
	}

	return noDupResults, nil
}

type Result struct {
	Name      string
	Title     string
	Quality   string
	InfoHash  string
	MagnetURL string

	Seeders int
	Fuzzy   bool
	Size    int
}

func createMagnetURL(_ context.Context, infoHash, title string, trackers []string) string {
	magnetURL := "magnet:?xt=urn:btih:" + infoHash + "&dn=" + url.QueryEscape(title)
	for _, tracker := range trackers {
		magnetURL += "&tr" + tracker
	}
	return magnetURL
}

func createSeriesSearch(ctx context.Context, metaGetter MetaGetter, imdbID string, season, episode int) (string, error) {
	id := imdbID + ":" + strconv.Itoa(season) + ":" + strconv.Itoa(episode)
	meta, err := metaGetter.GetEpisode(ctx, imdbID)
	if err != nil {
		return "", fmt.Errorf("couldn't get TV show title via Cinemeta for ID %v: %v", id, err)
	}
	seasonString := strconv.Itoa(season)
	episodeString := strconv.Itoa(episode)
	if season < 10 {
		seasonString = "0" + seasonString
	}
	if episode < 10 {
		episodeString = "0" + episodeString
	}
	return fmt.Sprintf("%v S%vE%v", meta.Title, seasonString, episodeString), nil
}
