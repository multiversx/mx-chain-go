package sync

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	testscommonDataRetriever "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

type branchAwareSyncFixture struct {
	boot          *baseBootstrap
	headers       map[string]data.HeaderHandler
	proofs        []data.HeaderProofHandler
	notarizedHash []byte
}

type ambiguityDuringCheckForkDetector struct {
	*shardForkDetector
	inject   func()
	injected bool
}

func (detector *ambiguityDuringCheckForkDetector) CheckFork() *process.ForkInfo {
	if !detector.injected {
		detector.injected = true
		detector.inject()
	}

	return detector.shardForkDetector.CheckFork()
}

func newBranchAwareSyncFixture(currentHeader data.HeaderHandler, currentHash []byte) *branchAwareSyncFixture {
	fixture := &branchAwareSyncFixture{
		headers: make(map[string]data.HeaderHandler),
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(currentHeader.GetShardID())
	proofsPool := &testscommonDataRetriever.ProofsPoolMock{
		GetProofByNonceCalled: func(_ uint64, _ uint32) (data.HeaderProofHandler, error) {
			if len(fixture.proofs) == 0 {
				return nil, errors.New("missing proof")
			}

			return fixture.proofs[0], nil
		},
		GetProofsByNonceCalled: func(_ uint64, _ uint32) ([]data.HeaderProofHandler, error) {
			if len(fixture.proofs) == 0 {
				return nil, errors.New("missing proofs")
			}

			return fixture.proofs, nil
		},
		GetProofCalled: func(_ uint32, hash []byte) (data.HeaderProofHandler, error) {
			for _, proof := range fixture.proofs {
				if bytes.Equal(proof.GetHeaderHash(), hash) {
					return proof, nil
				}
			}

			return nil, errors.New("missing proof")
		},
		HasProofCalled: func(_ uint32, hash []byte) bool {
			for _, proof := range fixture.proofs {
				if bytes.Equal(proof.GetHeaderHash(), hash) {
					return true
				}
			}

			return false
		},
	}

	fixture.boot = &baseBootstrap{
		headers: &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				header, ok := fixture.headers[string(hash)]
				if !ok {
					return nil, errors.New("missing header")
				}

				return header, nil
			},
		},
		proofs: proofsPool,
		chainHandler: &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return currentHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return currentHash
			},
		},
		forkDetector: &mock.ForkDetectorMock{
			GetNotarizedHeaderHashCalled: func(_ uint64) []byte {
				return fixture.notarizedHash
			},
			ProbableHighestNonceCalled: func() uint64 {
				return currentHeader.GetNonce() + 1
			},
		},
		shardCoordinator: shardCoordinator,
		blackListHandler: &testscommon.TimeCacheStub{},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
				return flag == common.AndromedaFlag || flag == common.SupernovaFlag
			},
		},
		enableRoundsHandler: &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, _ uint64) bool {
				return flag == common.SupernovaRoundFlag
			},
		},
		requestHandler:    &testscommon.RequestHandlerStub{},
		blockBootstrapper: &blockBootstrapperStub{},
		forkInfo:          process.NewForkInfo(),
		chRcvHdrHash:      make(chan bool, 1),
		chRcvHdrNonce:     make(chan bool, 1),
	}

	return fixture
}

func createBranchAwareHeader(nonce uint64, hash []byte, prevHash []byte) (*block.HeaderV3, *block.HeaderProof) {
	header := &block.HeaderV3{
		Nonce:    nonce,
		Round:    nonce + 10,
		Epoch:    1,
		ShardID:  0,
		PrevHash: prevHash,
	}
	proof := &block.HeaderProof{
		HeaderHash:    hash,
		HeaderNonce:   nonce,
		HeaderRound:   header.Round,
		HeaderEpoch:   header.Epoch,
		HeaderShardId: header.ShardID,
	}

	return header, proof
}

func createBranchAwareMetaHeader(nonce uint64, hash []byte, prevHash []byte) (*block.MetaBlockV3, *block.HeaderProof) {
	header := &block.MetaBlockV3{
		Nonce:    nonce,
		Round:    nonce + 10,
		Epoch:    1,
		PrevHash: prevHash,
	}
	proof := &block.HeaderProof{
		HeaderHash:    hash,
		HeaderNonce:   nonce,
		HeaderRound:   header.Round,
		HeaderEpoch:   header.Epoch,
		HeaderShardId: core.MetachainShardId,
	}

	return header, proof
}

func TestBaseBootstrap_GetNextHeaderPreservesDirectedV3Hash(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	directedHash := []byte("D")
	directedHeader, directedProof := createBranchAwareHeader(11, directedHash, currentHash)
	competitorHash := []byte("C")
	competitorHeader, competitorProof := createBranchAwareHeader(11, competitorHash, []byte("B"))
	fixture.headers[string(directedHash)] = directedHeader
	fixture.headers[string(competitorHash)] = competitorHeader
	fixture.proofs = []data.HeaderProofHandler{competitorProof, directedProof}
	fixture.notarizedHash = directedHash

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.NoError(t, err)
	require.Same(t, directedHeader, header)
	require.Equal(t, directedHash, hash)
}

