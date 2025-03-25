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
	"github.com/multiversx/mx-chain-go/testscommon"
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

func TestVerifyProofAgainstHeader(t *testing.T) {
	t.Parallel()

	t.Run("nil proof or header", func(t *testing.T) {
		t.Parallel()

		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("aggSig"),
			HeaderHash:          []byte("hash"),
			HeaderEpoch:         2,
			HeaderNonce:         2,
			HeaderShardId:       2,
			HeaderRound:         2,
			IsStartOfEpoch:      true,
		}

		header := &block.HeaderV2{
			Header: &block.Header{
				Nonce:              2,
				ShardID:            2,
				Round:              2,
				Epoch:              2,
				EpochStartMetaHash: []byte("epoch start meta hash"),
			},
		}

		err := common.VerifyProofAgainstHeader(nil, header)
		require.Equal(t, common.ErrNilHeaderProof, err)

		err = common.VerifyProofAgainstHeader(proof, nil)
		require.Equal(t, common.ErrNilHeaderHandler, err)
	})

	t.Run("nonce mismatch", func(t *testing.T) {
		t.Parallel()

		proof := &block.HeaderProof{
			HeaderNonce: 2,
		}

		header := &block.HeaderV2{
			Header: &block.Header{
				Nonce: 3,
			},
		}

		err := common.VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, common.ErrInvalidHeaderProof))
	})

	t.Run("round mismatch", func(t *testing.T) {
		t.Parallel()

		proof := &block.HeaderProof{
			HeaderRound: 2,
		}

		header := &block.HeaderV2{
			Header: &block.Header{
				Round: 3,
			},
		}

		err := common.VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, common.ErrInvalidHeaderProof))
	})

	t.Run("epoch mismatch", func(t *testing.T) {
		t.Parallel()

		proof := &block.HeaderProof{
			HeaderEpoch: 2,
		}

		header := &block.HeaderV2{
			Header: &block.Header{
				Epoch: 3,
			},
		}

		err := common.VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, common.ErrInvalidHeaderProof))
	})

	t.Run("shard mismatch", func(t *testing.T) {
		t.Parallel()

		proof := &block.HeaderProof{
			HeaderShardId: 2,
		}

		header := &block.HeaderV2{
			Header: &block.Header{
				ShardID: 3,
			},
		}

		err := common.VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, common.ErrInvalidHeaderProof))
	})

	t.Run("nonce mismatch", func(t *testing.T) {
		t.Parallel()

		proof := &block.HeaderProof{
			IsStartOfEpoch: false,
		}

		header := &block.HeaderV2{
			Header: &block.Header{
				EpochStartMetaHash: []byte("meta blockk hash"),
			},
		}

		err := common.VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, common.ErrInvalidHeaderProof))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("aggSig"),
			HeaderHash:          []byte("hash"),
			HeaderEpoch:         2,
			HeaderNonce:         2,
			HeaderShardId:       2,
			HeaderRound:         2,
			IsStartOfEpoch:      true,
		}

		header := &block.HeaderV2{
			Header: &block.Header{
				Nonce:              2,
				ShardID:            2,
				Round:              2,
				Epoch:              2,
				EpochStartMetaHash: []byte("epoch start meta hash"),
			},
		}

		err := common.VerifyProofAgainstHeader(proof, header)
		require.Nil(t, err)

	})
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

func TestIsEpochStartProofForFlagActivation(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(123)
	eeh := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			require.Equal(t, common.EquivalentMessagesFlag, flag)
			return providedEpoch
		},
	}

	epochStartProofSameEpoch := &block.HeaderProof{
		IsStartOfEpoch: true,
		HeaderEpoch:    providedEpoch,
	}
	notEpochStartProofSameEpoch := &block.HeaderProof{
		IsStartOfEpoch: false,
		HeaderEpoch:    providedEpoch,
	}
	epochStartProofOtherEpoch := &block.HeaderProof{
		IsStartOfEpoch: true,
		HeaderEpoch:    providedEpoch + 1,
	}
	notEpochStartProofOtherEpoch := &block.HeaderProof{
		IsStartOfEpoch: false,
		HeaderEpoch:    providedEpoch + 1,
	}

	require.True(t, common.IsEpochStartProofForFlagActivation(epochStartProofSameEpoch, eeh))
	require.False(t, common.IsEpochStartProofForFlagActivation(notEpochStartProofSameEpoch, eeh))
	require.False(t, common.IsEpochStartProofForFlagActivation(epochStartProofOtherEpoch, eeh))
	require.False(t, common.IsEpochStartProofForFlagActivation(notEpochStartProofOtherEpoch, eeh))
}

func TestShouldBlockHavePrevProof(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(123)
	eeh := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			require.Equal(t, testFlag, flag)
			return epoch >= providedEpoch
		},
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			require.Equal(t, testFlag, flag)
			return providedEpoch
		},
	}

	epochStartHeaderInActivationEpoch := &block.HeaderV2{
		Header: &block.Header{
			EpochStartMetaHash: []byte("meta hash"),
			Epoch:              providedEpoch,
			Nonce:              2,
		},
	}
	notEpochStartHeaderInActivationEpoch := &block.HeaderV2{
		Header: &block.Header{
			Epoch: providedEpoch,
			Nonce: 2,
		},
	}
	epochStartHeaderPrevEpoch := &block.HeaderV2{
		Header: &block.Header{
			EpochStartMetaHash: []byte("meta hash"),
			Epoch:              providedEpoch - 1,
			Nonce:              2,
		},
	}
	notEpochStartHeaderPrevEpoch := &block.HeaderV2{
		Header: &block.Header{
			Epoch: providedEpoch - 1,
			Nonce: 2,
		},
	}
	epochStartHeaderNextEpoch := &block.HeaderV2{
		Header: &block.Header{
			EpochStartMetaHash: []byte("meta hash"),
			Epoch:              providedEpoch + 1,
			Nonce:              2,
		},
	}
	notEpochStartHeaderNextEpoch := &block.HeaderV2{
		Header: &block.Header{
			Epoch: providedEpoch + 1,
			Nonce: 2,
		},
	}

	require.False(t, common.ShouldBlockHavePrevProof(epochStartHeaderInActivationEpoch, eeh, testFlag))
	require.True(t, common.ShouldBlockHavePrevProof(notEpochStartHeaderInActivationEpoch, eeh, testFlag))
	require.False(t, common.ShouldBlockHavePrevProof(epochStartHeaderPrevEpoch, eeh, testFlag))
	require.False(t, common.ShouldBlockHavePrevProof(notEpochStartHeaderPrevEpoch, eeh, testFlag))
	require.True(t, common.ShouldBlockHavePrevProof(epochStartHeaderNextEpoch, eeh, testFlag))
	require.True(t, common.ShouldBlockHavePrevProof(notEpochStartHeaderNextEpoch, eeh, testFlag))
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
