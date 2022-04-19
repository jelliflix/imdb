IMDB
====

IMDB provides torrents for movies and TV shows by IMDB id.

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
GetSeries() ([]string, error)
```

GetX returns all watchlist movies or tv series, order guaranteed.

##### Examples

```go
import wl "github.com/jelliflix/imdb/watchlist"

imdb := wl.NewIMDB(wl.DefaultOptions, "ur012345678")
movies, _ := imdb.GetMovies()
series, _ := imdb.GetSeries()

log.Println(movies, series)
// Output:
// [tt9170516] [tt7772588]
```

#### Meta getter

```go
GetMovie(ctx context.Context, imdbID string) (Meta, error)
GetSeries(ctx context.Context, imdbID string) (Meta, error)
```

GetX returns meta for movie or tv series.

##### Examples

```go
import mg "github.com/jelliflix/imdb/meta"

omdb := mg.NewOMDB(mg.DefaultOptions, "xxxxxxxx")
meta, _ := omdb.GetSeries(context.Background(), "tt7772588")

log.Println(meta)
// Output:
// {For All Mankind 2019}
```