func TestBaseBootstrap_GetNextHeaderPreservesDirectedV3AuthorityWithoutHotEvidence(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	directedHash := []byte("D")
	competitorHash := []byte("C")
	competitorHeader, competitorProof := createBranchAwareHeader(11, competitorHash, currentHash)
	fixture.headers[string(competitorHash)] = competitorHeader
	fixture.proofs = []data.HeaderProofHandler{competitorProof}

	bfd := newBranchAwareForkDetector(0, 10, currentHash)
	bfd.headers[11] = []*headerInfo{{
		epoch: 1, nonce: 11, round: 11, hash: directedHash, prevHash: currentHash,
		state: process.BHNotarized,
	}}
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd
	fixture.boot.forkDetector = sfd
	fixture.boot.forkInfo = &process.ForkInfo{IsDetected: true, Nonce: 11, Hash: competitorHash}
	fixture.boot.store = &storageStubs.ChainStorerStub{}
	fixture.boot.roundHandler = &testscommon.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 0
		},
	}
	var requestedHash []byte
	fixture.boot.requestHandler = &testscommon.RequestHandlerStub{
		RequestShardHeaderCalled: func(_ uint32, hash []byte) {
			requestedHash = hash
		},
	}

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.ErrorIs(t, err, process.ErrTimeIsOut)
	require.Nil(t, header)
	require.Equal(t, directedHash, hash)
	require.Equal(t, directedHash, requestedHash)
}

func TestBaseBootstrap_GetNextHeaderPreservesLegacyProofFallback(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader := &block.Header{Nonce: 10, Round: 10, Epoch: 1, ShardID: 0}
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)
	fixture.boot.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}

	directedHash := []byte("D")
	proofHash := []byte("C")
	proofHeader, proof := createBranchAwareHeader(11, proofHash, currentHash)
	fixture.headers[string(proofHash)] = proofHeader
	fixture.proofs = []data.HeaderProofHandler{proof}

	bfd := newBranchAwareForkDetector(0, 10, currentHash)
	bfd.enableRoundsHandler = &testscommon.EnableRoundsHandlerStub{}
	bfd.headers[11] = []*headerInfo{{
		epoch: 1, nonce: 11, round: 11, hash: directedHash, prevHash: currentHash,
		state: process.BHNotarized,
	}}
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd
	fixture.boot.forkDetector = sfd

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.NoError(t, err)
	require.Same(t, proofHeader, header)
	require.Equal(t, proofHash, hash)
}

func TestBaseBootstrap_GetNextHeaderWaitsForUniqueV3Notarization(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	bfd := newBranchAwareForkDetector(0, 10, currentHash)
	bfd.headers[11] = []*headerInfo{
		{
			epoch: 1, nonce: 11, round: 11, hash: []byte("B"), prevHash: currentHash,
			state: process.BHNotarized,
		},
		{
			epoch: 1, nonce: 11, round: 12, hash: []byte("C"), prevHash: currentHash,
			state: process.BHNotarized,
		},
	}
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd
	fixture.boot.forkDetector = sfd
	fixture.boot.forkInfo = &process.ForkInfo{IsDetected: true, Nonce: 11, Hash: []byte("B")}

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.ErrorIs(t, err, errBranchAwareSyncRetry)
	require.Nil(t, header)
	require.Nil(t, hash)
}

func TestBaseBootstrap_ResolvesAmbiguousNotarizationFromTrackerAuthority(t *testing.T) {
	t.Parallel()

	for _, reverseOrder := range []bool{false, true} {
		reverseOrder := reverseOrder
		t.Run(fmt.Sprintf("reverse=%v", reverseOrder), func(t *testing.T) {
			t.Parallel()

			currentHash := []byte("P")
			currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
			fixture := newBranchAwareSyncFixture(currentHeader, currentHash)
			selectedHash := []byte("B")
			staleHash := []byte("A")

			bfd := newBranchAwareForkDetector(0, 10, currentHash)
			bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: currentHash}
			bfd.fork.checkpoint = []*checkpointInfo{{nonce: 10, round: 10, hash: currentHash}}
			infos := []*headerInfo{
				{epoch: 1, nonce: 11, round: 11, hash: staleHash, prevHash: currentHash, state: process.BHNotarized},
				{epoch: 1, nonce: 11, round: 12, hash: selectedHash, prevHash: currentHash, state: process.BHNotarized},
			}
			if reverseOrder {
				infos[0], infos[1] = infos[1], infos[0]
			}
			bfd.headers[11] = infos
			bfd.hasAmbiguousNotarization.Store(true)
			sfd := &shardForkDetector{baseForkDetector: bfd}
			bfd.forkDetector = sfd
			fixture.boot.forkDetector = sfd
			fixture.boot.settlementChecker = &settlementCheckerStub{
				resolveNotarizedHeaderCalled: func(nonce uint64, candidates []notarizedHeaderCandidate) []byte {
					require.Equal(t, uint64(11), nonce)
					require.Len(t, candidates, 2)
					return selectedHash
				},
			}

			require.False(t, fixture.boot.tryResolveNotarizedAmbiguity(20))
			selection := sfd.getNotarizedHeaderSelection(11)
			require.Equal(t, selectedHash, selection.hash)
			require.Empty(t, selection.candidates)
		})
	}
}

