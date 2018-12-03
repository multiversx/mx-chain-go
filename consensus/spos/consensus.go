package spos

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// sleepTime defines the time in milliseconds between each iteration made in DoWork methods of the subrounds
const sleepTime = time.Duration(5 * time.Millisecond)

const (
	// SrStartRound defines ID of subround "Start round"
	SrStartRound chronology.SubroundId = iota
	// SrBlock defines ID of subround "block"
	SrBlock
	// SrCommitmentHash defines ID of subround "commitment hash"
	SrCommitmentHash
	// SrBitmap defines ID of subround "bitmap"
	SrBitmap
	// SrCommitment defines ID of subround "commitment"
	SrCommitment
	// SrSignature defines ID of subround "signature"
	SrSignature
	// SrEndRound defines ID of subround "End round"
	SrEndRound
)

// SubroundStatus defines the type used to refer the state of the current subround
type SubroundStatus int

const (
	// SsNotFinished defines the un-finished state of the subround
	SsNotFinished SubroundStatus = iota
	// SsExtended defines the extended state of the subround
	SsExtended
	// SsFinished defines the finished state of the subround
	SsFinished
)

// RoundStatus defines the data needed by spos to know the state of each subround in the current round
type RoundStatus struct {
	status map[chronology.SubroundId]SubroundStatus
	mut    sync.RWMutex
}

// NewRoundStatus creates a new RoundStatus object
func NewRoundStatus() *RoundStatus {
	rs := RoundStatus{}
	rs.status = make(map[chronology.SubroundId]SubroundStatus)
	return &rs
}

// ResetRoundStatus method resets the state of each subround
func (rs *RoundStatus) ResetRoundStatus() {
	for k := range rs.status {
		rs.mut.Lock()
		rs.status[k] = SsNotFinished
		rs.mut.Unlock()
	}
}

// Status returns the status of the given subround id
func (rs *RoundStatus) Status(subroundId chronology.SubroundId) SubroundStatus {
	rs.mut.RLock()
	retcode := rs.status[subroundId]
	rs.mut.RUnlock()
	return retcode
}

// SetStatus sets the status of the given subround id
func (rs *RoundStatus) SetStatus(subroundId chronology.SubroundId, subroundStatus SubroundStatus) {
	rs.mut.Lock()
	rs.status[subroundId] = subroundStatus
	rs.mut.Unlock()
}

// RoundThreshold defines the minimum agreements needed for each subround to consider the subround finished.
// (Ex: PBFT threshold has 2 / 3 + 1 agreements)
type RoundThreshold struct {
	threshold map[chronology.SubroundId]int
	mut       sync.RWMutex
}

// NewRoundThreshold creates a new RoundThreshold object
func NewRoundThreshold() *RoundThreshold {
	rt := RoundThreshold{}
	rt.threshold = make(map[chronology.SubroundId]int)
	return &rt
}

// Threshold returns the threshold of agreements needed in the given subround id
func (rt *RoundThreshold) Threshold(subroundId chronology.SubroundId) int {
	rt.mut.RLock()
	retcode := rt.threshold[subroundId]
	rt.mut.RUnlock()
	return retcode
}

// SetThreshold sets the threshold of agreements needed in the given subround id
func (rt *RoundThreshold) SetThreshold(subroundId chronology.SubroundId, threshold int) {
	rt.mut.Lock()
	rt.threshold[subroundId] = threshold
	rt.mut.Unlock()
}

// Consensus defines the data needed by spos to do the consensus in each round
type Consensus struct {
	log bool

	Data *[]byte // hold the data on which validators do the consensus
	// (could be for example a hash of the block header proposed by the leader)
	*Validators
	*RoundThreshold
	*RoundStatus

	shouldCheckConsensus bool

	Chr *chronology.Chronology
}

// NewConsensus creates a new Consensus object
func NewConsensus(
	log bool,
	data *[]byte,
	vld *Validators,
	thr *RoundThreshold,
	rs *RoundStatus,
	chr *chronology.Chronology,
) *Consensus {

	cns := Consensus{
		log:            log,
		Data:           data,
		Validators:     vld,
		RoundThreshold: thr,
		RoundStatus:    rs,
		Chr:            chr,
	}

	return &cns
}

