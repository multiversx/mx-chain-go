package fee

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
)

// ArgsNewFeeComputer holds the arguments for constructing a feeComputer
type ArgsNewFeeComputer struct {
	BuiltInFunctionsCostHandler economics.BuiltInFunctionsCostHandler
	EconomicsConfig             config.EconomicsConfig
	EnableEpochsConfig          config.EnableEpochs
}

func (args *ArgsNewFeeComputer) check() error {
	if check.IfNil(args.BuiltInFunctionsCostHandler) {
		return process.ErrNilBuiltInFunctionsCostHandler
	}

	return nil
}