func TestBaseBootstrap_UnresolvedAmbiguityRequestsMissingCandidateEvidenceOncePerRound(t *testing.T) {
	t.Parallel()

	currentHash := []byte("P")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)
	bfd := newBranchAwareForkDetector(0, 10, currentHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: currentHash}
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 11, hash: []byte("A"), prevHash: currentHash, state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 12, hash: []byte("B"), prevHash: currentHash, state: process.BHNotarized},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd
	fixture.boot.forkDetector = sfd
	fixture.boot.settlementChecker = &settlementCheckerStub{}

	requestedHeaders := 0
	requestedProofs := 0
	fixture.boot.requestHandler = &testscommon.RequestHandlerStub{
		RequestShardHeaderForEpochCalled: func(_ uint32, _ []byte, _ uint32) {
			requestedHeaders++
		},
		RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, _ uint32) {
			requestedProofs++
		},
	}

	require.True(t, fixture.boot.tryResolveNotarizedAmbiguity(20))
	require.True(t, fixture.boot.tryResolveNotarizedAmbiguity(20))
	require.Equal(t, 2, requestedHeaders)
	require.Equal(t, 2, requestedProofs)
	require.True(t, fixture.boot.tryResolveNotarizedAmbiguity(21))
	require.Equal(t, 4, requestedHeaders)
	require.Equal(t, 4, requestedProofs)
}

func TestBaseBootstrap_ExactFinalAuthorityUsesRoundGatedReconciliation(t *testing.T) {
	t.Parallel()

	localHash := []byte("A")
	selectedHash := []byte("B")
	parentHash := []byte("P")
	currentHeader, _ := createBranchAwareHeader(11, localHash, parentHash)
	fixture := newBranchAwareSyncFixture(currentHeader, localHash)
	_, selectedProof := createBranchAwareHeader(11, selectedHash, parentHash)
	fixture.proofs = []data.HeaderProofHandler{selectedProof}

	bfd := newBranchAwareForkDetector(0, 11, localHash)
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
	bfd.fork.checkpoint = []*checkpointInfo{
		{nonce: 10, round: 10, hash: parentHash},
		{nonce: 11, round: 21, hash: localHash},
	}
	bfd.fork.rollBackNonce = math.MaxUint64
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 21, hash: localHash, prevHash: parentHash, state: process.BHProcessed, hasProof: true},
		{epoch: 1, nonce: 11, round: 21, hash: localHash, prevHash: parentHash, state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 22, hash: selectedHash, prevHash: parentHash, state: process.BHNotarized},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd
	fixture.boot.forkDetector = sfd
	fixture.boot.statusHandler = &statusHandlerMock.AppStatusHandlerStub{}
	fixture.boot.settlementChecker = &settlementCheckerStub{
		resolveNotarizedHeaderCalled: func(_ uint64, _ []notarizedHeaderCandidate) []byte {
			return selectedHash
		},
	}

	require.True(t, fixture.boot.tryResolveNotarizedAmbiguity(20))
	require.False(t, fixture.boot.tryReconcileEquivocation(20))
	require.Equal(t, uint64(11), sfd.GetHighestFinalBlockNonce())
	require.Equal(t, uint64(10), sfd.settledCheckpoint().nonce)

	require.True(t, fixture.boot.tryReconcileEquivocation(21))
	require.Equal(t, uint64(10), sfd.GetHighestFinalBlockNonce())
	require.Equal(t, uint64(10), sfd.settledCheckpoint().nonce)
	require.Equal(t, uint64(11), sfd.getRollBackNonce())
}