// SetShouldCheckConsensus sets the flag which says that the consensus should be checked again, because some actions
// happens meantime which could changed its state
func (cns *Consensus) SetShouldCheckConsensus(value bool) {
	cns.shouldCheckConsensus = value
}

// ShouldCheckConsensus returns the flag which says if the consensus should be checked again
func (cns *Consensus) ShouldCheckConsensus() bool {
	return cns.shouldCheckConsensus
}

// CheckConsensus method checks if the consensus is achieved in each subround from first subround to the given
// subround. If the consensus is achieved in one subround, the subround status is marked as finished
func (cns *Consensus) CheckConsensus(subroundId chronology.SubroundId) bool {
	if subroundId == SrStartRound {
		return true
	}

	if subroundId == SrEndRound {
		cns.SetShouldCheckConsensus(true)
	}

	if !cns.ShouldCheckConsensus() {
		return false
	}

	cns.SetShouldCheckConsensus(false)

	for i := SrBlock; i <= subroundId; i++ {
		switch i {
		case SrBlock:
			if !cns.CheckBlockConsensus() {
				return false
			}
		case SrCommitmentHash:
			if !cns.CheckCommitmentHashConsensus() {
				return false
			}
		case SrBitmap:
			if !cns.CheckBitmapConsensus() {
				return false
			}
		case SrCommitment:
			if !cns.CheckCommitmentConsensus() {
				return false
			}
		case SrSignature:
			if !cns.CheckSignatureConsensus() {
				return false
			}
		}
	}

	return true
}

// CheckBlockConsensus method checks if the consensus in the <BLOCK> subround is achieved
func (cns *Consensus) CheckBlockConsensus() bool {
	if cns.Status(SrBlock) == SsFinished {
		return true
	}

	if cns.IsBlockReceived(cns.Threshold(SrBlock)) {
		cns.PrintBlockCM() // only for printing block consensus messages
		cns.SetStatus(SrBlock, SsFinished)
		return true
	}

	return false
}

// CheckCommitmentHashConsensus method checks if the consensus in the <COMMITMENT_HASH> subround is achieved
func (cns *Consensus) CheckCommitmentHashConsensus() bool {
	if cns.Status(SrCommitmentHash) == SsFinished {
		return true
	}

	threshold := cns.Threshold(SrCommitmentHash)

	if !cns.IsNodeLeaderInCurrentRound(cns.selfId) {
		threshold = len(cns.consensusGroup)
	}

	if cns.IsCommitmentHashReceived(threshold) {
		cns.PrintCommitmentHashCM() // only for printing commitment hash consensus messages
		cns.SetStatus(SrCommitmentHash, SsFinished)
		return true
	}

	if cns.CommitmentHashesCollected(cns.Threshold(SrBitmap)) {
		cns.PrintCommitmentHashCM() // only for printing commitment hash consensus messages
		cns.SetStatus(SrCommitmentHash, SsFinished)
		return true
	}

	return false
}

// CheckBitmapConsensus method checks if the consensus in the <BITMAP> subround is achieved
func (cns *Consensus) CheckBitmapConsensus() bool {
	if cns.Status(SrBitmap) == SsFinished {
		return true
	}

	if cns.CommitmentHashesCollected(cns.Threshold(SrBitmap)) {
		cns.PrintBitmapCM() // only for printing bitmap consensus messages
		cns.SetStatus(SrBitmap, SsFinished)
		return true
	}

	return false
}

// CheckCommitmentConsensus method checks if the consensus in the <COMMITMENT> subround is achieved
func (cns *Consensus) CheckCommitmentConsensus() bool {
	if cns.Status(SrCommitment) == SsFinished {
		return true
	}

	if cns.CommitmentsCollected(cns.Threshold(SrCommitment)) {
		cns.PrintCommitmentCM() // only for printing commitment consensus messages
		cns.SetStatus(SrCommitment, SsFinished)
		return true
	}

	return false
}

// CheckSignatureConsensus method checks if the consensus in the <SIGNATURE> subround is achieved
func (cns *Consensus) CheckSignatureConsensus() bool {
	if cns.Status(SrSignature) == SsFinished {
		return true
	}

	if cns.SignaturesCollected(cns.Threshold(SrSignature)) {
		cns.PrintSignatureCM() // only for printing signature consensus messages
		cns.SetStatus(SrSignature, SsFinished)
		return true
	}

	return false
}

