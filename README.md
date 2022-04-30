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
ExportWatchList() ([]Item, error)
```

ExportWatchList returns all watchlist movies or tv episodes, order not guaranteed.

##### Examples

```go
import "github.com/jelliflix/imdb"

client, err := imdb.NewIMDB(imdb.DefaultOptions, "ur152083192")
if err != nil {
    log.Fatal(err)
}

watchList, err := client.ExportWatchList()
if err != nil {
    log.Fatal(err)
}

log.Println(watchList)
// Output:
// [{tt9170516 Skyggen i mit øje movie} {tt11650328 Severance: Good News About Hell tvEpisode}]
```
