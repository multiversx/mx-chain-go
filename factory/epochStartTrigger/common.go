package epochStartTrigger

import (
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
)

func checkNilArgs(args factory.ArgsEpochStartTrigger) error {
	if check.IfNil(args.DataComps) {
		return process.ErrNilDataComponentsHolder
	}
	if check.IfNil(args.DataComps.Datapool()) {
		return process.ErrNilDataPoolHolder
	}
	if check.IfNil(args.DataComps.Blockchain()) {
		return process.ErrNilBlockChain
	}
	if check.IfNil(args.DataComps.Datapool().MiniBlocks()) {
		return dataRetriever.ErrNilMiniblocksPool
	}
	if check.IfNil(args.DataComps.Datapool().ValidatorsInfo()) {
		return process.ErrNilValidatorInfoPool
	}
	if check.IfNil(args.BootstrapComponents) {
		return process.ErrNilBootstrapComponentsHolder
	}
	if check.IfNil(args.BootstrapComponents.ShardCoordinator()) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.RequestHandler) {
		return process.ErrNilRequestHandler
	}

	return nil
}