func TestBaseBootstrap_AmbiguityAppearingDuringForkCheckKeepsNodeUnsynchronized(t *testing.T) {
	t.Parallel()

	currentHash := []byte("P")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)
	bfd := newBranchAwareForkDetector(0, 10, currentHash)
	bfd.fork.checkpoint = []*checkpointInfo{{nonce: 10, round: 10, hash: currentHash}}
	bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: currentHash}
	bfd.fork.rollBackNonce = math.MaxUint64
	bfd.setProbableHighestNonce(10)
	bfd.roundHandler = &testscommon.RoundHandlerMock{
		IndexCalled: func() int64 {
			return 20
		},
	}
	bfd.processConfigsHandler = testscommon.GetDefaultProcessConfigsHandler()
	bfd.proofsPool = &testscommonDataRetriever.ProofsPoolMock{}
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd
	detector := &ambiguityDuringCheckForkDetector{
		shardForkDetector: sfd,
		inject: func() {
			bfd.headers[11] = []*headerInfo{
				{epoch: 1, nonce: 11, round: 21, hash: []byte("A"), prevHash: currentHash, state: process.BHNotarized},
				{epoch: 1, nonce: 11, round: 22, hash: []byte("B"), prevHash: currentHash, state: process.BHNotarized},
			}
			bfd.hasAmbiguousNotarization.Store(true)
			bfd.setProbableHighestNonce(11)
		},
	}
	fixture.boot.forkDetector = detector
	fixture.boot.roundHandler = bfd.roundHandler
	fixture.boot.networkWatcher = &p2pmocks.MessengerStub{}
	fixture.boot.statusHandler = &statusHandlerMock.AppStatusHandlerStub{}

	fixture.boot.computeNodeState(20)

	require.True(t, detector.hasUnresolvedNotarizedAmbiguity())
	require.False(t, fixture.boot.isNodeSynchronized)
	require.True(t, fixture.boot.nodeStateHasAmbiguity)

	bfd.mutHeaders.Lock()
	delete(bfd.headers, 11)
	bfd.hasAmbiguousNotarization.Store(false)
	bfd.mutHeaders.Unlock()
	fixture.boot.computeNodeState(20)
	require.False(t, fixture.boot.nodeStateHasAmbiguity)
}

func TestBaseBootstrap_MetaDoesNotRunShardAuthorityResolution(t *testing.T) {
	t.Parallel()

	currentHash := []byte("P")
	currentHeader, _ := createBranchAwareMetaHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)
	bfd := newBranchAwareForkDetector(core.MetachainShardId, 10, currentHash)
	bfd.headers[11] = []*headerInfo{
		{epoch: 1, nonce: 11, round: 21, hash: []byte("A"), prevHash: currentHash, state: process.BHNotarized},
		{epoch: 1, nonce: 11, round: 22, hash: []byte("B"), prevHash: currentHash, state: process.BHNotarized},
	}
	bfd.hasAmbiguousNotarization.Store(true)
	mfd := &metaForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = mfd
	fixture.boot.forkDetector = mfd
	resolverCalls := 0
	fixture.boot.settlementChecker = &settlementCheckerStub{
		resolveNotarizedHeaderCalled: func(_ uint64, _ []notarizedHeaderCandidate) []byte {
			resolverCalls++
			return nil
		},
	}

	require.False(t, fixture.boot.tryResolveNotarizedAmbiguity(20))
	require.False(t, fixture.boot.hasUnresolvedNotarizedAmbiguity())
	require.Zero(t, resolverCalls)
}

func TestBaseBootstrap_GetNextMetaHeaderPreservesMissingDirectedV3Hash(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareMetaHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	directedHash := []byte("D")
	_, directedProof := createBranchAwareMetaHeader(11, directedHash, currentHash)
	competitorHash := []byte("C")
	competitorHeader, competitorProof := createBranchAwareMetaHeader(11, competitorHash, currentHash)
	fixture.headers[string(competitorHash)] = competitorHeader
	fixture.proofs = []data.HeaderProofHandler{competitorProof, directedProof}
	fixture.notarizedHash = directedHash
	fixture.boot.store = &storageStubs.ChainStorerStub{}
	fixture.boot.roundHandler = &testscommon.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 0
		},
	}
	requestedHashes := make([][]byte, 0, 1)
	fixture.boot.requestHandler = &testscommon.RequestHandlerStub{
		RequestMetaHeaderCalled: func(hash []byte) {
			requestedHashes = append(requestedHashes, hash)
		},
	}

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.ErrorIs(t, err, process.ErrTimeIsOut)
	require.Nil(t, header)
	require.Equal(t, directedHash, hash)
	require.Equal(t, [][]byte{directedHash}, requestedHashes)
}

func TestBaseBootstrap_GetNextHeaderSelectsProofExtendingCurrentV3Head(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	offBranchHash := []byte("C")
	offBranchHeader, offBranchProof := createBranchAwareHeader(11, offBranchHash, []byte("B"))
	canonicalHash := []byte("D")
	canonicalHeader, canonicalProof := createBranchAwareHeader(11, canonicalHash, currentHash)
	fixture.headers[string(offBranchHash)] = offBranchHeader
	fixture.headers[string(canonicalHash)] = canonicalHeader
	fixture.proofs = []data.HeaderProofHandler{offBranchProof, canonicalProof}

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.NoError(t, err)
	require.Same(t, canonicalHeader, header)
	require.Equal(t, canonicalHash, hash)
}

func TestBaseBootstrap_GetNextMetaHeaderSelectsProofExtendingCurrentV3Head(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareMetaHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	offBranchHash := []byte("C")
	offBranchHeader, offBranchProof := createBranchAwareMetaHeader(11, offBranchHash, []byte("B"))
	canonicalHash := []byte("D")
	canonicalHeader, canonicalProof := createBranchAwareMetaHeader(11, canonicalHash, currentHash)
	fixture.headers[string(offBranchHash)] = offBranchHeader
	fixture.headers[string(canonicalHash)] = canonicalHeader
	fixture.proofs = []data.HeaderProofHandler{offBranchProof, canonicalProof}

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.NoError(t, err)
	require.Same(t, canonicalHeader, header)
	require.Equal(t, canonicalHash, hash)
}

