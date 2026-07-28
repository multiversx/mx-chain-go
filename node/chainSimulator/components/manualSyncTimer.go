package components

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/ntp"
)

var _ ntp.SyncTimer = (*manualSyncTimer)(nil)

// simulatedConsensusMaxWait bounds wall-clock timers armed by SPoS while the chronology is driven
// with a manual clock. A full simulator starts dozens of validator nodes; at an Andromeda boundary
// all consensus groups verify and propagate equivalent proofs concurrently, and 100 ms is too short
// under that CPU load (especially with DEBUG logging). Prematurely expiring END_ROUND creates an
// empty round and can shift the epoch boundary onto the Supernova duration switch. Non-participants
// still skip the wait entirely, while a successful participant returns as soon as the proof arrives.
const simulatedConsensusMaxWait = time.Second

type boundedWaitRoundHandler struct {
	consensus.RoundHandler

	mutParticipation     sync.RWMutex
	participationKnown   bool
	consensusParticipant bool
	waitInterrupted      bool
}

func (handler *boundedWaitRoundHandler) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	handler.mutParticipation.RLock()
	shouldSkipWait := handler.participationKnown && !handler.consensusParticipant
	waitInterrupted := handler.waitInterrupted
	handler.mutParticipation.RUnlock()
	if shouldSkipWait || waitInterrupted {
		return 0
	}

	remaining := handler.RoundHandler.RemainingTime(startTime, maxTime)
	if remaining > simulatedConsensusMaxWait {
		return simulatedConsensusMaxWait
	}

	return remaining
}

func (handler *boundedWaitRoundHandler) interruptConsensusWait() {
	handler.mutParticipation.Lock()
	handler.waitInterrupted = true
	handler.mutParticipation.Unlock()
}

func (handler *boundedWaitRoundHandler) prepareConsensusStep() {
	handler.mutParticipation.Lock()
	handler.waitInterrupted = false
	handler.mutParticipation.Unlock()
}

func (handler *boundedWaitRoundHandler) RevertOneRound() {
	revertHandler, ok := handler.RoundHandler.(consensus.RoundHandlerConsensusSwitch)
	if ok {
		revertHandler.RevertOneRound()
	}
}

func (handler *boundedWaitRoundHandler) setConsensusParticipant(isParticipant bool) {
	handler.mutParticipation.Lock()
	handler.participationKnown = true
	handler.consensusParticipant = isParticipant
	handler.mutParticipation.Unlock()
}

// shouldReceiveConsensusMessage keeps consensus-topic broadcasts on the physical nodes that can
// contribute to the current round. Before START_ROUND has computed the group, every node remains a
// recipient. Direct messages and non-consensus topics are never filtered by this value.
func (handler *boundedWaitRoundHandler) shouldReceiveConsensusMessage() bool {
	handler.mutParticipation.RLock()
	defer handler.mutParticipation.RUnlock()

	return !handler.participationKnown || handler.consensusParticipant
}

func (handler *boundedWaitRoundHandler) resetConsensusParticipation() {
	handler.mutParticipation.Lock()
	handler.participationKnown = false
	handler.consensusParticipant = false
	handler.waitInterrupted = false
	handler.mutParticipation.Unlock()
}

// manualSyncTimer is a SyncTimer implementation with a test-controlled clock: the current
// time only changes through SetCurrentTime or AdvanceTime calls. It performs no NTP queries
// and spawns no goroutines, making it suitable for manually driven consensus.
type manualSyncTimer struct {
	mut         sync.RWMutex
	currentTime time.Time
}

// NewManualSyncTimer creates a manual sync timer starting at the provided time
func NewManualSyncTimer(startTime time.Time) *manualSyncTimer {
	return &manualSyncTimer{
		currentTime: startTime,
	}
}

// StartSyncingTime does nothing, the manual sync timer has no background synchronization
func (mst *manualSyncTimer) StartSyncingTime() {
}

// ForceSync does nothing, the manual sync timer has no remote clock to synchronize from
func (mst *manualSyncTimer) ForceSync() {
}

// ClockOffset returns zero, the manual sync timer has no reference clock to drift from
func (mst *manualSyncTimer) ClockOffset() time.Duration {
	return 0
}

// FormattedCurrentTime returns the string representation of the current time
func (mst *manualSyncTimer) FormattedCurrentTime() string {
	currentTime := mst.CurrentTime()

	return fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.Nanosecond())
}

// CurrentTime returns the manually controlled current time
func (mst *manualSyncTimer) CurrentTime() time.Time {
	mst.mut.RLock()
	defer mst.mut.RUnlock()

	return mst.currentTime
}

// SetCurrentTime sets the current time to the provided value
func (mst *manualSyncTimer) SetCurrentTime(t time.Time) {
	mst.mut.Lock()
	defer mst.mut.Unlock()

	mst.currentTime = t
}

// AdvanceTime moves the current time forward by the provided duration and returns the new time
func (mst *manualSyncTimer) AdvanceTime(d time.Duration) time.Time {
	mst.mut.Lock()
	defer mst.mut.Unlock()

	mst.currentTime = mst.currentTime.Add(d)

	return mst.currentTime
}

// Close does nothing, the manual sync timer has no background resources
func (mst *manualSyncTimer) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mst *manualSyncTimer) IsInterfaceNil() bool {
	return mst == nil
}
