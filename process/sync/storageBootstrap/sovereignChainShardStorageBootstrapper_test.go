package storageBootstrap

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewSovereignChainShardStorageBootstrapper(t *testing.T) {
	t.Parallel()

	baseArgs := createMockShardStorageBoostrapperArgs()
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}

	t.Run("should error when shard storage bootstrapper is nil", func(t *testing.T) {
		t.Parallel()

		scesb, err := NewSovereignChainShardStorageBootstrapper(nil)

		assert.Nil(t, scesb)
		assert.Equal(t, process.ErrNilShardStorageBootstrapper, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ssb, err := NewShardStorageBootstrapper(args)
		scssb, err := NewSovereignChainShardStorageBootstrapper(ssb)

		assert.NotNil(t, scssb)
		assert.Nil(t, err)
	})
}
