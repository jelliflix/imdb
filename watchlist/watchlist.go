package watchlist

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type IMDB struct {
	user string
	opts Options
}

type Options struct {
	URL string

	Timeout time.Duration
}

func parseUID(id string) string {
	if strings.Contains(id, "ur") {
		return id
	}

	return fmt.Sprintf("ur%v", id)
}

func uniqueList(in []string) (l []string) {
	t := map[string]bool{}
	for _, s := range in {
		t[s] = true
	}

	for s := range t {
		l = append(l, s)
	}

	return
}

func NewIMDB(opts Options, user string) *IMDB {
	return &IMDB{opts: opts, user: parseUID(user)}
}

var DefaultOptions = Options{
	Timeout: 10 * time.Second,
	URL:     "https://www.imdb.com/",
}

func (i *IMDB) request(endpoint string, params url.Values) (body []byte, err error) {
	URL, err := url.Parse(i.opts.URL)
	if err != nil {
		return
	}

	URL.Path += fmt.Sprintf("user/%s", i.user)
	URL.Path += endpoint
	URL.RawQuery = params.Encode()

	c := &http.Client{Timeout: i.opts.Timeout}
	resp, err := c.Get(URL.String())
	if err != nil {
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("got http error %q", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func (i *IMDB) reqWatchlist(kind, sort string, page int) (list []string, err error) {
	params := url.Values{}
	params.Add("sort", sort)
	params.Add("title_type", kind)
	params.Add("page", strconv.Itoa(page))

	resp, err := i.request("/watchlist", params)
	if err != nil {
		return
	}

	return regexp.MustCompile(`tt\d{7}`).FindAllString(string(resp), -1), err
}

func (i *IMDB) watchlist(kind, sort string) (watchlist []string, err error) {
	page, prevCount := 1, 0

	for {
		curMatches, err := i.reqWatchlist(kind, sort, page)
		if err != nil {
			return watchlist, err
		}

		curMatches = uniqueList(curMatches)
		watchlist = append(watchlist, curMatches...)
		curCount := len(curMatches)
		if curCount == prevCount {
			break
		}

		prevCount = curCount
		page++
	}

	watchlist = uniqueList(watchlist)

	return
}

func (i *IMDB) GetMovies() ([]string, error) {
	return i.watchlist("movie", "date_added,desc")
}

func (i *IMDB) GetSeries() ([]string, error) {
	return i.watchlist("tvSeries", "date_added,desc")
}
