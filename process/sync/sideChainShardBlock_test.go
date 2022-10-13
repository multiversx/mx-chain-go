package sync_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/stretchr/testify/assert"
)

func TestNewSideChainShardBootstrap_ShouldErrNilShardBootstrap(t *testing.T) {
	t.Parallel()

	scsb, err := sync.NewSideChainShardBootstrap(nil)
	assert.Nil(t, scsb)
	assert.Equal(t, process.ErrNilShardBootstrap, err)
}

func TestNewSideChainShardBootstrap_ShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	sb, _ := sync.NewShardBootstrap(args)

	scsb, err := sync.NewSideChainShardBootstrap(sb)
	assert.NotNil(t, scsb)
	assert.Nil(t, err)
}
