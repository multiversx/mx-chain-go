package preprocess_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestGasComputation_ComputeGasConsumedInShardShouldWork(t *testing.T) {

	gc, _ := preprocess.NewGasComputation(&mock.FeeHandlerStub{})

	gasConsumedInShard, err := gc.ComputeGasConsumedInShard(
		2,
		0,
		1,
		100,
		200)

	assert.Equal(t, err, process.ErrInvalidShardId)

	gasConsumedInShard, _ = gc.ComputeGasConsumedInShard(
		0,
		0,
		1,
		100,
		200)

	assert.Equal(t, gasConsumedInShard, uint64(100))

	gasConsumedInShard, _ = gc.ComputeGasConsumedInShard(
		0,
		1,
		0,
		100,
		200)

	assert.Equal(t, gasConsumedInShard, uint64(200))
}
