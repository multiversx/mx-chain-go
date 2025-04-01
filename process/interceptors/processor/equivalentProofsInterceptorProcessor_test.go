package processor

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	processMocks "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
	"github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
)

var expectedErr = errors.New("expected error")

func createMockArgEquivalentProofsInterceptorProcessor() ArgEquivalentProofsInterceptorProcessor {
	return ArgEquivalentProofsInterceptorProcessor{
		EquivalentProofsPool: &dataRetriever.ProofsPoolMock{},
		Marshaller:           &marshallerMock.MarshalizerMock{},
		PeerShardMapper:      &processMocks.PeerShardMapperStub{},
		NodesCoordinator:     &shardingMocks.NodesCoordinatorMock{},
	}
}

func createInterceptedEquivalentProof(marshaller marshal.Marshalizer) process.InterceptedData {
	argInterceptedEquivalentProof := interceptedBlocks.ArgInterceptedEquivalentProof{
		Marshaller:        marshaller,
		ShardCoordinator:  &mock.ShardCoordinatorMock{},
		HeaderSigVerifier: &consensus.HeaderSigVerifierMock{},
		Proofs:            &dataRetriever.ProofsPoolMock{},
		Headers:           &pool.HeadersPoolStub{},
		Hasher:            &hashingMocks.HasherMock{},
		KeyRWMutexHandler: sync.NewKeyRWMutex(),
	}
	argInterceptedEquivalentProof.DataBuff, _ = argInterceptedEquivalentProof.Marshaller.Marshal(&block.HeaderProof{
		PubKeysBitmap:       []byte("bitmap"),
		AggregatedSignature: []byte("sig"),
		HeaderHash:          []byte("hash"),
		HeaderEpoch:         123,
		HeaderNonce:         345,
		HeaderShardId:       0,
	})
	iep, _ := interceptedBlocks.NewInterceptedEquivalentProof(argInterceptedEquivalentProof)

	return iep
}

func TestEquivalentProofsInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var epip *equivalentProofsInterceptorProcessor
	require.True(t, epip.IsInterfaceNil())

	epip, _ = NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
	require.False(t, epip.IsInterfaceNil())
}

func TestNewEquivalentProofsInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil EquivalentProofsPool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.EquivalentProofsPool = nil

		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.Equal(t, process.ErrNilProofsPool, err)
		require.Nil(t, epip)
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.Marshaller = nil

		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.Equal(t, process.ErrNilMarshalizer, err)
		require.Nil(t, epip)
	})
	t.Run("nil PeerShardMapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.PeerShardMapper = nil

		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.Equal(t, process.ErrNilPeerShardMapper, err)
		require.Nil(t, epip)
	})
	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.NodesCoordinator = nil

		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.Equal(t, process.ErrNilNodesCoordinator, err)
		require.Nil(t, epip)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
		require.NoError(t, err)
		require.NotNil(t, epip)
	})
}

func TestEquivalentProofsInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	t.Run("invalid data should error", func(t *testing.T) {
		t.Parallel()

		epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
		require.NoError(t, err)

		err = epip.Validate(nil, "")
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("nodes coordinator error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return nil, expectedErr
			},
		}
		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.NoError(t, err)

		err = epip.Validate(createInterceptedEquivalentProof(args.Marshaller), "pid")
		require.Equal(t, expectedErr, err)
	})
	t.Run("node not eligible should error", func(t *testing.T) {
		t.Parallel()

		providedPid := core.PeerID("providedPid")
		providedPK := []byte("providedPK")
		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				require.Equal(t, providedPid, pid)
				return core.P2PPeerInfo{
					PkBytes: providedPK,
				}
			},
		}
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return []string{"otherEligible1", "otherEligible2"}, nil
			},
		}
		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.NoError(t, err)

		err = epip.Validate(createInterceptedEquivalentProof(args.Marshaller), providedPid)
		require.True(t, errors.Is(err, process.ErrInvalidHeaderProof))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedPid := core.PeerID("providedPid")
		providedPK := []byte("providedPK")
		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				require.Equal(t, providedPid, pid)
				return core.P2PPeerInfo{
					PkBytes: providedPK,
				}
			},
		}
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return []string{string(providedPK)}, nil
			},
		}
		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.NoError(t, err)

		err = epip.Validate(createInterceptedEquivalentProof(args.Marshaller), providedPid)
		require.NoError(t, err)
	})
}

func TestEquivalentProofsInterceptorProcessor_Save(t *testing.T) {
	t.Parallel()

	t.Run("wrong assertion should error", func(t *testing.T) {
		t.Parallel()

		epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
		require.NoError(t, err)

		err = epip.Save(&transaction.InterceptedTransaction{}, "", "")
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cntWasAdded := 0
		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.EquivalentProofsPool = &dataRetriever.ProofsPoolMock{
			AddProofCalled: func(notarizedProof data.HeaderProofHandler) bool {
				cntWasAdded++
				return cntWasAdded == 1
			},
		}
		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.NoError(t, err)

		iep := createInterceptedEquivalentProof(args.Marshaller)

		err = epip.Save(iep, "", "")
		require.NoError(t, err)
		require.Equal(t, 1, cntWasAdded)

		err = epip.Save(iep, "", "")
		require.Equal(t, common.ErrAlreadyExistingEquivalentProof, err)
		require.Equal(t, 2, cntWasAdded)
	})
}

func TestEquivalentProofsInterceptorProcessor_RegisterHandler(t *testing.T) {
	t.Parallel()

	epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
	require.NoError(t, err)

	// coverage only
	epip.RegisterHandler(nil)
}
