package common_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/require"
)

func TestIsEpochStartProofForFlagActivation(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(123)
	eeh := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			require.Equal(t, common.AndromedaFlag, flag)
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

func TestGetEpochForConsensus(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(10)
	proof := &block.HeaderProof{
		HeaderEpoch:    providedEpoch,
		IsStartOfEpoch: false,
	}

	epoch := common.GetEpochForConsensus(proof)
	require.Equal(t, providedEpoch, epoch)

	proof = &block.HeaderProof{
		HeaderEpoch:    providedEpoch,
		IsStartOfEpoch: true,
	}

	epoch = common.GetEpochForConsensus(proof)
	require.Equal(t, providedEpoch-1, epoch)
}