// IsNodeLeaderInCurrentRound method checks if the node is leader in the current round
func (cns *Consensus) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := cns.GetLeader()

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	return leader == node
}

// GetLeader method gets the leader of the current round
func (cns *Consensus) GetLeader() (string, error) {
	if cns.Chr == nil {
		return "", errors.New("chronology is null")
	}

	if cns.Chr.Round() == nil {
		return "", errors.New("round is null")
	}

	if cns.Chr.Round().Index() < 0 {
		return "", errors.New("round index is negative")
	}

	if cns.consensusGroup == nil {
		return "", errors.New("consensusGroup is null")
	}

	if len(cns.consensusGroup) == 0 {
		return "", errors.New("consensusGroup is empty")
	}

	index := cns.Chr.Round().Index() % len(cns.consensusGroup)
	return cns.consensusGroup[index], nil
}

// GetSubroundName returns the name of each subround from a given subround ID
func (cns *Consensus) GetSubroundName(subroundId chronology.SubroundId) string {
	switch subroundId {
	case SrStartRound:
		return "<START_ROUND>"
	case SrBlock:
		return "<BLOCK>"
	case SrCommitmentHash:
		return "<COMMITMENT_HASH>"
	case SrBitmap:
		return "<BITMAP>"
	case SrCommitment:
		return "<COMMITMENT>"
	case SrSignature:
		return "<SIGNATURE>"
	case SrEndRound:
		return "<END_ROUND>"
	default:
		return "Undifined subround"
	}
}

// PrintBlockCM method prints the <BLOCK> consensus messages
func (cns *Consensus) PrintBlockCM() {
	if !cns.IsNodeLeaderInCurrentRound(cns.selfId) {
		cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
			"Step 1: Synchronized block"))
	}
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 1: SubroundId <BLOCK> has been finished"))
}

// PrintCommitmentHashCM method prints the <COMMITMENT_HASH> consensus messages
func (cns *Consensus) PrintCommitmentHashCM() {
	n := cns.ComputeSize(SrCommitmentHash)
	if n == len(cns.consensusGroup) {
		cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
			"Step 2: Received all (%d from %d) commitment hashes", n, len(cns.consensusGroup)))
	} else {
		cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
			"Step 2: Received %d from %d commitment hashes, which are enough", n, len(cns.consensusGroup)))
	}
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 2: SubroundId <COMMITMENT_HASH> has been finished"))
}

// PrintBitmapCM method prints the <BITMAP> consensus messages
func (cns *Consensus) PrintBitmapCM() {
	if !cns.IsNodeLeaderInCurrentRound(cns.selfId) {
		msg := fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
			"Step 3: Received bitmap from leader, matching with my own, and it got %d from %d commitment hashes, which are enough",
			cns.ComputeSize(SrBitmap), len(cns.consensusGroup))

		if cns.IsNodeInBitmapGroup(cns.selfId) {
			msg = fmt.Sprintf(msg+"%s", ", AND I WAS selected in this bitmap")
		} else {
			msg = fmt.Sprintf(msg+"%s", ", BUT I WAS NOT selected in this bitmap")
		}

		cns.Log(msg)
	}
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 3: SubroundId <BITMAP> has been finished"))
}

// PrintCommitmentCM method prints the <COMMITMENT> consensus messages
func (cns *Consensus) PrintCommitmentCM() {
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
		"Step 4: Received %d from %d commitments, which are matching with bitmap and are enough",
		cns.ComputeSize(SrCommitment), len(cns.consensusGroup)))
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 4: SubroundId <COMMITMENT> has been finished"))
}

// PrintSignatureCM method prints the <SIGNATURE> consensus messages
func (cns *Consensus) PrintSignatureCM() {
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
		"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough",
		cns.ComputeSize(SrSignature), len(cns.consensusGroup)))
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 5: SubroundId <SIGNATURE> has been finished"))
}

// Log method prints info about consensus (if log is true)
func (cns *Consensus) Log(message string) {
	if cns.log {
		fmt.Printf(message + "\n")
	}
}
