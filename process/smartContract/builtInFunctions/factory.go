package builtInFunctions

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/mitchellh/mapstructure"
)

// claimDeveloperRewardsFunctionName is a constant which defines the name for the claim developer rewards function
const claimDeveloperRewardsFunctionName = "ClaimDeveloperRewards"

// changeOwnerAddressFunctionName is a constant which defines the name for the change owner address function
const changeOwnerAddressFunctionName = "ChangeOwnerAddress"

const setUserName = "SetUserName"

// ArgsCreateBuiltInFunctionContainer -
type ArgsCreateBuiltInFunctionContainer struct {
	GasMap               map[string]map[string]uint64
	MapDNSAddresses      map[string]struct{}
	EnableUserNameChange bool
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
	err = container.Add(claimDeveloperRewardsFunctionName, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc = NewChangeOwnerAddressFunc(gasConfig.BuiltInCost.ChangeOwnerAddress)
	err = container.Add(changeOwnerAddressFunctionName, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewSaveUserNameFunc(gasConfig.BuiltInCost.SaveUserName, args.MapDNSAddresses, args.EnableUserNameChange)
	if err != nil {
		return nil, err
	}
	err = container.Add(setUserName, newFunc)
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