func TestBaseBootstrap_GetNextMetaHeaderRequestsMissingProofHeadersFromMeta(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareMetaHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	missingHash := []byte("C")
	_, missingProof := createBranchAwareMetaHeader(11, missingHash, currentHash)
	missingProof.HeaderEpoch = 7
	fixture.proofs = []data.HeaderProofHandler{missingProof}

	metaRequests := 0
	shardRequests := 0
	fixture.boot.requestHandler = &testscommon.RequestHandlerStub{
		RequestMetaHeaderForEpochCalled: func(hash []byte, epoch uint32) {
			require.Equal(t, missingHash, hash)
			require.Equal(t, uint32(7), epoch)
			metaRequests++
		},
		RequestShardHeaderForEpochCalled: func(_ uint32, _ []byte, _ uint32) {
			shardRequests++
		},
	}

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.ErrorIs(t, err, errBranchAwareSyncRetry)
	require.Nil(t, header)
	require.Nil(t, hash)
	require.Equal(t, 1, metaRequests)
	require.Zero(t, shardRequests)
}

func TestBaseBootstrap_GetNextHeaderRequestsMissingProofHeadersByExactEpoch(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	firstHash := []byte("C")
	_, firstProof := createBranchAwareHeader(11, firstHash, []byte("B"))
	firstProof.HeaderEpoch = 7
	secondHash := []byte("D")
	_, secondProof := createBranchAwareHeader(11, secondHash, currentHash)
	secondProof.HeaderEpoch = 8
	fixture.proofs = []data.HeaderProofHandler{firstProof, secondProof}

	type request struct {
		hash  []byte
		epoch uint32
	}
	requests := make([]request, 0, 2)
	numSetEpochCalls := 0
	fixture.boot.requestHandler = &requestHandlerWithSetEpochStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderForEpochCalled: func(_ uint32, hash []byte, epoch uint32) {
				requests = append(requests, request{hash: hash, epoch: epoch})
			},
		},
		SetEpochCalled: func(_ uint32) { numSetEpochCalls++ },
	}

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.ErrorIs(t, err, errBranchAwareSyncRetry)
	require.Nil(t, header)
	require.Nil(t, hash)
	require.Equal(t, []request{{hash: firstHash, epoch: 7}, {hash: secondHash, epoch: 8}}, requests)
	require.Nil(t, fixture.boot.requestedHeaderHash())
	require.Zero(t, numSetEpochCalls)
}

func TestBaseBootstrap_GetNextHeaderRequestsUnknownCanonicalCandidateByNonce(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	offBranchHash := []byte("C")
	offBranchHeader, offBranchProof := createBranchAwareHeader(11, offBranchHash, []byte("B"))
	fixture.headers[string(offBranchHash)] = offBranchHeader
	fixture.proofs = []data.HeaderProofHandler{offBranchProof}

	headerRequests := 0
	proofRequests := 0
	fixture.boot.blockBootstrapper = &blockBootstrapperStub{
		requestHeaderByNonceCalled: func(nonce uint64) {
			require.Equal(t, uint64(11), nonce)
			headerRequests++
		},
		requestProofByNonceCalled: func(nonce uint64) {
			require.Equal(t, uint64(11), nonce)
			proofRequests++
		},
	}

	_, _, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.ErrorIs(t, err, errBranchAwareSyncRetry)
	require.Equal(t, 1, headerRequests)
	require.Equal(t, 1, proofRequests)
}

func TestBaseBootstrap_GetNextHeaderDoesNotRequestKnownLosingSuffix(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	offBranchHash := []byte("C")
	offBranchHeader, offBranchProof := createBranchAwareHeader(11, offBranchHash, []byte("B"))
	fixture.headers[string(offBranchHash)] = offBranchHeader
	fixture.proofs = []data.HeaderProofHandler{offBranchProof}
	fixture.boot.forkDetector = &mock.ForkDetectorMock{
		ProbableHighestNonceCalled: func() uint64 {
			return currentHeader.GetNonce()
		},
	}

	numRequests := 0
	fixture.boot.blockBootstrapper = &blockBootstrapperStub{
		requestHeaderByNonceCalled: func(_ uint64) { numRequests++ },
		requestProofByNonceCalled:  func(_ uint64) { numRequests++ },
	}

	_, _, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.ErrorIs(t, err, errBranchAwareSyncRetry)
	require.Zero(t, numRequests)
}

