package bootstrap

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignChainEpochStartBootstrap(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()

	t.Run("should error when epoch start bootstrapper is nil", func(t *testing.T) {
		t.Parallel()

		scesb, err := NewSovereignChainEpochStartBootstrap(nil)

		assert.Nil(t, scesb)
		assert.Equal(t, errors.ErrNilEpochStartBootstrapper, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		esb, _ := NewEpochStartBootstrap(args)
		scesb, err := NewSovereignChainEpochStartBootstrap(esb)

		assert.NotNil(t, scesb)
		assert.Nil(t, err)
	})
}

func TestGetDataToSync_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	esb, _ := NewEpochStartBootstrap(args)
	scesb, _ := NewSovereignChainEpochStartBootstrap(esb)

	rootHash := []byte("rootHash")
	hdr := &block.Header{
		RootHash: rootHash,
	}

	dts, err := scesb.getDataToSync(nil, hdr)

	require.Nil(t, err)
	assert.Equal(t, hdr, dts.ownShardHdr)
	assert.Equal(t, rootHash, dts.rootHashToSync)
	assert.False(t, dts.withScheduled)
	assert.Nil(t, dts.additionalHeaders)
}
