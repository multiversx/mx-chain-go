package spos

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	commonConsensus "github.com/multiversx/mx-chain-go/common/consensus"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
)

var _ consensus.SubroundHandler = (*Subround)(nil)

const (
	singleKeyStartMsg = " (my turn)"
	multiKeyStartMsg  = " (my turn in multi-key)"
)

// Subround struct contains the needed data for one Subround and the Subround properties. It defines a Subround
// with its properties (its ID, next Subround ID, its duration, its name) and also it has some handler functions
// which should be set. Job function will be the main function of this Subround, Extend function will handle the overtime
// situation of the Subround and Check function will decide if in this Subround the consensus is achieved
type Subround struct {
	ConsensusCoreHandler
	ConsensusStateHandler

	previous                        int
	current                         int
	next                            int
	mutDuration                     sync.RWMutex
	baseDuration                    time.Duration
	startTimePercent                float64
	endTimePercent                  float64
	signatureSubroundEndTimePercent float64
	processingThresholdPercent      int
	startTime                       int64
	endTime                         int64
	name                            string
	chainID                         []byte
	currentPid                      core.PeerID

	consensusStateChangedChannel chan bool
	executeStoredMessages        func()
	appStatusHandler             core.AppStatusHandler

	Job    func(ctx context.Context) bool // method does the Subround Job and send the result to the peers
	Check  func() bool                    // method checks if the consensus of the Subround is done
	Extend func(subroundId int)           // method is called when round time is out
}

// NewSubround creates a new SubroundId object
func NewSubround(
	previous int,
	current int,
	next int,
	baseDuration time.Duration,
	startTimePercent float64,
	endTimePercent float64,
	name string,
	consensusState ConsensusStateHandler,
	consensusStateChangedChannel chan bool,
	executeStoredMessages func(),
	container ConsensusCoreHandler,
	chainID []byte,
	currentPid core.PeerID,
	appStatusHandler core.AppStatusHandler,
) (*Subround, error) {
	err := checkNewSubroundParams(
		consensusState,
		consensusStateChangedChannel,
		executeStoredMessages,
		container,
		chainID,
		appStatusHandler,
	)
	if err != nil {
		return nil, err
	}

	startTime, endTime := computeStartAndEndTime(baseDuration, startTimePercent, endTimePercent)
	sr := Subround{
		ConsensusCoreHandler:         container,
		ConsensusStateHandler:        consensusState,
		previous:                     previous,
		current:                      current,
		next:                         next,
		baseDuration:                 baseDuration,
		startTimePercent:             startTimePercent,
		endTimePercent:               endTimePercent,
		startTime:                    startTime,
		endTime:                      endTime,
		name:                         name,
		chainID:                      chainID,
		consensusStateChangedChannel: consensusStateChangedChannel,
		executeStoredMessages:        executeStoredMessages,
		Job:                          nil,
		Check:                        nil,
		Extend:                       nil,
		appStatusHandler:             appStatusHandler,
		currentPid:                   currentPid,
	}

	return &sr, nil
}

func checkNewSubroundParams(
	state ConsensusStateHandler,
	consensusStateChangedChannel chan bool,
	executeStoredMessages func(),
	container ConsensusCoreHandler,
	chainID []byte,
	appStatusHandler core.AppStatusHandler,
) error {
	err := ValidateConsensusCore(container)
	if err != nil {
		return err
	}
	if consensusStateChangedChannel == nil {
		return ErrNilChannel
	}
	if state == nil {
		return ErrNilConsensusState
	}
	if executeStoredMessages == nil {
		return ErrNilExecuteStoredMessages
	}
	if len(chainID) == 0 {
		return ErrInvalidChainID
	}
	if check.IfNil(appStatusHandler) {
		return ErrNilAppStatusHandler
	}

	return nil
}

func computeStartAndEndTime(baseDuration time.Duration, startTimePercent float64, endTimePercent float64) (int64, int64) {
	startTime := int64(float64(baseDuration) * startTimePercent)
	endTime := int64(float64(baseDuration) * endTimePercent)
	return startTime, endTime
}

