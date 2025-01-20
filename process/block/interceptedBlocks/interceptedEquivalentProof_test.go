package interceptedBlocks

import (
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
)

var (
	expectedErr    = errors.New("expected error")
	testMarshaller = &marshallerMock.MarshalizerMock{}
	providedEpoch  = uint32(123)
	providedNonce  = uint64(345)
	providedShard  = uint32(0)
	providedRound  = uint64(123456)
)

func createMockDataBuff() []byte {
	proof := &block.HeaderProof{
		PubKeysBitmap:       []byte("bitmap"),
		AggregatedSignature: []byte("sig"),
		HeaderHash:          []byte("hash"),
		HeaderEpoch:         providedEpoch,
		HeaderNonce:         providedNonce,
		HeaderShardId:       providedShard,
		HeaderRound:         providedRound,
	}

	dataBuff, _ := testMarshaller.Marshal(proof)
	return dataBuff
}

func createMockArgInterceptedEquivalentProof() ArgInterceptedEquivalentProof {
	return ArgInterceptedEquivalentProof{
		DataBuff:          createMockDataBuff(),
		Marshaller:        testMarshaller,
		ShardCoordinator:  &mock.ShardCoordinatorMock{},
		HeaderSigVerifier: &consensus.HeaderSigVerifierMock{},
		Proofs:            &dataRetriever.ProofsPoolMock{},
		Hasher:            &hashingMocks.HasherMock{},
		Headers: &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{
					EpochField: providedEpoch,
					RoundField: providedRound,
					GetNonceCalled: func() uint64 {
						return providedNonce
					},
					GetShardIDCalled: func() uint32 {
						return providedShard
					},
				}, nil
			},
		},
	}
}

func TestInterceptedEquivalentProof_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var iep *interceptedEquivalentProof
	require.True(t, iep.IsInterfaceNil())

	iep, _ = NewInterceptedEquivalentProof(createMockArgInterceptedEquivalentProof())
	require.False(t, iep.IsInterfaceNil())
}

func TestNewInterceptedEquivalentProof(t *testing.T) {
	t.Parallel()

	t.Run("nil DataBuff should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.DataBuff = nil
		iep, err := NewInterceptedEquivalentProof(args)
		require.Equal(t, process.ErrNilBuffer, err)
		require.Nil(t, iep)
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.Marshaller = nil
		iep, err := NewInterceptedEquivalentProof(args)
		require.Equal(t, process.ErrNilMarshalizer, err)
		require.Nil(t, iep)
	})
	t.Run("nil ShardCoordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.ShardCoordinator = nil
		iep, err := NewInterceptedEquivalentProof(args)
		require.Equal(t, process.ErrNilShardCoordinator, err)
		require.Nil(t, iep)
	})
	t.Run("nil HeaderSigVerifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.HeaderSigVerifier = nil
		iep, err := NewInterceptedEquivalentProof(args)
		require.Equal(t, process.ErrNilHeaderSigVerifier, err)
		require.Nil(t, iep)
	})
	t.Run("nil proofs pool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.Proofs = nil
		iep, err := NewInterceptedEquivalentProof(args)
		require.Equal(t, process.ErrNilProofsPool, err)
		require.Nil(t, iep)
	})
	t.Run("nil headers pool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.Headers = nil
		iep, err := NewInterceptedEquivalentProof(args)
		require.Equal(t, process.ErrNilHeadersDataPool, err)
		require.Nil(t, iep)
	})
	t.Run("nil Hasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.Hasher = nil
		iep, err := NewInterceptedEquivalentProof(args)
		require.Equal(t, process.ErrNilHasher, err)
		require.Nil(t, iep)
	})
	t.Run("unmarshal error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.Marshaller = &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		iep, err := NewInterceptedEquivalentProof(args)
		require.Equal(t, expectedErr, err)
		require.Nil(t, iep)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		iep, err := NewInterceptedEquivalentProof(createMockArgInterceptedEquivalentProof())
		require.NoError(t, err)
		require.NotNil(t, iep)
	})
}

