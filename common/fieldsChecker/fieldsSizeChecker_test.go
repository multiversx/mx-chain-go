package fieldsChecker_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/fieldsChecker"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/stretchr/testify/require"
)

func TestNewFieldsSizeChecker(t *testing.T) {
	t.Parallel()

	t.Run("nil chain parameters handler", func(t *testing.T) {
		t.Parallel()

		fsc, err := fieldsChecker.NewFieldsSizeChecker(
			nil,
			&testscommon.HasherStub{},
		)
		require.Equal(t, errors.ErrNilChainParametersHandler, err)
		require.Nil(t, fsc)
	})

	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		fsc, err := fieldsChecker.NewFieldsSizeChecker(
			&chainParameters.ChainParametersHandlerStub{},
			nil,
		)
		require.Equal(t, core.ErrNilHasher, err)
		require.Nil(t, fsc)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		fsc, err := fieldsChecker.NewFieldsSizeChecker(
			&chainParameters.ChainParametersHandlerStub{},
			&testscommon.HasherStub{},
		)
		require.Nil(t, err)
		require.NotNil(t, fsc)
		require.False(t, fsc.IsInterfaceNil())
	})
}

func TestFieldsSizeChecker_IsProofSizeValid(t *testing.T) {
	t.Parallel()

	fsc, err := fieldsChecker.NewFieldsSizeChecker(
		&chainParameters.ChainParametersHandlerStub{
			ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{
					ShardConsensusGroupSize: 8,
				}, nil
			},
		},
		&testscommon.HasherStub{
			SizeCalled: func() int {
				return 32
			},
		},
	)
	require.Nil(t, err)

	ok := fsc.IsProofSizeValid(&block.HeaderProof{
		PubKeysBitmap:       []byte{1},
		AggregatedSignature: []byte("aggSig"),
		HeaderHash:          bytes.Repeat([]byte("h"), 32),
		HeaderShardId:       1,
	})
	require.True(t, ok)

	ok = fsc.IsProofSizeValid(&block.HeaderProof{
		PubKeysBitmap:       []byte{1, 2}, // bigger bitmap
		AggregatedSignature: []byte("aggSig"),
		HeaderHash:          bytes.Repeat([]byte("h"), 32),
		HeaderShardId:       1,
	})
	require.False(t, ok)

	ok = fsc.IsProofSizeValid(&block.HeaderProof{
		PubKeysBitmap:       []byte{1},
		AggregatedSignature: []byte{}, // empty agg sig
		HeaderHash:          bytes.Repeat([]byte("h"), 32),
		HeaderShardId:       1,
	})
	require.False(t, ok)

	ok = fsc.IsProofSizeValid(&block.HeaderProof{
		PubKeysBitmap:       []byte{1},
		AggregatedSignature: []byte("aggSig"),
		HeaderHash:          bytes.Repeat([]byte("h"), 33), // bigger hash size
		HeaderShardId:       1,
	})
	require.False(t, ok)

	ok = fsc.IsProofSizeValid(&block.HeaderProof{
		PubKeysBitmap:       []byte{1},
		AggregatedSignature: bytes.Repeat([]byte("h"), 101), // bigger sig size
		HeaderHash:          bytes.Repeat([]byte("h"), 32),
		HeaderShardId:       1,
	})
	require.False(t, ok)
}