func TestBranchAwareProbableAndSyncSelectionFollowFinalBranch(t *testing.T) {
	t.Parallel()

	currentHash := []byte("A")
	currentHeader, _ := createBranchAwareHeader(10, currentHash, []byte("parent"))
	fixture := newBranchAwareSyncFixture(currentHeader, currentHash)

	bfd := newBranchAwareForkDetector(0, currentHeader.GetNonce(), currentHash)
	sfd := &shardForkDetector{baseForkDetector: bfd}
	bfd.forkDetector = sfd
	fixture.boot.forkDetector = sfd

	offBranchHash := []byte("C")
	offBranchHeader, offBranchProof := createBranchAwareHeader(11, offBranchHash, []byte("B"))
	fixture.headers[string(offBranchHash)] = offBranchHeader
	fixture.proofs = []data.HeaderProofHandler{offBranchProof}
	addProvenHeaderInfo(bfd, 10, []byte("B"), []byte("parent"), process.BHReceived)
	addProvenHeaderInfo(bfd, 11, offBranchHash, offBranchHeader.GetPrevHash(), process.BHReceived)
	bfd.setProbableHighestNonce(bfd.computeProbableHighestNonce())
	require.Equal(t, currentHeader.GetNonce(), sfd.ProbableHighestNonce())

	_, _, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.ErrorIs(t, err, errBranchAwareSyncRetry)

	canonicalHash := []byte("D")
	canonicalHeader, canonicalProof := createBranchAwareHeader(11, canonicalHash, currentHash)
	fixture.headers[string(canonicalHash)] = canonicalHeader
	fixture.proofs = append(fixture.proofs, canonicalProof)
	addProvenHeaderInfo(bfd, 11, canonicalHash, currentHash, process.BHReceived)
	bfd.setProbableHighestNonce(bfd.computeProbableHighestNonce())
	require.Equal(t, uint64(11), sfd.ProbableHighestNonce())

	header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
	require.NoError(t, err)
	require.Same(t, canonicalHeader, header)
	require.Equal(t, canonicalHash, hash)
}

