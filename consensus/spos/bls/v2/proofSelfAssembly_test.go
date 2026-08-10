package v2

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclMultisig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	factoryCrypto "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverTests "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

const selfAssemblyGroupSize = 9

type selfAssemblyTestSetup struct {
	subround       *spos.Subround
	store          *signatureEvidenceStore
	signingHandler consensus.SigningHandler
	keys           []string
	proofsPool     *dataRetrieverTests.ProofsPoolMock
	broadcast      *atomic.Pointer[block.HeaderProof]
	numAdded       *atomic.Int32
}

func createRealBLSSigningHandler(t *testing.T, numKeys int) ([]string, map[string]crypto.PrivateKey, consensus.SigningHandler) {
	llSigner := &mclMultisig.BlsMultiSignerKOSK{}
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)

	multiSigHandler, err := multisig.NewBLSMultisig(llSigner, kg)
	require.Nil(t, err)

	mapKeys := make(map[string]crypto.PrivateKey)
	for i := 0; i < numKeys; i++ {
		sk, pk := kg.GeneratePair()
		pubKey, _ := pk.ToByteArray()
		mapKeys[string(pubKey)] = sk
	}
	keys := initializers.CreateEligibleListFromMap(mapKeys)

	keysHandlerMock := &testscommon.KeysHandlerStub{
		IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
			_, exists := mapKeys[string(pkBytes)]
			return exists
		},
		GetHandledPrivateKeyCalled: func(pkBytes []byte) crypto.PrivateKey {
			return mapKeys[string(pkBytes)]
		},
	}

	pubKeysCache, err := cache.NewLRUCache(1000)
	require.Nil(t, err)

	signingHandler, err := factoryCrypto.NewSigningHandler(factoryCrypto.ArgsSigningHandler{
		PubKeys: keys,
		MultiSignerContainer: &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSignerV2, error) {
				return multiSigHandler, nil
			},
		},
		SingleSigner: &cryptoMocks.SingleSignerStub{},
		KeyGenerator: kg,
		KeysHandler:  keysHandlerMock,
		PubKeysCache: pubKeysCache,
	})
	require.Nil(t, err)

	return keys, mapKeys, signingHandler
}

func createSelfAssemblySetup(t *testing.T) *selfAssemblyTestSetup {
	keys, mapKeys, signingHandler := createRealBLSSigningHandler(t, selfAssemblyGroupSize)

	container := consensusMocks.InitConsensusCore()
	container.SetSigningHandler(signingHandler)

	var broadcast atomic.Pointer[block.HeaderProof]
	var numAdded atomic.Int32
	proofsPool := &dataRetrieverTests.ProofsPoolMock{
		GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
			return nil, errors.New("proof not found")
		},
		AddProofIfNoneAtNonceCalled: func(headerProof data.HeaderProofHandler) (bool, data.HeaderProofHandler) {
			numAdded.Add(1)
			return true, nil
		},
	}
	container.SetEquivalentProofsPool(proofsPool)
	container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
		BroadcastEquivalentProofCalled: func(proof data.HeaderProofHandler, pkBytes []byte) error {
			broadcast.Store(proof.(*block.HeaderProof))
			return nil
		},
	})

	keysHandlerMock := &testscommon.KeysHandlerStub{
		IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
			_, exists := mapKeys[string(pkBytes)]
			return exists
		},
	}
	consensusState := initializers.InitConsensusStateWithArgsVerifySignature(keysHandlerMock, keys)
	ch := make(chan bool, 1)

	sr, err := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		flowRoundTimeDuration,
		0.25,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		func() {},
		container,
		flowChainID,
		flowCurrentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	require.Nil(t, err)

	store, err := newSignatureEvidenceStore(proofsPool)
	require.Nil(t, err)

	return &selfAssemblyTestSetup{
		subround:       sr,
		store:          store,
		signingHandler: signingHandler,
		keys:           keys,
		proofsPool:     proofsPool,
		broadcast:      &broadcast,
		numAdded:       &numAdded,
	}
}

