package builtInFunctions

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/mitchellh/mapstructure"
)

var log = logger.GetOrCreate("process/smartContract/builtInFunctions")

// ArgsCreateBuiltInFunctionContainer -
type ArgsCreateBuiltInFunctionContainer struct {
	GasMap               map[string]map[string]uint64
	MapDNSAddresses      map[string]struct{}
	EnableUserNameChange bool
	Marshalizer          marshal.Marshalizer
	Accounts             state.AccountsAdapter
}

// CreateBuiltInFunctionContainer will create the list of built-in functions
func CreateBuiltInFunctionContainer(args ArgsCreateBuiltInFunctionContainer) (process.BuiltInFunctionContainer, error) {
	gasConfig, err := createGasConfig(args.GasMap)
	if err != nil {
		return nil, err
	}

	container := NewBuiltInFunctionContainer()
	var newFunc process.BuiltinFunction
	newFunc = NewClaimDeveloperRewardsFunc(gasConfig.BuiltInCost.ClaimDeveloperRewards)
	err = container.Add(core.BuiltInFunctionClaimDeveloperRewards, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc = NewChangeOwnerAddressFunc(gasConfig.BuiltInCost.ChangeOwnerAddress)
	err = container.Add(core.BuiltInFunctionChangeOwnerAddress, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewSaveUserNameFunc(gasConfig.BuiltInCost.SaveUserName, args.MapDNSAddresses, args.EnableUserNameChange)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionSetUserName, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewSaveKeyValueStorageFunc(gasConfig.BaseOperationCost, gasConfig.BuiltInCost.SaveKeyValue)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionSaveKeyValue, newFunc)
	if err != nil {
		return nil, err
	}

	pauseFunc, err := NewESDTPauseFunc(args.Accounts, true)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionESDTPause, pauseFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewESDTTransferFunc(gasConfig.BuiltInCost.ESDTTransfer, args.Marshalizer, pauseFunc)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionESDTTransfer, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewESDTBurnFunc(gasConfig.BuiltInCost.ESDTBurn, args.Marshalizer, pauseFunc)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionESDTBurn, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewESDTFreezeWipeFunc(args.Marshalizer, true, false)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionESDTFreeze, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewESDTFreezeWipeFunc(args.Marshalizer, false, false)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionESDTUnFreeze, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewESDTFreezeWipeFunc(args.Marshalizer, false, true)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionESDTWipe, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewESDTPauseFunc(args.Accounts, false)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionESDTUnPause, newFunc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func createGasConfig(gasMap map[string]map[string]uint64) (*GasCost, error) {
	baseOps := &BaseOperationCost{}
	err := mapstructure.Decode(gasMap[core.BaseOperationCost], baseOps)
	if err != nil {
		return nil, err
	}

	err = check.ForZeroUintFields(*baseOps)
	if err != nil {
		return nil, err
	}

	builtInOps := &BuiltInCost{}
	err = mapstructure.Decode(gasMap[core.BuiltInCost], builtInOps)
	if err != nil {
		return nil, err
	}

	err = check.ForZeroUintFields(*builtInOps)
	if err != nil {
		return nil, err
	}

	gasCost := GasCost{
		BaseOperationCost: *baseOps,
		BuiltInCost:       *builtInOps,
	}

	return &gasCost, nil
}

// SetPayableHandler sets the payable interface to the needed functions
func SetPayableHandler(container process.BuiltInFunctionContainer, isPayableHandler process.PayableHandler) error {
	builtInFunc, err := container.Get(core.BuiltInFunctionESDTTransfer)
	if err != nil {
		log.Warn("SetIsPayable", "error", err.Error())
		return err
	}

	esdtTransferFunc, ok := builtInFunc.(*esdtTransfer)
	if !ok {
		log.Warn("SetIsPayable", "error", process.ErrWrongTypeAssertion)
		return process.ErrWrongTypeAssertion
	}

	return esdtTransferFunc.setPayableHandler(isPayableHandler)
}