func TestBranchAwareRecoveryConvergesAcrossEvidenceOrders(t *testing.T) {
	t.Parallel()

	parentHash := []byte("P")
	preferredHash := []byte("A")
	laterSiblingHash := []byte("B")
	losingChildHash := []byte("C")
	canonicalChildHash := []byte("D")

	preferredHeader := &block.HeaderV3{Nonce: 11, Round: 11, Epoch: 1, ShardID: 0, PrevHash: parentHash}
	laterSiblingHeader := &block.HeaderV3{Nonce: 11, Round: 12, Epoch: 1, ShardID: 0, PrevHash: parentHash}
	losingChildHeader := &block.HeaderV3{Nonce: 12, Round: 13, Epoch: 1, ShardID: 0, PrevHash: laterSiblingHash}
	canonicalChildHeader := &block.HeaderV3{Nonce: 12, Round: 14, Epoch: 1, ShardID: 0, PrevHash: preferredHash}

	orders := []struct {
		name       string
		operations []string
	}{
		{
			name:       "observed order",
			operations: []string{"proof B", "header B", "proof C", "header C", "settle A"},
		},
		{
			name:       "proofs before headers",
			operations: []string{"proof B", "proof C", "settle A", "header C", "header B"},
		},
		{
			name:       "headers before proofs",
			operations: []string{"header C", "header B", "settle A", "proof B", "proof C"},
		},
		{
			name:       "settlement first",
			operations: []string{"settle A", "proof C", "header C", "proof B", "header B"},
		},
	}

	for _, testCase := range orders {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			knownProofs := map[string]bool{string(preferredHash): true}
			bfd := newBranchAwareForkDetector(0, 10, parentHash)
			bfd.roundHandler = &testscommon.RoundHandlerMock{
				IndexCalled: func() int64 {
					return int64(preferredHeader.Round)
				},
			}
			bfd.processConfigsHandler = testscommon.GetDefaultProcessConfigsHandler()
			bfd.fork.rollBackNonce = math.MaxUint64
			bfd.fork.settledCheckpoint = &checkpointInfo{nonce: 10, round: 10, hash: parentHash}
			bfd.fork.checkpoint = []*checkpointInfo{
				{nonce: 10, round: 10, hash: parentHash},
				{nonce: preferredHeader.Nonce, round: preferredHeader.Round, hash: preferredHash},
			}
			bfd.proofsPool = &testscommonDataRetriever.ProofsPoolMock{
				HasProofCalled: func(_ uint32, hash []byte) bool {
					return knownProofs[string(hash)]
				},
			}
			sfd := &shardForkDetector{baseForkDetector: bfd}
			bfd.forkDetector = sfd

			appendInfo := func(header data.HeaderHandler, hash []byte, hasProof bool, state process.BlockHeaderState) {
				bfd.appendHeaderInfo(&headerInfo{
					epoch:    header.GetEpoch(),
					nonce:    header.GetNonce(),
					round:    header.GetRound(),
					hash:     hash,
					prevHash: header.GetPrevHash(),
					state:    state,
					hasProof: hasProof,
				})
				bfd.setProbableHighestNonce(bfd.computeProbableHighestNonce())
			}
			appendProof := func(header data.HeaderHandler, hash []byte) {
				knownProofs[string(hash)] = true
				bfd.appendHeaderInfo(&headerInfo{
					epoch:    header.GetEpoch(),
					nonce:    header.GetNonce(),
					round:    header.GetRound(),
					hash:     hash,
					state:    process.BHReceived,
					hasProof: true,
				})
				bfd.setProbableHighestNonce(bfd.computeProbableHighestNonce())
			}

			appendInfo(preferredHeader, preferredHash, true, process.BHReceived)
			appendInfo(preferredHeader, preferredHash, true, process.BHProcessed)

			for _, operation := range testCase.operations {
				switch operation {
				case "proof B":
					appendProof(laterSiblingHeader, laterSiblingHash)
				case "header B":
					appendInfo(laterSiblingHeader, laterSiblingHash, knownProofs[string(laterSiblingHash)], process.BHReceived)
				case "proof C":
					appendProof(losingChildHeader, losingChildHash)
				case "header C":
					appendInfo(losingChildHeader, losingChildHash, knownProofs[string(losingChildHash)], process.BHReceived)
				case "settle A":
					sfd.ReceivedSelfNotarizedFromCrossHeaders(
						core.MetachainShardId,
						[]data.HeaderHandler{preferredHeader},
						[][]byte{preferredHash},
					)
				default:
					require.FailNow(t, "unknown operation", operation)
				}
			}

			require.Equal(t, preferredHeader.Nonce, sfd.ProbableHighestNonce())
			selection := sfd.getNotarizedHeaderSelection(preferredHeader.Nonce)
			require.Equal(t, preferredHash, selection.hash)
			require.True(t, selection.isV3)
			require.Empty(t, selection.candidates)
			forkInfo := sfd.CheckFork()
			require.False(t, forkInfo.IsDetected, "fork info: %+v; final nonce: %d", forkInfo, bfd.finalCheckpoint().nonce)

			losingChildRetained := false
			for _, info := range bfd.headers[losingChildHeader.Nonce] {
				if bytes.Equal(info.hash, losingChildHash) {
					losingChildRetained = true
					break
				}
			}
			require.True(t, losingChildRetained)

			fixture := newBranchAwareSyncFixture(preferredHeader, preferredHash)
			fixture.boot.forkDetector = sfd
			fixture.headers[string(losingChildHash)] = losingChildHeader
			_, losingChildProof := createBranchAwareHeader(losingChildHeader.Nonce, losingChildHash, laterSiblingHash)
			fixture.proofs = []data.HeaderProofHandler{losingChildProof}
			require.Equal(t, preferredHeader.Nonce, sfd.ProbableHighestNonce())

			appendProof(canonicalChildHeader, canonicalChildHash)
			appendInfo(canonicalChildHeader, canonicalChildHash, true, process.BHReceived)
			require.Equal(t, canonicalChildHeader.Nonce, sfd.ProbableHighestNonce())

			_, canonicalChildProof := createBranchAwareHeader(canonicalChildHeader.Nonce, canonicalChildHash, preferredHash)
			fixture.headers[string(canonicalChildHash)] = canonicalChildHeader
			fixture.proofs = []data.HeaderProofHandler{losingChildProof, canonicalChildProof}
			header, hash, err := fixture.boot.getNextHeaderRequestingIfMissing()
			require.NoError(t, err)
			require.Same(t, canonicalChildHeader, header)
			require.Equal(t, canonicalChildHash, hash)
			appendInfo(canonicalChildHeader, canonicalChildHash, true, process.BHProcessed)

			appendProof(laterSiblingHeader, laterSiblingHash)
			appendInfo(losingChildHeader, losingChildHash, true, process.BHReceived)
			sfd.ReceivedSelfNotarizedFromCrossHeaders(
				core.MetachainShardId,
				[]data.HeaderHandler{preferredHeader},
				[][]byte{preferredHash},
			)
			require.Equal(t, canonicalChildHeader.Nonce, sfd.ProbableHighestNonce())
			require.False(t, sfd.CheckFork().IsDetected)
		})
	}
}

func TestBaseBootstrap_DoJobOnBranchAwareSyncRetryHasNoFailureSideEffects(t *testing.T) {
	t.Parallel()

	boot := &baseBootstrap{
		mapNonceSyncedWithErrors: make(map[uint64]uint32),
	}

	boot.doJobOnSyncBlockFail(nil, nil, errBranchAwareSyncRetry)
	require.Empty(t, boot.mapNonceSyncedWithErrors)
}