// createRealEvidence builds evidence with real BLS shares over headerHash; corruptIndices
// get shares signed over a different message so they fail verification
func (setup *selfAssemblyTestSetup) createRealEvidence(t *testing.T, headerHash []byte, threshold int, corruptIndices ...int) *roundSignatureEvidence {
	corrupt := make(map[int]struct{})
	for _, idx := range corruptIndices {
		corrupt[idx] = struct{}{}
	}

	bitmap := make([]byte, len(setup.keys)/8+1)
	shares := make([][]byte, len(setup.keys))
	for i, pk := range setup.keys {
		message := headerHash
		if _, isCorrupt := corrupt[i]; isCorrupt {
			message = []byte("another message")
		}

		share, err := setup.signingHandler.CreateSignatureShareForPublicKey(context.Background(), message, uint16(i), 0, []byte(pk))
		require.Nil(t, err)

		shares[i] = share
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	return &roundSignatureEvidence{
		roundIndex:     10,
		nonce:          5,
		headerHash:     headerHash,
		epoch:          0,
		headerRound:    20,
		shardID:        0,
		threshold:      threshold,
		consensusGroup: setup.keys,
		bitmap:         bitmap,
		shares:         shares,
		count:          len(setup.keys),
	}
}

func TestTrySelfAssembleProof_HappyPath(t *testing.T) {
	t.Parallel()

	setup := createSelfAssemblySetup(t)
	headerHash := []byte("header hash to be signed")
	ev := setup.createRealEvidence(t, headerHash, 7)

	trySelfAssembleProof(setup.subround, setup.store, ev)

	proof := setup.broadcast.Load()
	require.NotNil(t, proof, "proof should have been broadcast")
	assert.Equal(t, headerHash, proof.HeaderHash)
	assert.Equal(t, uint64(5), proof.HeaderNonce)
	assert.Equal(t, uint64(20), proof.HeaderRound)
	assert.Equal(t, int32(1), setup.numAdded.Load())

	err := setup.signingHandler.VerifyAggregatedSigWithKeys(setup.keys, proof.PubKeysBitmap, headerHash, proof.AggregatedSignature, 0)
	assert.Nil(t, err, "the self-assembled aggregated signature must verify")
}

func TestTrySelfAssembleProof_SingleAttemptPerRound(t *testing.T) {
	t.Parallel()

	setup := createSelfAssemblySetup(t)
	ev := setup.createRealEvidence(t, []byte("header hash"), 7)

	trySelfAssembleProof(setup.subround, setup.store, ev)
	trySelfAssembleProof(setup.subround, setup.store, ev)

	assert.Equal(t, int32(1), setup.numAdded.Load(), "a second attempt in the same round must be skipped")
}

func TestTrySelfAssembleProof_StripsInvalidSharesAndRetries(t *testing.T) {
	t.Parallel()

	setup := createSelfAssemblySetup(t)
	headerHash := []byte("header hash to be signed")
	// one corrupt share: 8 valid shares remain, still above threshold 7
	ev := setup.createRealEvidence(t, headerHash, 7, 3)

	trySelfAssembleProof(setup.subround, setup.store, ev)

	proof := setup.broadcast.Load()
	require.NotNil(t, proof, "proof should have been broadcast after stripping the invalid share")
	assert.Equal(t, 8, ev.getCount())
	assert.Equal(t, byte(0), proof.PubKeysBitmap[0]&(1<<3), "the corrupt share must be excluded from the bitmap")

	err := setup.signingHandler.VerifyAggregatedSigWithKeys(setup.keys, proof.PubKeysBitmap, headerHash, proof.AggregatedSignature, 0)
	assert.Nil(t, err)
}

func TestTrySelfAssembleProof_DemotesBelowThreshold(t *testing.T) {
	t.Parallel()

	setup := createSelfAssemblySetup(t)
	headerHash := []byte("header hash to be signed")
	// three corrupt shares: 6 valid shares remain, below threshold 7
	ev := setup.createRealEvidence(t, headerHash, 7, 2, 4, 6)

	// place the evidence in the retained slot to verify demotion drops it
	setup.store.Capture(ev)
	setup.store.Capture(nil)
	_, ok := setup.store.GetRetainedQuorumEvidence(ev.nonce)
	require.True(t, ok)

	trySelfAssembleProof(setup.subround, setup.store, ev)

	assert.Nil(t, setup.broadcast.Load(), "no proof should be broadcast below threshold")
	assert.Equal(t, int32(0), setup.numAdded.Load())
	assert.Equal(t, 6, ev.getCount())

	_, ok = setup.store.GetRetainedQuorumEvidence(ev.nonce)
	assert.False(t, ok, "demoted evidence must be dropped from the retained slot")
}

func TestTrySelfAssembleProof_AbortsOnCompetingProof(t *testing.T) {
	t.Parallel()

	t.Run("competing proof at entry", func(t *testing.T) {
		t.Parallel()

		setup := createSelfAssemblySetup(t)
		ev := setup.createRealEvidence(t, []byte("header hash"), 7)

		setup.proofsPool.GetProofByNonceCalled = func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
			return &block.HeaderProof{HeaderHash: []byte("competing hash"), HeaderNonce: headerNonce}, nil
		}

		setup.store.Capture(ev)
		setup.store.Capture(nil)

		trySelfAssembleProof(setup.subround, setup.store, ev)

		assert.Nil(t, setup.broadcast.Load())
		assert.Equal(t, int32(0), setup.numAdded.Load())

		_, ok := setup.store.GetRetainedQuorumEvidence(ev.nonce)
		assert.False(t, ok, "settled nonce must drop the retained slot")
	})

	t.Run("competing proof arrives during assembly", func(t *testing.T) {
		t.Parallel()

		setup := createSelfAssemblySetup(t)
		ev := setup.createRealEvidence(t, []byte("header hash"), 7)

		// entry check finds no proof, the atomic add is beaten by a competing proof
		setup.proofsPool.AddProofIfNoneAtNonceCalled = func(headerProof data.HeaderProofHandler) (bool, data.HeaderProofHandler) {
			return false, &block.HeaderProof{HeaderHash: []byte("competing hash"), HeaderNonce: headerProof.GetHeaderNonce()}
		}

		setup.store.Capture(ev)
		setup.store.Capture(nil)

		trySelfAssembleProof(setup.subround, setup.store, ev)

		assert.Nil(t, setup.broadcast.Load(), "broadcast must be aborted by the atomic add")

		_, ok := setup.store.GetRetainedQuorumEvidence(ev.nonce)
		assert.False(t, ok, "settled nonce must drop the retained slot")
	})

	t.Run("proof for the own hash means already done", func(t *testing.T) {
		t.Parallel()

		setup := createSelfAssemblySetup(t)
		ev := setup.createRealEvidence(t, []byte("header hash"), 7)

		setup.proofsPool.GetProofByNonceCalled = func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
			return &block.HeaderProof{HeaderHash: ev.headerHash, HeaderNonce: headerNonce}, nil
		}

		trySelfAssembleProof(setup.subround, setup.store, ev)

		assert.Nil(t, setup.broadcast.Load())
		assert.Equal(t, int32(0), setup.numAdded.Load())
	})
}

