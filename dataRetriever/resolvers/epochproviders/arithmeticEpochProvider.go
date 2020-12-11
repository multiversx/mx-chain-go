package epochproviders

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
)

// deltaEpochActive represents how many epochs behind the current computed epoch are to be considered "active" and
//cause the requests to be sent to all peers regardless of being full observers or not. Usually, a node will have
// [config.toml].[StoragePruning].NumActivePersisters opened persisters but to the fact that a shorter epoch can happen,
// that value is lowered at a maximum 1.
const deltaEpochActive = uint32(1)
const millisecondsInOneSecond = uint64(1000)

// ArgArithmeticEpochProvider is the argument structure for the arithmetic epoch provider
type ArgArithmeticEpochProvider struct {
	RoundsPerEpoch          uint32
	RoundTimeInMilliseconds uint64
	StartTime               int64
}

type arithmeticEpochProvider struct {
	sync.RWMutex
	currentComputedEpoch       uint32
	headerEpoch                uint32
	headerTimestampForNewEpoch uint64
	roundsPerEpoch             uint32
	roundTimeInMilliseconds    uint64
	startTime                  int64
	getUnixHandler             func() int64
}

// NewArithmeticEpochProvider returns a new arithmetic epoch provider able to mathematically compute the current network epoch
// based on the last block saved, considering the block's timestamp and epoch in respect with the current time
func NewArithmeticEpochProvider(arg ArgArithmeticEpochProvider) (*arithmeticEpochProvider, error) {
	if arg.RoundsPerEpoch == 0 {
		return nil, ErrInvalidRoundsPerEpoch
	}
	if arg.RoundTimeInMilliseconds == 0 {
		return nil, ErrInvalidRoundTimeInMilliseconds
	}
	if arg.StartTime < 0 {
		return nil, ErrInvalidStartTime
	}
	aep := &arithmeticEpochProvider{
		headerEpoch:                0,
		headerTimestampForNewEpoch: uint64(arg.StartTime),
		roundsPerEpoch:             arg.RoundsPerEpoch,
		roundTimeInMilliseconds:    arg.RoundTimeInMilliseconds,
		startTime:                  arg.StartTime,
	}
	aep.getUnixHandler = func() int64 {
		return time.Now().Unix()
	}
	aep.computeCurrentEpoch() //based on the genesis provided data

	return aep, nil
}

// EpochIsActiveInNetwork returns true if the persister for the given epoch is active in the network
func (aep *arithmeticEpochProvider) EpochIsActiveInNetwork(epoch uint32) bool {
	aep.RLock()
	defer aep.RUnlock()

	if aep.currentComputedEpoch < epoch {
		return true
	}

	return aep.currentComputedEpoch-epoch <= deltaEpochActive
}

// EpochStartAction is called whenever the epoch was changed and will cause the re-computation of the current epoch
func (aep *arithmeticEpochProvider) EpochStartAction(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	aep.Lock()
	aep.headerEpoch = header.GetEpoch()
	aep.headerTimestampForNewEpoch = header.GetTimeStamp()

	aep.computeCurrentEpoch()
	aep.Unlock()
}

// EpochStartPrepare does nothing
func (aep *arithmeticEpochProvider) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyOrder will return the core.CurrentNetworkEpochProvider value
func (aep *arithmeticEpochProvider) NotifyOrder() uint32 {
	return core.CurrentNetworkEpochProvider
}

func (aep *arithmeticEpochProvider) computeCurrentEpoch() {
	currentTimeStamp := uint64(aep.getUnixHandler())

	if currentTimeStamp < aep.headerTimestampForNewEpoch {
		aep.currentComputedEpoch = aep.headerEpoch
		return
	}

	diffTimeStampInSeconds := currentTimeStamp - aep.headerTimestampForNewEpoch
	diffTimeStampInMilliseconds := diffTimeStampInSeconds * millisecondsInOneSecond
	diffRounds := diffTimeStampInMilliseconds / aep.roundTimeInMilliseconds
	diffEpochs := diffRounds / uint64(aep.roundsPerEpoch+1)

	aep.currentComputedEpoch = aep.headerEpoch + uint32(diffEpochs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (aep *arithmeticEpochProvider) IsInterfaceNil() bool {
	return aep == nil
}