func TestBaseBootstrap_ProcessReceivedProofEnrichesV3AncestryAfterRecordingProof(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")
	header, proof := createBranchAwareHeader(11, hash, []byte("parent"))
	callOrder := make([]string, 0, 2)
	boot := &baseBootstrap{
		headers: &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(requestedHash []byte) (data.HeaderHandler, error) {
				require.Equal(t, hash, requestedHash)
				return header, nil
			},
		},
		forkDetector: &mock.ForkDetectorMock{
			ReceivedProofCalled: func(receivedProof data.HeaderProofHandler) {
				require.Same(t, proof, receivedProof)
				callOrder = append(callOrder, "proof")
			},
			AddHeaderCalled: func(receivedHeader data.HeaderHandler, receivedHash []byte, state process.BlockHeaderState, _ []data.HeaderHandler, _ [][]byte) error {
				require.Same(t, header, receivedHeader)
				require.Equal(t, hash, receivedHash)
				require.Equal(t, process.BHReceived, state)
				callOrder = append(callOrder, "header")
				return nil
			},
		},
		shardCoordinator: mock.NewOneShardCoordinatorMock(),
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
				return flag == common.SupernovaFlag
			},
		},
		enableRoundsHandler: &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, _ uint64) bool {
				return flag == common.SupernovaRoundFlag
			},
		},
	}

	boot.processReceivedProof(proof)
	require.Equal(t, []string{"proof", "header"}, callOrder)
}

func TestBaseBootstrap_ProcessReceivedMetaProofEnrichesV3AncestryAfterRecordingProof(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")
	header, proof := createBranchAwareMetaHeader(11, hash, []byte("parent"))
	callOrder := make([]string, 0, 2)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)
	boot := &baseBootstrap{
		headers: &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(requestedHash []byte) (data.HeaderHandler, error) {
				require.Equal(t, hash, requestedHash)
				return header, nil
			},
		},
		forkDetector: &mock.ForkDetectorMock{
			ReceivedProofCalled: func(receivedProof data.HeaderProofHandler) {
				require.Same(t, proof, receivedProof)
				callOrder = append(callOrder, "proof")
			},
			AddHeaderCalled: func(receivedHeader data.HeaderHandler, receivedHash []byte, state process.BlockHeaderState, _ []data.HeaderHandler, _ [][]byte) error {
				require.Same(t, header, receivedHeader)
				require.Equal(t, hash, receivedHash)
				require.Equal(t, process.BHReceived, state)
				callOrder = append(callOrder, "header")
				return nil
			},
		},
		shardCoordinator: shardCoordinator,
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
				return flag == common.SupernovaFlag
			},
		},
		enableRoundsHandler: &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, _ uint64) bool {
				return flag == common.SupernovaRoundFlag
			},
		},
	}

	boot.processReceivedProof(proof)
	require.Equal(t, []string{"proof", "header"}, callOrder)
}

func TestBaseBootstrap_ProcessReceivedMetaProofBeforeHeaderPreservesEvidence(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")
	header, proof := createBranchAwareMetaHeader(11, hash, []byte("parent"))
	headerAvailable := false
	numReceivedProofs := 0
	numAddedHeaders := 0
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)
	boot := &baseBootstrap{
		headers: &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(requestedHash []byte) (data.HeaderHandler, error) {
				require.Equal(t, hash, requestedHash)
				if !headerAvailable {
					return nil, errors.New("header not found")
				}

				return header, nil
			},
		},
		forkDetector: &mock.ForkDetectorMock{
			ProbableHighestNonceCalled: func() uint64 {
				return header.GetNonce()
			},
			ReceivedProofCalled: func(receivedProof data.HeaderProofHandler) {
				require.Same(t, proof, receivedProof)
				numReceivedProofs++
			},
			AddHeaderCalled: func(receivedHeader data.HeaderHandler, receivedHash []byte, state process.BlockHeaderState, _ []data.HeaderHandler, _ [][]byte) error {
				require.Same(t, header, receivedHeader)
				require.Equal(t, hash, receivedHash)
				require.Equal(t, process.BHReceived, state)
				numAddedHeaders++
				return nil
			},
		},
		shardCoordinator: shardCoordinator,
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
				return flag == common.SupernovaFlag
			},
		},
		enableRoundsHandler: &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, _ uint64) bool {
				return flag == common.SupernovaRoundFlag
			},
		},
		requestMiniBlocks: func(data.HeaderHandler) {},
	}

	boot.processReceivedProof(proof)
	require.Equal(t, 1, numReceivedProofs)
	require.Zero(t, numAddedHeaders)

	headerAvailable = true
	boot.processReceivedHeader(header, hash)
	require.Equal(t, 1, numReceivedProofs)
	require.Equal(t, 1, numAddedHeaders)
}

func TestBaseBootstrap_EnrichForkDetectorWithProofHeaderSkipsNonV3Proofs(t *testing.T) {
	t.Parallel()

	numHeaderLookups := 0
	boot := &baseBootstrap{
		headers: &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
				numHeaderLookups++
				return nil, errors.New("unexpected lookup")
			},
		},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		enableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
	}

	boot.enrichForkDetectorWithProofHeader(&block.HeaderProof{HeaderShardId: 0, HeaderHash: []byte("legacy")})
	boot.enrichForkDetectorWithProofHeader(&block.HeaderProof{HeaderShardId: core.MetachainShardId, HeaderHash: []byte("meta")})
	require.Zero(t, numHeaderLookups)
}