func TestTrySelfAssembleProof_EpochBoundaryUsesSnapshotGroup(t *testing.T) {
	t.Parallel()

	setup := createSelfAssemblySetup(t)
	headerHash := []byte("header hash to be signed")
	ev := setup.createRealEvidence(t, headerHash, 7)

	// simulate the next epoch: the per-round signing state is reset with a rotated group
	rotatedKeys, _, _ := createRealBLSSigningHandler(t, selfAssemblyGroupSize)
	require.Nil(t, setup.signingHandler.Reset(rotatedKeys))

	trySelfAssembleProof(setup.subround, setup.store, ev)

	proof := setup.broadcast.Load()
	require.NotNil(t, proof, "assembly must work from the snapshot group after the reset")

	err := setup.signingHandler.VerifyAggregatedSigWithKeys(setup.keys, proof.PubKeysBitmap, headerHash, proof.AggregatedSignature, 0)
	assert.Nil(t, err)
}

func TestTrySelfAssembleProof_InProgressAttemptBlocksUntilDone(t *testing.T) {
	t.Parallel()

	var currentRound atomic.Int64
	currentRound.Store(10)

	proceed := make(chan struct{})
	started := make(chan struct{}, 4)
	var numAggregations atomic.Int32

	container := consensusMocks.InitConsensusCore()
	container.SetRoundHandler(&testscommon.RoundHandlerMock{
		IndexCalled: func() int64 {
			return currentRound.Load()
		},
	})
	container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
		AggregateSigsWithKeysCalled: func(pubKeys []string, bitmap []byte, sigShares [][]byte, epoch uint32) ([]byte, error) {
			numAggregations.Add(1)
			started <- struct{}{}
			<-proceed
			return []byte("agg"), nil
		},
	})
	container.SetEquivalentProofsPool(createUnsettledProofsPool())

	sr := createFlowSubround(t, container, bls.SrSignature, "(SIGNATURE)")
	store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())
	ev := createTestEvidence(9, 5, 7, 8)

	// first attempt starts in round 10 and blocks inside the aggregation
	go trySelfAssembleProof(sr, store, ev)
	<-started

	// next round while the first attempt is still executing: no second attempt
	currentRound.Store(11)
	trySelfAssembleProof(sr, store, ev)
	assert.Equal(t, int32(1), numAggregations.Load())

	close(proceed)
	require.Eventually(t, func() bool {
		return !ev.assemblyRunning.Load()
	}, time.Second, time.Millisecond)

	// round 11 had no attempt started, so a retry is allowed after completion
	trySelfAssembleProof(sr, store, ev)
	assert.Equal(t, int32(2), numAggregations.Load())

	// a second attempt in the same round is skipped
	trySelfAssembleProof(sr, store, ev)
	assert.Equal(t, int32(2), numAggregations.Load())

	// next round: retried again
	currentRound.Store(12)
	trySelfAssembleProof(sr, store, ev)
	assert.Equal(t, int32(3), numAggregations.Load())
}

