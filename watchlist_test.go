package imdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifyUser(t *testing.T) {
	tests := map[string]bool{
		"560188055":    true,
		"ur152083192":  true,
		"ls560188055":  true,
		"ur152083192x": false,
		"":             false,
	}

	for id, mustFound := range tests {
		_, err := NewIMDB(DefaultOptions, id)
		if mustFound {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestExportWatchList(t *testing.T) {
	imdb, err := NewIMDB(DefaultOptions, "560188055")
	assert.NoError(t, err)

	watchList, err := imdb.ExportWatchList()
	assert.NoError(t, err)
	assert.NotEmpty(t, watchList)
}
