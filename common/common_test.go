package common_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/require"
)

var testFlag = core.EnableEpochFlag("test flag")

func TestIsValidRelayedTxV3(t *testing.T) {
	t.Parallel()

	scr := &smartContractResult.SmartContractResult{}
	require.False(t, common.IsValidRelayedTxV3(scr))
	require.False(t, common.IsRelayedTxV3(scr))

	notRelayedTxV3 := &transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(100),
		RcvAddr:   []byte("receiver"),
		SndAddr:   []byte("sender0"),
		GasPrice:  100,
		GasLimit:  10,
		Signature: []byte("signature"),
	}
	require.False(t, common.IsValidRelayedTxV3(notRelayedTxV3))
	require.False(t, common.IsRelayedTxV3(notRelayedTxV3))

	invalidRelayedTxV3 := &transaction.Transaction{
		Nonce:       1,
		Value:       big.NewInt(100),
		RcvAddr:     []byte("receiver"),
		SndAddr:     []byte("sender0"),
		GasPrice:    100,
		GasLimit:    10,
		Signature:   []byte("signature"),
		RelayerAddr: []byte("relayer"),
	}
	require.False(t, common.IsValidRelayedTxV3(invalidRelayedTxV3))
	require.True(t, common.IsRelayedTxV3(invalidRelayedTxV3))

	invalidRelayedTxV3 = &transaction.Transaction{
		Nonce:            1,
		Value:            big.NewInt(100),
		RcvAddr:          []byte("receiver"),
		SndAddr:          []byte("sender0"),
		GasPrice:         100,
		GasLimit:         10,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	}
	require.False(t, common.IsValidRelayedTxV3(invalidRelayedTxV3))
	require.True(t, common.IsRelayedTxV3(invalidRelayedTxV3))

	relayedTxV3 := &transaction.Transaction{
		Nonce:            1,
		Value:            big.NewInt(100),
		RcvAddr:          []byte("receiver"),
		SndAddr:          []byte("sender1"),
		GasPrice:         100,
		GasLimit:         10,
		Signature:        []byte("signature"),
		RelayerAddr:      []byte("relayer"),
		RelayerSignature: []byte("signature"),
	}
	require.True(t, common.IsValidRelayedTxV3(relayedTxV3))
	require.True(t, common.IsRelayedTxV3(relayedTxV3))
}

func TestIsConsensusBitmapValid(t *testing.T) {
	t.Parallel()

	log := &testscommon.LoggerStub{}

	pubKeys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	t.Run("wrong size bitmap", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, len(pubKeys)/8)

		err := common.IsConsensusBitmapValid(log, pubKeys, bitmap, false)
		require.Equal(t, common.ErrWrongSizeBitmap, err)
	})

	t.Run("not enough signatures", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, len(pubKeys)/8+1)
		bitmap[0] = 0x07

		err := common.IsConsensusBitmapValid(log, pubKeys, bitmap, false)
		require.Equal(t, common.ErrNotEnoughSignatures, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, len(pubKeys)/8+1)
		bitmap[0] = 0x77
		bitmap[1] = 0x01

		err := common.IsConsensusBitmapValid(log, pubKeys, bitmap, false)
		require.Nil(t, err)
	})

	t.Run("should work with fallback validation", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, len(pubKeys)/8+1)
		bitmap[0] = 0x77
		bitmap[1] = 0x01

		err := common.IsConsensusBitmapValid(log, pubKeys, bitmap, true)
		require.Nil(t, err)
	})
}

func TestIsEpochChangeBlockForFlagActivation(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(123)
	eeh := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			require.Equal(t, testFlag, flag)
			return providedEpoch
		},
	}

	epochStartHeaderSameEpoch := &block.HeaderV2{
		Header: &block.Header{
			EpochStartMetaHash: []byte("meta hash"),
			Epoch:              providedEpoch,
		},
	}
	notEpochStartHeaderSameEpoch := &block.HeaderV2{
		Header: &block.Header{
			Epoch: providedEpoch,
		},
	}
	epochStartHeaderOtherEpoch := &block.HeaderV2{
		Header: &block.Header{
			EpochStartMetaHash: []byte("meta hash"),
			Epoch:              providedEpoch + 1,
		},
	}
	notEpochStartHeaderOtherEpoch := &block.HeaderV2{
		Header: &block.Header{
			Epoch: providedEpoch + 1,
		},
	}

	require.True(t, common.IsEpochChangeBlockForFlagActivation(epochStartHeaderSameEpoch, eeh, testFlag))
	require.False(t, common.IsEpochChangeBlockForFlagActivation(notEpochStartHeaderSameEpoch, eeh, testFlag))
	require.False(t, common.IsEpochChangeBlockForFlagActivation(epochStartHeaderOtherEpoch, eeh, testFlag))
	require.False(t, common.IsEpochChangeBlockForFlagActivation(notEpochStartHeaderOtherEpoch, eeh, testFlag))
}

func TestGetShardIDs(t *testing.T) {
	t.Parallel()

	shardIDs := common.GetShardIDs(2)
	require.Equal(t, 3, len(shardIDs))
	_, hasShard0 := shardIDs[0]
	require.True(t, hasShard0)
	_, hasShard1 := shardIDs[1]
	require.True(t, hasShard1)
	_, hasShardM := shardIDs[core.MetachainShardId]
	require.True(t, hasShardM)
}

func TestGetBitmapSize(t *testing.T) {
	t.Parallel()

	require.Equal(t, 1, common.GetBitmapSize(8))
	require.Equal(t, 2, common.GetBitmapSize(8+1))
	require.Equal(t, 2, common.GetBitmapSize(8*2-1))
	require.Equal(t, 50, common.GetBitmapSize(8*50)) // 400 consensus size
}

func TestConsesusGroupSizeForShardAndEpoch(t *testing.T) {
	t.Parallel()

	t.Run("shard node", func(t *testing.T) {
		t.Parallel()

		groupSize := uint32(400)

		size := common.ConsensusGroupSizeForShardAndEpoch(
			&testscommon.LoggerStub{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						ShardConsensusGroupSize: groupSize,
					}, nil
				},
			},
			1,
			2,
		)

		require.Equal(t, int(groupSize), size)
	})

	t.Run("meta node", func(t *testing.T) {
		t.Parallel()

		groupSize := uint32(400)

		size := common.ConsensusGroupSizeForShardAndEpoch(
			&testscommon.LoggerStub{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						MetachainConsensusGroupSize: groupSize,
					}, nil
				},
			},
			core.MetachainShardId,
			2,
		)

		require.Equal(t, int(groupSize), size)
	})

	t.Run("on fail, use current parameters", func(t *testing.T) {
		t.Parallel()

		groupSize := uint32(400)

		size := common.ConsensusGroupSizeForShardAndEpoch(
			&testscommon.LoggerStub{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{}, errors.New("fail")
				},
				CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
					return config.ChainParametersByEpochConfig{
						MetachainConsensusGroupSize: groupSize,
					}
				},
			},
			core.MetachainShardId,
			2,
		)

		require.Equal(t, int(groupSize), size)
	})
}
