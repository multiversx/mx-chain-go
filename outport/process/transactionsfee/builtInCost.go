package transactionsfee

import (
	"github.com/mitchellh/mapstructure"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

// ArgsBuiltInFunctionCost holds all components that are needed to create a new instance of builtInFunctionsCost
type ArgsBuiltInFunctionCost struct {
	GasSchedule core.GasScheduleNotifier
	ArgsParser  process.ArgumentsParser
}

type builtInFunctionsCost struct {
	gasConfig *process.GasCost
}

// NewBuiltInFunctionsCost will create a new instance of builtInFunctionsCost
func NewBuiltInFunctionsCost(gasSchedule core.GasScheduleNotifier) (*builtInFunctionsCost, error) {
	if check.IfNil(gasSchedule) {
		return nil, process.ErrNilGasSchedule
	}

	bs := &builtInFunctionsCost{}

	var err error
	bs.gasConfig, err = createGasConfig(gasSchedule.LatestGasSchedule())
	if err != nil {
		return nil, err
	}

	gasSchedule.RegisterNotifyHandler(bs)

	return bs, nil
}

// GasScheduleChange is called when gas schedule is changed, thus all contracts must be updated
func (bc *builtInFunctionsCost) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	newGasConfig, err := createGasConfig(gasSchedule)
	if err != nil {
		return
	}

	bc.gasConfig = newGasConfig
}

// GetESDTTransferBuiltInCost -
func (bc *builtInFunctionsCost) GetESDTTransferBuiltInCost() uint64 {
	return bc.gasConfig.BuiltInCost.ESDTTransfer
}

// IsInterfaceNil returns true if underlying object is nil
func (bc *builtInFunctionsCost) IsInterfaceNil() bool {
	return bc == nil
}

func createGasConfig(gasMap map[string]map[string]uint64) (*process.GasCost, error) {
	baseOps := &process.BaseOperationCost{}
	err := mapstructure.Decode(gasMap[common.BaseOperationCost], baseOps)
	if err != nil {
		return nil, err
	}

	err = check.ForZeroUintFields(*baseOps)
	if err != nil {
		return nil, err
	}

	builtInOps := &process.BuiltInCost{}
	err = mapstructure.Decode(gasMap[common.BuiltInCost], builtInOps)
	if err != nil {
		return nil, err
	}

	err = check.ForZeroUintFields(*builtInOps)
	if err != nil {
		return nil, err
	}

	gasCost := process.GasCost{
		BaseOperationCost: *baseOps,
		BuiltInCost:       *builtInOps,
	}

	return &gasCost, nil
}
