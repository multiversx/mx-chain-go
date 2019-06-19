package serviceContainer_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/mock"

	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/stretchr/testify/assert"
)

func TestServiceContainer_NewServiceContainerEmpty(t *testing.T) {
	sc, err := serviceContainer.NewServiceContainer()
	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Nil(t, sc.Indexer())
}

func TestServiceContainer_NewServiceContainerWithIndexer(t *testing.T) {
	indexer := &mock.IndexerMock{}
	sc, err := serviceContainer.NewServiceContainer(serviceContainer.WithIndexer(indexer))

	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Equal(t, indexer, sc.Indexer())
}

func TestServiceContainer_NewServiceContainerWithNilIndexer(t *testing.T) {
	indexer := &mock.IndexerMock{}
	indexer = nil

	sc, err := serviceContainer.NewServiceContainer(serviceContainer.WithIndexer(indexer))

	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Nil(t, sc.Indexer())
}

func TestServiceContainer_NewServiceContainerWithTPSBenchmark(t *testing.T) {
	tpsBenchmark := &mock.TpsBenchmarkMock{}

	sc, err := serviceContainer.NewServiceContainer(serviceContainer.WithTPSBenchmark(tpsBenchmark))
	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Equal(t, tpsBenchmark, sc.TPSBenchmark())
}

func TestServiceContainer_NewServiceContainerWithNilTPSBenchmark(t *testing.T) {
	tpsBenchmark := &mock.TpsBenchmarkMock{}
	tpsBenchmark = nil

	sc, err := serviceContainer.NewServiceContainer(serviceContainer.WithTPSBenchmark(tpsBenchmark))
	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Nil(t, sc.TPSBenchmark())
}
