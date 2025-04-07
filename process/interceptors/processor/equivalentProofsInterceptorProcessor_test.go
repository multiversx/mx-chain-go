package processor

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
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

func createMockArgEquivalentProofsInterceptorProcessor() ArgEquivalentProofsInterceptorProcessor {
	return ArgEquivalentProofsInterceptorProcessor{
		EquivalentProofsPool: &dataRetriever.ProofsPoolMock{},
		Marshaller:           &marshallerMock.MarshalizerMock{},
	}
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
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
		require.NoError(t, err)
		require.NotNil(t, epip)
	})
}

func TestEquivalentProofsInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
	require.NoError(t, err)

	// coverage only
	require.Nil(t, epip.Validate(nil, ""))
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

		argInterceptedEquivalentProof := interceptedBlocks.ArgInterceptedEquivalentProof{
			Marshaller:        args.Marshaller,
			ShardCoordinator:  &mock.ShardCoordinatorMock{},
			HeaderSigVerifier: &consensus.HeaderSigVerifierMock{},
			Proofs:            &dataRetriever.ProofsPoolMock{},
			Headers:           &pool.HeadersPoolStub{},
			Hasher:            &hashingMocks.HasherMock{},
			ProofSizeChecker:  &testscommon.FieldsSizeCheckerMock{},
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
