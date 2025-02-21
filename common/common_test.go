package common

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestIsValidRelayedTxV3(t *testing.T) {
	t.Parallel()

	scr := &smartContractResult.SmartContractResult{}
	require.False(t, IsValidRelayedTxV3(scr))
	require.False(t, IsRelayedTxV3(scr))

	notRelayedTxV3 := &transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(100),
		RcvAddr:   []byte("receiver"),
		SndAddr:   []byte("sender0"),
		GasPrice:  100,
		GasLimit:  10,
		Signature: []byte("signature"),
	}
	require.False(t, IsValidRelayedTxV3(notRelayedTxV3))
	require.False(t, IsRelayedTxV3(notRelayedTxV3))

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
	require.False(t, IsValidRelayedTxV3(invalidRelayedTxV3))
	require.True(t, IsRelayedTxV3(invalidRelayedTxV3))

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
	require.False(t, IsValidRelayedTxV3(invalidRelayedTxV3))
	require.True(t, IsRelayedTxV3(invalidRelayedTxV3))

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
	require.True(t, IsValidRelayedTxV3(relayedTxV3))
	require.True(t, IsRelayedTxV3(relayedTxV3))
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

		err := VerifyProofAgainstHeader(nil, header)
		require.Equal(t, ErrNilHeaderProof, err)

		err = VerifyProofAgainstHeader(proof, nil)
		require.Equal(t, ErrNilHeaderHandler, err)
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

		err := VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, ErrInvalidHeaderProof))
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

		err := VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, ErrInvalidHeaderProof))
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

		err := VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, ErrInvalidHeaderProof))
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

		err := VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, ErrInvalidHeaderProof))
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

		err := VerifyProofAgainstHeader(proof, header)
		require.True(t, errors.Is(err, ErrInvalidHeaderProof))
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

		err := VerifyProofAgainstHeader(proof, header)
		require.Nil(t, err)

	})
}