// DoWork method actually does the work of this Subround. First it tries to do the Job of the Subround then it will
// Check the consensus. If the upper time limit of this Subround is reached, the Extend method will be called before
// returning. If this method returns true the chronology will advance to the next Subround.
func (sr *Subround) DoWork(ctx context.Context, roundHandler consensus.RoundHandler) bool {
	if sr.Job == nil || sr.Check == nil {
		return false
	}

	// execute stored messages which were received in this new round but before this initialisation
	go sr.executeStoredMessages()

	startTime := roundHandler.TimeStamp()
	maxTime := roundHandler.TimeDuration() * MaxThresholdPercent / 100

	sr.Job(ctx)
	if sr.Check() {
		return true
	}

	for {
		select {
		case <-ctx.Done():
			sr.SetRoundCanceled(true)
			return false
		case <-sr.consensusStateChangedChannel:
			if sr.Check() {
				return true
			}
		case <-time.After(roundHandler.RemainingTime(startTime, maxTime)):
			if sr.Extend != nil {
				sr.SetRoundCanceled(true)
				sr.Extend(sr.current)
			}

			return false
		}
	}
}

// Previous method returns the ID of the previous Subround
func (sr *Subround) Previous() int {
	return sr.previous
}

// Current method returns the ID of the current Subround
func (sr *Subround) Current() int {
	return sr.current
}

// Next method returns the ID of the next Subround
func (sr *Subround) Next() int {
	return sr.next
}

// StartTime method returns the start time of the Subround
func (sr *Subround) StartTime() int64 {
	sr.mutDuration.RLock()
	defer sr.mutDuration.RUnlock()

	return sr.startTime
}

// EndTime method returns the upper time limit of the Subround
func (sr *Subround) EndTime() int64 {
	sr.mutDuration.RLock()
	defer sr.mutDuration.RUnlock()

	return sr.endTime
}

// SignatureSubroundEndTime returns the end time limit of the signature subround
func (sr *Subround) SignatureSubroundEndTime() time.Duration {
	sr.mutDuration.RLock()
	defer sr.mutDuration.RUnlock()

	return time.Duration(float64(sr.baseDuration) * sr.signatureSubroundEndTimePercent)
}

// SetSignatureSubroundEndTimePercentage sets the end time percent of the signature subround
func (sr *Subround) SetSignatureSubroundEndTimePercentage(percent float64) {
	sr.mutDuration.Lock()
	defer sr.mutDuration.Unlock()

	sr.signatureSubroundEndTimePercent = percent
}

// ProcessingThresholdPercent returns the processing threshold percent of the subround
func (sr *Subround) ProcessingThresholdPercent() int {
	sr.mutDuration.RLock()
	defer sr.mutDuration.RUnlock()

	return sr.processingThresholdPercent
}

// SetProcessingThresholdPercent sets the processing threshold percent of the subround
func (sr *Subround) SetProcessingThresholdPercent(percent int) {
	sr.mutDuration.Lock()
	defer sr.mutDuration.Unlock()

	sr.processingThresholdPercent = percent
}

// SetBaseDuration sets the base duration of the subround
func (sr *Subround) SetBaseDuration(baseDuration time.Duration) {
	sr.mutDuration.Lock()
	defer sr.mutDuration.Unlock()

	sr.baseDuration = baseDuration
	sr.startTime, sr.endTime = computeStartAndEndTime(baseDuration, sr.startTimePercent, sr.endTimePercent)
}

// SetTimingPercentage sets the start time and end time percent of the subround and recomputes its start and end time
func (sr *Subround) SetTimingPercentage(startTimePercent float64, endTimePercent float64) {
	sr.mutDuration.Lock()
	defer sr.mutDuration.Unlock()

	sr.startTimePercent = startTimePercent
	sr.endTimePercent = endTimePercent
	sr.startTime, sr.endTime = computeStartAndEndTime(sr.baseDuration, sr.startTimePercent, sr.endTimePercent)
}

// Name method returns the name of the Subround
func (sr *Subround) Name() string {
	return sr.name
}

// ChainID method returns the current chain ID
func (sr *Subround) ChainID() []byte {
	return sr.chainID
}

// CurrentPid returns the current p2p peer ID
func (sr *Subround) CurrentPid() core.PeerID {
	return sr.currentPid
}

// AppStatusHandler method returns the appStatusHandler instance
func (sr *Subround) AppStatusHandler() core.AppStatusHandler {
	return sr.appStatusHandler
}

