package sync_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/stretchr/testify/assert"
)

func TestNewSovereignChainShardBootstrap_ShouldErrNilShardBootstrap(t *testing.T) {
	t.Parallel()

	scsb, err := sync.NewSovereignChainShardBootstrap(nil)
	assert.Nil(t, scsb)
	assert.Equal(t, process.ErrNilShardBootstrap, err)
}

func TestNewSovereignChainShardBootstrap_ShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	sb, _ := sync.NewShardBootstrap(args)

	scsb, err := sync.NewSovereignChainShardBootstrap(sb)
	assert.NotNil(t, scsb)
	assert.Nil(t, err)
}
