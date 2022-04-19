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

