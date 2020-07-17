package serviceContainer

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	elasticIndexer "github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/assert"
)

func TestServiceContainer_NewServiceContainerEmpty(t *testing.T) {
	sc, err := NewServiceContainer()
	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Nil(t, sc.Indexer())
}

func TestServiceContainer_NewServiceContainerWithIndexer(t *testing.T) {
	indexer := elasticIndexer.NewNilIndexer()
	sc, err := NewServiceContainer(WithIndexer(indexer))

	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Equal(t, indexer, sc.Indexer())
}

func TestServiceContainer_NewServiceContainerWithNilIndexer(t *testing.T) {
	sc, err := NewServiceContainer(WithIndexer(nil))

	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Nil(t, sc.Indexer())
}

func TestServiceContainer_NewServiceContainerWithTPSBenchmark(t *testing.T) {
	tpsBenchmark := &mock.TpsBenchmarkMock{}

	sc, err := NewServiceContainer(WithTPSBenchmark(tpsBenchmark))
	assert.Nil(t, err)
	assert.False(t, check.IfNil(sc))
	assert.Equal(t, tpsBenchmark, sc.TPSBenchmark())
}

func TestServiceContainer_NewServiceContainerWithNilTPSBenchmark(t *testing.T) {
	sc, err := NewServiceContainer(WithTPSBenchmark(nil))
	assert.Nil(t, err)
	assert.NotNil(t, sc)
	assert.Nil(t, sc.TPSBenchmark())
}