// ConsensusChannel method returns the consensus channel
func (sr *Subround) ConsensusChannel() chan bool {
	return sr.consensusStateChangedChannel
}

// HasProofForCompetingBlock checks if there is a proof for a competing block in the equivalent proofs pool
func (sr *Subround) HasProofForCompetingBlock() bool {
	currentBlock, currentBlockHash := sr.Blockchain().GetCurrentBlockHeaderAndHash()
	if check.IfNil(currentBlock) {
		return false
	}
	competingBlockNonce := currentBlock.GetNonce() + 1
	proof, err := sr.EquivalentProofsPool().GetProofByNonce(competingBlockNonce, sr.ShardCoordinator().SelfId())
	if err != nil || check.IfNil(proof) {
		return false
	}

	consensusBlockHash := sr.GetData()
	if !sr.usesV3CompetingBlockRules() {
		return len(consensusBlockHash) == 0 || !bytes.Equal(proof.GetHeaderHash(), consensusBlockHash)
	}

	// proof for current consensus block does not count as competing
	if len(consensusBlockHash) > 0 && bytes.Equal(proof.GetHeaderHash(), consensusBlockHash) {
		return false
	}

	proofExtendsCurrentBlock, ancestryKnown := sr.proofExtendsBlock(proof, currentBlockHash)
	if !ancestryKnown || proofExtendsCurrentBlock {
		return true
	}

	proofs, err := sr.EquivalentProofsPool().GetProofsByNonce(competingBlockNonce, sr.ShardCoordinator().SelfId())
	if err != nil {
		return true
	}

	for _, candidateProof := range proofs {
		if check.IfNil(candidateProof) {
			return true
		}
		if bytes.Equal(candidateProof.GetHeaderHash(), proof.GetHeaderHash()) {
			continue
		}
		if bytes.Equal(candidateProof.GetHeaderHash(), consensusBlockHash) {
			continue
		}

		proofExtendsCurrentBlock, ancestryKnown = sr.proofExtendsBlock(candidateProof, currentBlockHash)
		if !ancestryKnown || proofExtendsCurrentBlock {
			return true
		}
	}

	return false
}

func (sr *Subround) usesV3CompetingBlockRules() bool {
	header := sr.GetHeader()
	if !check.IfNil(header) {
		return common.IsAsyncExecutionEnabledForEpochAndRound(
			sr.EnableEpochsHandler(),
			sr.EnableRoundsHandler(),
			header.GetEpoch(),
			header.GetRound(),
		)
	}

	round := sr.RoundHandler().Index()
	if round < 0 {
		return false
	}

	return sr.EnableEpochsHandler().IsFlagEnabled(common.SupernovaFlag) &&
		sr.EnableRoundsHandler().IsFlagEnabledInRound(common.SupernovaRoundFlag, uint64(round))
}

// ShouldRefuseCompetingParent reports whether an unproven meta candidate extends a superseded parent.
func (sr *Subround) ShouldRefuseCompetingParent(epoch uint32, round uint64) bool {
	if sr.ShardCoordinator().SelfId() != core.MetachainShardId {
		return false
	}
	if !common.IsAsyncExecutionEnabledForEpochAndRound(
		sr.EnableEpochsHandler(),
		sr.EnableRoundsHandler(),
		epoch,
		round,
	) {
		return false
	}

	return sr.HasProofForCompetingParent()
}

// ShouldRefuseCompetingParentInCurrentEpoch applies the parent gate before a proposal header exists.
func (sr *Subround) ShouldRefuseCompetingParentInCurrentEpoch(round uint64) bool {
	if sr.ShardCoordinator().SelfId() != core.MetachainShardId {
		return false
	}
	if !sr.EnableEpochsHandler().IsFlagEnabled(common.SupernovaFlag) ||
		!sr.EnableRoundsHandler().IsFlagEnabledInRound(common.SupernovaRoundFlag, round) {
		return false
	}

	return sr.HasProofForCompetingParent()
}

