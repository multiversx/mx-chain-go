package storageBootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
)

func TestBaseStorageBootstrapper_SovereignChainGetScheduledRootHash(t *testing.T) {
	t.Parallel()

	baseArgs := createMockShardStorageBoostrapperArgs()
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}

	ssb, _ := NewShardStorageBootstrapper(args)
	scssb, _ := NewSovereignChainShardStorageBootstrapper(ssb)

	expectedRootHash := []byte("rootHash")
	hdr := &block.Header{
		RootHash: expectedRootHash,
	}
	rootHash := scssb.sovereignChainGetScheduledRootHash(hdr, nil)

	assert.Equal(t, expectedRootHash, rootHash)
}
