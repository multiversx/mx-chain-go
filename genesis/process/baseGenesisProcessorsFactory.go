package process

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
)

type genesisProcessors struct {
	txCoordinator       process.TransactionCoordinator
	systemSCs           vm.SystemSCContainer
	txProcessor         process.TransactionProcessor
	scProcessor         process.SmartContractProcessor
	scrProcessor        process.SmartContractResultProcessor
	rwdProcessor        process.RewardTransactionProcessor
	blockchainHook      process.BlockChainHookHandler
	queryService        external.SCQueryService
	vmContainersFactory process.VirtualMachinesContainerFactory
	vmContainer         process.VirtualMachinesContainer
}