func TestInterceptedEquivalentProof_CheckValidity(t *testing.T) {
	t.Parallel()

	t.Run("invalid proof should error", func(t *testing.T) {
		t.Parallel()

		// no header hash
		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("sig"),
		}
		args := createMockArgInterceptedEquivalentProof()
		args.DataBuff, _ = args.Marshaller.Marshal(proof)
		iep, err := NewInterceptedEquivalentProof(args)
		require.NoError(t, err)

		err = iep.CheckValidity()
		require.Equal(t, ErrInvalidProof, err)
	})
	t.Run("already exiting proof should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedEquivalentProof()
		args.Proofs = &dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		}

		iep, err := NewInterceptedEquivalentProof(args)
		require.NoError(t, err)

		err = iep.CheckValidity()
		require.Equal(t, proofscache.ErrAlreadyExistingEquivalentProof, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		iep, err := NewInterceptedEquivalentProof(createMockArgInterceptedEquivalentProof())
		require.NoError(t, err)

		err = iep.CheckValidity()
		require.NoError(t, err)
	})
}

func TestInterceptedEquivalentProof_IsForCurrentShard(t *testing.T) {
	t.Parallel()

	t.Run("meta should return true", func(t *testing.T) {
		t.Parallel()

		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("sig"),
			HeaderHash:          []byte("hash"),
			HeaderShardId:       core.MetachainShardId,
		}
		args := createMockArgInterceptedEquivalentProof()
		args.DataBuff, _ = args.Marshaller.Marshal(proof)
		args.ShardCoordinator = &mock.ShardCoordinatorMock{ShardID: core.MetachainShardId}
		iep, err := NewInterceptedEquivalentProof(args)
		require.NoError(t, err)

		require.True(t, iep.IsForCurrentShard())
	})
	t.Run("self shard id return true", func(t *testing.T) {
		t.Parallel()

		selfShardId := uint32(1234)
		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("sig"),
			HeaderHash:          []byte("hash"),
			HeaderShardId:       selfShardId,
		}
		args := createMockArgInterceptedEquivalentProof()
		args.DataBuff, _ = args.Marshaller.Marshal(proof)
		args.ShardCoordinator = &mock.ShardCoordinatorMock{ShardID: selfShardId}
		iep, err := NewInterceptedEquivalentProof(args)
		require.NoError(t, err)

		require.True(t, iep.IsForCurrentShard())
	})
	t.Run("other shard id return true", func(t *testing.T) {
		t.Parallel()

		selfShardId := uint32(1234)
		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("sig"),
			HeaderHash:          []byte("hash"),
			HeaderShardId:       selfShardId,
		}
		args := createMockArgInterceptedEquivalentProof()
		args.DataBuff, _ = args.Marshaller.Marshal(proof)
		iep, err := NewInterceptedEquivalentProof(args)
		require.NoError(t, err)

		require.False(t, iep.IsForCurrentShard())
	})
}

func TestInterceptedEquivalentProof_Getters(t *testing.T) {
	t.Parallel()

	proof := &block.HeaderProof{
		PubKeysBitmap:       []byte("bitmap"),
		AggregatedSignature: []byte("sig"),
		HeaderHash:          []byte("hash"),
		HeaderEpoch:         123,
		HeaderNonce:         345,
		HeaderShardId:       0,
	}
	args := createMockArgInterceptedEquivalentProof()
	args.DataBuff, _ = args.Marshaller.Marshal(proof)
	hash := args.Hasher.Compute(string(args.DataBuff))
	iep, err := NewInterceptedEquivalentProof(args)
	require.NoError(t, err)

	require.Equal(t, proof, iep.GetProof()) // pointer testing
	require.Equal(t, hash, iep.Hash())
	require.Equal(t, [][]byte{proof.HeaderHash}, iep.Identifiers())
	require.Equal(t, interceptedEquivalentProofType, iep.Type())
	expectedStr := fmt.Sprintf("bitmap=%s, signature=%s, hash=%s, epoch=123, shard=0, nonce=345",
		logger.DisplayByteSlice(proof.PubKeysBitmap),
		logger.DisplayByteSlice(proof.AggregatedSignature),
		logger.DisplayByteSlice(proof.HeaderHash))
	require.Equal(t, expectedStr, iep.String())
}
