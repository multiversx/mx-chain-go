package serviceContainer_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/serviceContainer"
	"github.com/stretchr/testify/assert"
)

func TestServiceContainer_NewServiceContainerEmpty(t *testing.T) {
	sc, err := serviceContainer.NewServiceContainer()
	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Nil(t, sc.Indexer())
}

/*func TestServiceContainer_NewServiceContainerWithIndexer(t *testing.T) {
	indexer := &mock.IndexerMock{}
	sc, err := core.NewServiceContainer(core.WithIndexer(indexer))

	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Equal(t, indexer, sc.Indexer())
}*/
