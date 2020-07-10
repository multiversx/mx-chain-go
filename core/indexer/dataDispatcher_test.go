package indexer_test

import (
	"fmt"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
)

var log = logger.GetOrCreate("coretest/indexer")

func TestDataDispatcher(t *testing.T) {
	_, err := indexer.NewDataDispatcher(indexer.ElasticIndexerArgs{
		Url: "http://localhost:9200",
	})
	if err != nil {
		fmt.Println(err)
	}


}
