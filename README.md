IMDB
====

Torrent, meta and watchlist provider for movies and TV series by IMDB.

## Usage

### Installation

#### `go get`

```shell
$ go get -u -v github.com/jelliflix/imdb
```

#### `go mod` (Recommended)

```go
import "github.com/jelliflix/imdb"
```

```shell
$ go mod tidy
```

### API

#### Watchlist

```go
GetMovies() ([]string, error)
GetEpisodes() ([]string, error)
```

GetX returns all watchlist movies or tv episodes, order guaranteed.

##### Examples

```go
import wl "github.com/jelliflix/imdb/watchlist"

imdb := wl.NewIMDB(wl.DefaultOptions, "ur152083192")
movies, _ := imdb.GetMovies()
episodes, _ := imdb.GetEpisodes()

log.Println(movies, episodes)
// Output:
// [tt9170516] [tt12076928]
```

#### Meta getter

```go
GetMovie(ctx context.Context, imdbID string) (Meta, error)
GetEpisode(ctx context.Context, imdbID string) (Meta, error)
```

GetX returns meta for movie or tv episodes.

##### Examples

```go
import mg "github.com/jelliflix/imdb/meta"

omdb := mg.NewOMDB(mg.DefaultOptions, "xxxxxxxx")
meta, _ := omdb.GetEpisode(context.Background(), "tt12076928")

log.Println(meta)
// Output:
// {4 2 2021 Pathfinder}
```

#### Magnet finder

```go
FindMovie(ctx context.Context, imdbID string) ([]Result, error)
FindEpisode(ctx context.Context, imdbID string, season, episode int) ([]Result, error)
```

GetX returns magnet links for movie or tv episodes.

##### Examples

```go
import (
    "context"
    "fmt"
    "time"
    
    mg "github.com/jelliflix/imdb/meta"
    "github.com/jelliflix/imdb/torrent"
    "go.uber.org/zap"
)

logger := zap.NewNop()
timeout := time.Second * 10
cache := torrent.NewInMemCache()
meta := mg.NewOMDB(mg.DefaultOptions, "xxxxxxxx")

yts := torrent.NewYTS(torrent.DefaultYTSOpts, cache, logger)
tpb := torrent.NewTPB(torrent.DefaultTPBOpts, cache, meta, logger)
rarbg := torrent.NewRARBG(torrent.DefaultRARBOpts, cache, logger)

client := torrent.NewTorrent([]torrent.MagnetFinder{yts, tpb, rarbg}, timeout, logger)

torrents, err := client.FindEpisode(context.Background(), "tt12076928", 2, 4)
if err != nil {
    panic(err)
}

for _, t := range torrents {
    fmt.Printf("found torrent: %v [%v - %v Bytes] \n", t.Title, t.Quality, t.Size)
    fmt.Printf("%s\n\n", t.MagnetURL)
}
// Output:
// found torrent: Pathfinder [1080p - 4983076697 Bytes] 
// magnet:?xt=urn:btih:011eac...
// found torrent: Pathfinder [720p - 1657418601 Bytes]
// magnet:?xt=urn:btih:29b3ea...
```

