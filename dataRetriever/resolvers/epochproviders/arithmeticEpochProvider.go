package epochproviders

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// deltaEpochActive represents how many epochs behind the current computed epoch are to be considered "active" and
// cause the requests to be sent to all peers regardless of being full observers or not. Usually, a node will have
// [config.toml].[StoragePruning].NumActivePersisters opened persisters but to the fact that a shorter epoch can happen,
// that value is lowered at a maximum 1.
const deltaEpochActive = uint32(1)
const millisecondsInOneSecond = uint64(1000)

var log = logger.GetOrCreate("resolvers/epochproviders")

// ArgArithmeticEpochProvider is the argument structure for the arithmetic epoch provider
type ArgArithmeticEpochProvider struct {
	ChainParametersHandler process.ChainParametersHandler
	StartTime              int64
	EnableEpochsHandler    common.EnableEpochsHandler
}

type arithmeticEpochProvider struct {
	sync.RWMutex
	currentComputedEpoch       uint32
	headerEpoch                uint32
	headerTimestampForNewEpoch uint64
	getUnixHandler             func() int64
	enableEpochsHandler        common.EnableEpochsHandler
	chainParamsHandler         process.ChainParametersHandler
}

// NewArithmeticEpochProvider returns a new arithmetic epoch provider able to mathematically compute the current network epoch
// based on the last block saved, considering the block's timestamp and epoch in respect with the current time
func NewArithmeticEpochProvider(arg ArgArithmeticEpochProvider) (*arithmeticEpochProvider, error) {
	if arg.StartTime < 0 {
		return nil, fmt.Errorf("%w in NewArithmeticEpochProvider", ErrInvalidStartTime)
	}
	if check.IfNil(arg.EnableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}
	if check.IfNil(arg.ChainParametersHandler) {
		return nil, process.ErrNilChainParametersHandler
	}

	aep := &arithmeticEpochProvider{
		headerEpoch:                0,
		headerTimestampForNewEpoch: uint64(arg.StartTime),
		enableEpochsHandler:        arg.EnableEpochsHandler,
		chainParamsHandler:         arg.ChainParametersHandler,
	}
	aep.getUnixHandler = func() int64 {
		if aep.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, aep.headerEpoch) {
			return time.Now().UnixMilli()
		}

		return time.Now().Unix()
	}
	aep.computeCurrentEpoch() //based on the genesis provided data

	return aep, nil
}

// EpochIsActiveInNetwork returns true if the persister for the given epoch is active in the network
func (aep *arithmeticEpochProvider) EpochIsActiveInNetwork(epoch uint32) bool {
	aep.RLock()
	defer aep.RUnlock()

	subtractWillOverflow := aep.currentComputedEpoch < epoch
	if subtractWillOverflow {
		return true
	}

	return aep.currentComputedEpoch-epoch <= deltaEpochActive
}

// EpochConfirmed is called whenever the epoch was changed and will cause the re-computation of the network's current epoch
func (aep *arithmeticEpochProvider) EpochConfirmed(newEpoch uint32, newTimestamp uint64) {
	isTimestampInvalid := newTimestamp == 0
	if isTimestampInvalid {
		return
	}

	aep.Lock()
	aep.headerEpoch = newEpoch
	aep.headerTimestampForNewEpoch = newTimestamp

	aep.computeCurrentEpoch()
	aep.Unlock()
}

func (aep *arithmeticEpochProvider) computeCurrentEpoch() {
	currentTimeStamp := uint64(aep.getUnixHandler())

	if currentTimeStamp < aep.headerTimestampForNewEpoch {
		aep.currentComputedEpoch = aep.headerEpoch
		return
	}

	currentChainParameters := aep.chainParamsHandler.CurrentChainParameters()

	diffTimeStampInMilliseconds := aep.getDiffTimeStampInMilliseconds(currentTimeStamp)
	diffRounds := diffTimeStampInMilliseconds / currentChainParameters.RoundDuration
	diffEpochs := diffRounds / uint64(currentChainParameters.RoundsPerEpoch+1)

	aep.currentComputedEpoch = aep.headerEpoch + uint32(diffEpochs)

	log.Info("arithmeticEpochProvider.computeCurrentEpoch",
		"computed network epoch", aep.currentComputedEpoch,
	)
}

func (aep *arithmeticEpochProvider) getDiffTimeStampInMilliseconds(currentTimeStamp uint64) uint64 {
	if aep.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, aep.headerEpoch) {
		return currentTimeStamp - aep.headerTimestampForNewEpoch
	}

	diffTimeStampInSeconds := currentTimeStamp - aep.headerTimestampForNewEpoch
	diffTimeStampInMilliseconds := diffTimeStampInSeconds * millisecondsInOneSecond

	return diffTimeStampInMilliseconds
}

// IsInterfaceNil returns true if there is no value under the interface
func (aep *arithmeticEpochProvider) IsInterfaceNil() bool {
	return aep == nil
}
