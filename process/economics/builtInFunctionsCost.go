package economics

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/mitchellh/mapstructure"
)

// ArgsBuiltInFunctionCost holds all components that are needed to create a new instance of builtInFunctionsCost
type ArgsBuiltInFunctionCost struct {
	GasSchedule core.GasScheduleNotifier
	ArgsParser  process.ArgumentsParser
}

type builtInFunctionsCost struct {
	gasConfig               *process.GasCost
	specialBuiltInFunctions map[string]struct{}
	argsParser              process.ArgumentsParser
}

// NewBuiltInFunctionsCost will create a new instance of builtInFunctionsCost
func NewBuiltInFunctionsCost(args *ArgsBuiltInFunctionCost) (*builtInFunctionsCost, error) {
	if args == nil {
		return nil, process.ErrNilArgsBuiltInFunctionsConstHandler
	}
	if check.IfNil(args.ArgsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if check.IfNil(args.GasSchedule) {
		return nil, process.ErrNilGasSchedule
	}

	bs := &builtInFunctionsCost{
		argsParser: args.ArgsParser,
	}

	bs.initSpecialBuiltInFunctionCostMap()

	var err error
	bs.gasConfig, err = createGasConfig(args.GasSchedule.LatestGasSchedule())
	if err != nil {
		return nil, err
	}

	args.GasSchedule.RegisterNotifyHandler(bs)

	return bs, nil
}

func (bc *builtInFunctionsCost) initSpecialBuiltInFunctionCostMap() {
	bc.specialBuiltInFunctions = map[string]struct{}{
		core.BuiltInFunctionClaimDeveloperRewards: {},
		core.BuiltInFunctionChangeOwnerAddress:    {},
		core.BuiltInFunctionSetUserName:           {},
		core.BuiltInFunctionSaveKeyValue:          {},
		core.BuiltInFunctionESDTTransfer:          {},
		core.BuiltInFunctionESDTBurn:              {},
		core.BuiltInFunctionESDTLocalBurn:         {},
		core.BuiltInFunctionESDTLocalMint:         {},
		core.BuiltInFunctionESDTNFTAddQuantity:    {},
		core.BuiltInFunctionESDTNFTBurn:           {},
		core.BuiltInFunctionESDTNFTCreate:         {},
	}
}

// GasScheduleChange is called when gas schedule is changed, thus all contracts must be updated
func (bc *builtInFunctionsCost) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	newGasConfig, err := createGasConfig(gasSchedule)
	if err != nil {
		return
	}

	bc.gasConfig = newGasConfig
}

// ComputeBuiltInCost will compute built in function cost
func (bc *builtInFunctionsCost) ComputeBuiltInCost(tx data.TransactionWithFeeHandler) uint64 {
	function, arguments, err := bc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return 0
	}

	switch function {
	case core.BuiltInFunctionClaimDeveloperRewards:
		return bc.gasConfig.BuiltInCost.ClaimDeveloperRewards
	case core.BuiltInFunctionChangeOwnerAddress:
		return bc.gasConfig.BuiltInCost.ChangeOwnerAddress
	case core.BuiltInFunctionSetUserName:
		return bc.gasConfig.BuiltInCost.SaveUserName
	case core.BuiltInFunctionSaveKeyValue:
		return bc.gasConfig.BuiltInCost.SaveKeyValue
	case core.BuiltInFunctionESDTTransfer:
		return bc.gasConfig.BuiltInCost.ESDTTransfer
	case core.BuiltInFunctionESDTBurn:
		return bc.gasConfig.BuiltInCost.ESDTBurn
	case core.BuiltInFunctionESDTLocalBurn:
		return bc.gasConfig.BuiltInCost.ESDTLocalBurn
	case core.BuiltInFunctionESDTLocalMint:
		return bc.gasConfig.BuiltInCost.ESDTLocalMint
	case core.BuiltInFunctionESDTNFTAddQuantity:
		return bc.gasConfig.BuiltInCost.ESDTNFTAddQuantity
	case core.BuiltInFunctionESDTNFTBurn:
		return bc.gasConfig.BuiltInCost.ESDTNFTBurn
	case core.BuiltInFunctionESDTNFTCreate:
		costStorage := calculateLenOfArguments(arguments) * bc.gasConfig.BaseOperationCost.StorePerByte
		return bc.gasConfig.BuiltInCost.ESDTNFTCreate + costStorage
	default:
		return 0
	}
}

func calculateLenOfArguments(arguments [][]byte) uint64 {
	totalLen := uint64(0)
	for _, arg := range arguments {
		totalLen += uint64(len(arg))
	}

	return totalLen
}

// IsBuiltInFuncCall will check is the provided transaction is a build in function call
func (bc *builtInFunctionsCost) IsBuiltInFuncCall(tx data.TransactionWithFeeHandler) bool {
	function, arguments, err := bc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return false
	}

	_, isSpecialBuiltIn := bc.specialBuiltInFunctions[function]
	isSCCallAfter := core.IsSmartContractAddress(tx.GetRcvAddr()) && len(arguments) > core.MinLenArgumentsESDTTransfer

	return isSpecialBuiltIn && !isSCCallAfter
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
