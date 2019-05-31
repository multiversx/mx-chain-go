package indexer_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/indexer"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/mock"
	"github.com/stretchr/testify/assert"
)

var (
	url              = "https://elrond.com"
	shardCoordinator = mock.ShardCoordinatorMock{}
	marshalizer      = &mock.MarshalizerMock{}
	hasher           = mock.HasherMock{}
	log              = logger.DefaultLogger()
)

func TestNewElasticIndexerIncorrectUrl(t *testing.T) {
	url := string([]byte{1, 2, 3})

	ind, err := indexer.NewElasticIndexer(url, shardCoordinator, marshalizer, hasher, log)
	assert.Nil(t, ind)
	assert.NotNil(t, err)
}
