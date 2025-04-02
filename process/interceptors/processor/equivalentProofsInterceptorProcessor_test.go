package processor

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreSync "github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	processMocks "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
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

func createInterceptedEquivalentProof(
	epoch uint32,
	shard uint32,
	marshaller marshal.Marshalizer,
) process.InterceptedData {
	argInterceptedEquivalentProof := interceptedBlocks.ArgInterceptedEquivalentProof{
		Marshaller:        marshaller,
		ShardCoordinator:  &mock.ShardCoordinatorMock{},
		HeaderSigVerifier: &consensus.HeaderSigVerifierMock{},
		Proofs:            &dataRetriever.ProofsPoolMock{},
		Headers:           &pool.HeadersPoolStub{},
		Hasher:            &hashingMocks.HasherMock{},
		ProofSizeChecker:  &testscommon.FieldsSizeCheckerMock{},
		KeyRWMutexHandler: coreSync.NewKeyRWMutex(),
	}
	argInterceptedEquivalentProof.DataBuff, _ = argInterceptedEquivalentProof.Marshaller.Marshal(&block.HeaderProof{
		PubKeysBitmap:       []byte("bitmap"),
		AggregatedSignature: []byte("sig"),
		HeaderHash:          []byte("hash"),
		HeaderEpoch:         epoch,
		HeaderNonce:         345,
		HeaderShardId:       shard,
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
	t.Run("nodes coordinator error on first call should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return nil, expectedErr
			},
		}
		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.NoError(t, err)

		err = epip.Validate(createInterceptedEquivalentProof(0, 0, args.Marshaller), "pid")
		require.Equal(t, expectedErr, err)
	})
	t.Run("nodes coordinator error on second call should error", func(t *testing.T) {
		t.Parallel()

		providedPid := core.PeerID("providedPid")
		providedPK := []byte("providedPK")
		args := createMockArgEquivalentProofsInterceptorProcessor()
		cnt := 0
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				cnt++
				if cnt > 1 {
					return nil, expectedErr
				}
				return []string{string(providedPK)}, nil
			},
		}
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				require.Equal(t, providedPid, pid)
				return core.P2PPeerInfo{
					PkBytes: providedPK,
				}
			},
		}
		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.NoError(t, err)

		err = epip.Validate(createInterceptedEquivalentProof(0, 0, args.Marshaller), providedPid)
		require.NoError(t, err)

		argInterceptedEquivalentProof := interceptedBlocks.ArgInterceptedEquivalentProof{
			Marshaller:        args.Marshaller,
			ShardCoordinator:  &mock.ShardCoordinatorMock{},
			HeaderSigVerifier: &consensus.HeaderSigVerifierMock{},
			Proofs:            &dataRetriever.ProofsPoolMock{},
			Headers:           &pool.HeadersPoolStub{},
			Hasher:            &hashingMocks.HasherMock{},
			ProofSizeChecker:  &testscommon.FieldsSizeCheckerMock{},
			KeyRWMutexHandler: coreSync.NewKeyRWMutex(),
		}
		argInterceptedEquivalentProof.DataBuff, _ = argInterceptedEquivalentProof.Marshaller.Marshal(&block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("sig"),
			HeaderHash:          []byte("hash"),
			HeaderEpoch:         2, // new epoch, same shard, force second call
			HeaderNonce:         345,
			HeaderShardId:       0,
			IsStartOfEpoch:      true, // for extra coverage only, will call nodesCoordinator for epoch 2-1
		})
		iep, _ := interceptedBlocks.NewInterceptedEquivalentProof(argInterceptedEquivalentProof)

		err = epip.Validate(iep, providedPid)
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

		err = epip.Validate(createInterceptedEquivalentProof(0, 0, args.Marshaller), providedPid)
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
		cntGetAllEligibleValidatorsPublicKeysForShardCalled := 0
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				cntGetAllEligibleValidatorsPublicKeysForShardCalled++
				return []string{string(providedPK)}, nil
			},
		}
		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.NoError(t, err)

		err = epip.Validate(createInterceptedEquivalentProof(0, 0, args.Marshaller), providedPid)
		require.NoError(t, err)

		// should return from cache, same epoch and shard
		err = epip.Validate(createInterceptedEquivalentProof(0, 0, args.Marshaller), providedPid)
		require.NoError(t, err)

		// should not return from cache, new shard
		err = epip.Validate(createInterceptedEquivalentProof(0, 1, args.Marshaller), providedPid)
		require.NoError(t, err)

		// should return from cache, same epoch and shard
		err = epip.Validate(createInterceptedEquivalentProof(0, 1, args.Marshaller), providedPid)
		require.NoError(t, err)

		// should not return from cache, new epoch
		err = epip.Validate(createInterceptedEquivalentProof(1, 1, args.Marshaller), providedPid)
		require.NoError(t, err)

		// expecting to calls to nodesCoordinator:
		//	- first call on the initial proof
		//	- second call on the third proof, which has a different shard
		// 	- third call on the last proof, which has a different epoch
		expectedCalls := 3
		require.Equal(t, expectedCalls, cntGetAllEligibleValidatorsPublicKeysForShardCalled)

		// new epoch should reset each shard, shard 1 specific for this test
		eligibleMap := epip.GetEligibleNodesMap()
		require.Equal(t, 2, len(eligibleMap)) // 2 shards
		eligibleListForShard0ByEpoch, ok := eligibleMap[0]
		require.True(t, ok) // must have shard 0
		eligibleListForShard1ByEpoch, ok := eligibleMap[1]
		require.True(t, ok) // must have shard 1

		require.Equal(t, 1, len(eligibleListForShard0ByEpoch)) // tests ran with epoch 0 only
		require.Equal(t, 1, len(eligibleListForShard1ByEpoch)) // cleaned on epoch 0 after epoch 1

		shard0EligibleNodes, ok := eligibleListForShard0ByEpoch[0]
		require.True(t, ok) // must have one eligible for epoch 0
		require.Equal(t, 1, len(shard0EligibleNodes))

		shardMetaEligibleNodes, ok := eligibleListForShard1ByEpoch[1]
		require.True(t, ok) // must have one eligible for epoch 0, cleaned on epoch 0 after epoch 1
		require.Equal(t, 1, len(shardMetaEligibleNodes))
	})
	t.Run("should work, concurrent calls", func(t *testing.T) {
		t.Parallel()

		providedPidShard0 := core.PeerID("providedPid_0")
		providedPKShard0 := []byte("providedPK_0")
		providedPidShardMeta := core.PeerID("providedPid_Meta")
		providedPKShardMeta := []byte("providedPK_Meta")
		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				switch pid {
				case providedPidShard0:
					return core.P2PPeerInfo{
						PkBytes: providedPKShard0,
					}
				case providedPidShardMeta:
					return core.P2PPeerInfo{
						PkBytes: providedPKShardMeta,
					}
				default:
					require.Fail(t, fmt.Sprintf("should have not been called for pid %s", pid))
				}

				return core.P2PPeerInfo{}
			},
		}
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				switch shardID {
				case 0:
					return []string{string(providedPKShard0)}, nil
				case core.MetachainShardId:
					return []string{string(providedPKShardMeta)}, nil
				default:
					require.Fail(t, fmt.Sprintf("should have not been called for shard %d", shardID))
				}

				return nil, expectedErr
			},
		}
		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.NoError(t, err)

		numCalls := 3000
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				shard := uint32(0)
				pid := providedPidShard0
				// 3000 calls => 2000 proofs for shard 0, 1000 for meta
				if idx%3 == 0 {
					shard = core.MetachainShardId
					pid = providedPidShardMeta
				}

				// same epoch 0 each time
				proof := createInterceptedEquivalentProof(0, shard, args.Marshaller)
				err = epip.Validate(proof, pid)
				require.NoError(t, err)

				wg.Done()
			}(i)
		}

		wg.Wait()

		eligibleMap := epip.GetEligibleNodesMap()
		require.Equal(t, 2, len(eligibleMap)) // 2 shards
		eligibleListForShard0ByEpoch, ok := eligibleMap[0]
		require.True(t, ok) // must have shard 0
		eligibleListForShardMetaByEpoch, ok := eligibleMap[core.MetachainShardId]
		require.True(t, ok) // must have shard meta

		require.Equal(t, 1, len(eligibleListForShard0ByEpoch))    // tests ran with epoch 0 only
		require.Equal(t, 1, len(eligibleListForShardMetaByEpoch)) // tests ran with epoch 0 only

		shard0EligibleNodes, ok := eligibleListForShard0ByEpoch[0]
		require.True(t, ok) // must have one eligible for epoch 0
		require.Equal(t, 1, len(shard0EligibleNodes))
		_, ok = shard0EligibleNodes[string(providedPKShard0)]
		require.True(t, ok) // must have only the provided shard 0 pk

		shardMetaEligibleNodes, ok := eligibleListForShardMetaByEpoch[0]
		require.True(t, ok) // must have one eligible for epoch 0
		require.Equal(t, 1, len(shardMetaEligibleNodes))
		_, ok = shardMetaEligibleNodes[string(providedPKShardMeta)]
		require.True(t, ok) // must have only the provided shard meta pk
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

		iep := createInterceptedEquivalentProof(0, 0, args.Marshaller)

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