func (sr *Subround) proofExtendsBlock(proof data.HeaderProofHandler, parentHash []byte) (bool, bool) {
	if len(parentHash) == 0 {
		return false, false
	}

	header, err := sr.HeadersPool().GetHeaderByHash(proof.GetHeaderHash())
	if err != nil || check.IfNil(header) {
		return false, false
	}

	return bytes.Equal(header.GetPrevHash(), parentHash), true
}

// HasProofForCompetingParent returns true if an actionable proof outranks the current header
func (sr *Subround) HasProofForCompetingParent() bool {
	currentBlock, currentHash := sr.Blockchain().GetCurrentBlockHeaderAndHash()
	if check.IfNil(currentBlock) {
		return false
	}

	preferredProof, err := sr.EquivalentProofsPool().GetProofByNonce(currentBlock.GetNonce(), sr.ShardCoordinator().SelfId())
	if err != nil || check.IfNil(preferredProof) {
		return false
	}
	if !currentBlock.IsHeaderV3() {
		return preferredProof.GetHeaderRound() < currentBlock.GetRound()
	}

	if !proofRanksBeforeHeader(preferredProof, currentBlock, currentHash) {
		return false
	}

	extendsCurrentParent, ancestryKnown := sr.proofExtendsBlock(preferredProof, currentBlock.GetPrevHash())
	if !ancestryKnown || extendsCurrentParent {
		return true
	}

	proofs, err := sr.EquivalentProofsPool().GetProofsByNonce(currentBlock.GetNonce(), sr.ShardCoordinator().SelfId())
	if err != nil {
		return true
	}

	for _, proof := range proofs {
		if check.IfNil(proof) {
			return true
		}
		if !proofRanksBeforeHeader(proof, currentBlock, currentHash) {
			continue
		}

		extendsCurrentParent, ancestryKnown = sr.proofExtendsBlock(proof, currentBlock.GetPrevHash())
		if !ancestryKnown || extendsCurrentParent {
			return true
		}
	}

	return false
}

func proofRanksBeforeHeader(proof data.HeaderProofHandler, header data.HeaderHandler, headerHash []byte) bool {
	if proof.GetHeaderRound() != header.GetRound() {
		return proof.GetHeaderRound() < header.GetRound()
	}

	return bytes.Compare(proof.GetHeaderHash(), headerHash) < 0
}

// GetAssociatedPid returns the associated PeerID to the provided public key bytes
func (sr *Subround) GetAssociatedPid(pkBytes []byte) core.PeerID {
	return sr.GetKeysHandler().GetAssociatedPid(pkBytes)
}

// IsSelfInConsensusGroup returns true is the current node is in consensus group in single
// key or in multi-key mode
func (sr *Subround) IsSelfInConsensusGroup() bool {
	return sr.IsNodeInConsensusGroup(sr.SelfPubKey()) || sr.IsMultiKeyInConsensusGroup()
}

// IsSelfLeader returns true is the current node is leader is single key or in
// multi-key mode
func (sr *Subround) IsSelfLeader() bool {
	return sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound()
}

// IsSelfLeaderInCurrentRound method checks if the current node is leader in the current round
func (sr *Subround) IsSelfLeaderInCurrentRound() bool {
	return sr.IsNodeLeaderInCurrentRound(sr.SelfPubKey()) && commonConsensus.ShouldConsiderSelfKeyInConsensus(sr.NodeRedundancyHandler())
}

// GetLeaderStartRoundMessage returns the leader start round message based on single key
// or multi-key node type
func (sr *Subround) GetLeaderStartRoundMessage() string {
	if sr.IsMultiKeyLeaderInCurrentRound() {
		return multiKeyStartMsg
	}
	if sr.IsSelfLeaderInCurrentRound() {
		return singleKeyStartMsg
	}

	return ""
}

// GetUnixTimestampForHeader returns unix timestamp in seconds or milliseconds based on epoch
func (sr *Subround) GetUnixTimestampForHeader(
	headerEpoch uint32,
) uint64 {
	if sr.EnableEpochsHandler().IsFlagEnabledInEpoch(common.SupernovaFlag, headerEpoch) {
		return uint64(sr.RoundHandler().TimeStamp().UnixMilli())
	}

	return uint64(sr.RoundHandler().TimeStamp().Unix())
}

// IsInterfaceNil returns true if there is no value under the interface
func (sr *Subround) IsInterfaceNil() bool {
	return sr == nil
}
