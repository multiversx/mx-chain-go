package sync

import (
	"bytes"
	"errors"
	"testing"

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
)

type branchAwareSyncFixture struct {
	boot          *baseBootstrap
	headers       map[string]data.HeaderHandler
	proofs        []data.HeaderProofHandler
	notarizedHash []byte
}

func newBranchAwareSyncFixture(currentHeader data.HeaderHandler, currentHash []byte) *branchAwareSyncFixture {
	fixture := &branchAwareSyncFixture{
		headers: make(map[string]data.HeaderHandler),
	}
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
		shardCoordinator: mock.NewOneShardCoordinatorMock(),
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