func TestTrySelfAssembleProof_SettledNonceDropsRetainedEvenWhenIneligible(t *testing.T) {
	t.Parallel()

	var numAggregations atomic.Int32
	container := consensusMocks.InitConsensusCore()
	container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
		AggregateSigsWithKeysCalled: func(pubKeys []string, bitmap []byte, sigShares [][]byte, epoch uint32) ([]byte, error) {
			numAggregations.Add(1)
			return []byte("agg"), nil
		},
	})
	container.SetEquivalentProofsPool(createSettledProofsPool([]byte("competing hash")))

	sr := createFlowSubround(t, container, bls.SrSignature, "(SIGNATURE)")
	store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

	ev := createTestEvidence(9, 5, 7, 8)
	for i := range ev.consensusGroup {
		ev.consensusGroup[i] = "other_" + ev.consensusGroup[i]
	}
	store.Capture(ev)
	store.Capture(nil)
	_, ok := store.GetRetainedQuorumEvidence(ev.nonce)
	require.True(t, ok)

	trySelfAssembleProof(sr, store, ev)

	assert.Equal(t, int32(0), numAggregations.Load())
	_, ok = store.GetRetainedQuorumEvidence(ev.nonce)
	assert.False(t, ok, "the settled nonce must drop the retained slot before the eligibility check")
}

func TestTrySelfAssembleProof_KeepsEvidenceWhenAllSharesFailVerification(t *testing.T) {
	t.Parallel()

	var numAdded atomic.Int32
	container := consensusMocks.InitConsensusCore()
	container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
		AggregateSigsWithKeysCalled: func(pubKeys []string, bitmap []byte, sigShares [][]byte, epoch uint32) ([]byte, error) {
			return nil, errors.New("infrastructure error")
		},
		VerifySigShareWithKeyCalled: func(pubKey []byte, sigShare []byte, message []byte, epoch uint32) error {
			return errors.New("infrastructure error")
		},
	})
	proofsPool := createUnsettledProofsPool()
	proofsPool.AddProofIfNoneAtNonceCalled = func(headerProof data.HeaderProofHandler) (bool, data.HeaderProofHandler) {
		numAdded.Add(1)
		return true, nil
	}
	container.SetEquivalentProofsPool(proofsPool)

	sr := createFlowSubround(t, container, bls.SrSignature, "(SIGNATURE)")
	store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

	ev := createTestEvidence(9, 5, 7, 8)
	store.Capture(ev)
	store.Capture(nil)

	trySelfAssembleProof(sr, store, ev)

	assert.Equal(t, 8, ev.getCount(), "evidence must stay intact on an all-shares failure")
	assert.Equal(t, int32(0), numAdded.Load())
	_, ok := store.GetRetainedQuorumEvidence(ev.nonce)
	assert.True(t, ok, "no demotion on an all-shares failure")
}

func TestTrySelfAssembleProof_NotEligibleWithoutGroupKeys(t *testing.T) {
	t.Parallel()

	var numAggregations atomic.Int32
	container := consensusMocks.InitConsensusCore()
	container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
		AggregateSigsWithKeysCalled: func(pubKeys []string, bitmap []byte, sigShares [][]byte, epoch uint32) ([]byte, error) {
			numAggregations.Add(1)
			return []byte("agg"), nil
		},
	})
	container.SetEquivalentProofsPool(createUnsettledProofsPool())

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)
	sr, err := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		flowRoundTimeDuration,
		0.25,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		func() {},
		container,
		flowChainID,
		flowCurrentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	require.Nil(t, err)

	store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

	// evidence group does not contain the self key and none of its keys are managed
	ev := createTestEvidence(10, 5, 7, 8)
	for i := range ev.consensusGroup {
		ev.consensusGroup[i] = "other_" + ev.consensusGroup[i]
	}

	trySelfAssembleProof(sr, store, ev)

	assert.Equal(t, int32(0), numAggregations.Load(), "an ineligible node must not assemble")
}
