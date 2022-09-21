package processProxy

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/processorV2"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/scrCommon"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("processProxy")

var _ process.SmartContractProcessorFull = (*SCProcessorProxy)(nil)

type configuredProcessor string

const (
	procV1 configuredProcessor = "processorV1"
	procV2 configuredProcessor = "processorV2"
)

type SCProcessorProxy struct {
	processor           process.SmartContractProcessorFull
	args                scrCommon.ArgsNewSmartContractProcessor
	configuredProcessor configuredProcessor
	mutRc               sync.Mutex
}

// NewSmartContractProcessorProxy creates a smart contract processor proxy
func NewSmartContractProcessorProxy(args scrCommon.ArgsNewSmartContractProcessor, epochNotifier vmcommon.EpochNotifier) (*SCProcessorProxy, error) {
	scProcessorProxy := &SCProcessorProxy{
		args: args,
	}

	var err error
	scProcessorProxy.processor, err = scProcessorProxy.createProcessorV1()
	if err != nil {
		return nil, err
	}
	scProcessorProxy.configuredProcessor = procV1

	_, err = scProcessorProxy.createProcessorV2()
	if err != nil {
		return nil, err
	}

	log.Info("processorProxy", "configured", procV1)

	epochNotifier.RegisterNotifyHandler(scProcessorProxy)

	return scProcessorProxy, nil
}

func (scPProxy *SCProcessorProxy) createProcessorV1() (process.SmartContractProcessorFull, error) {
	return smartContract.NewSmartContractProcessor(scPProxy.args)
}

func (scPProxy *SCProcessorProxy) createProcessorV2() (process.SmartContractProcessorFull, error) {
	return processorV2.NewSmartContractProcessorV2(scPProxy.args)
}

func (scPProxy *SCProcessorProxy) ExecuteSmartContractTransaction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return scPProxy.processor.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
}

func (scPProxy *SCProcessorProxy) ExecuteBuiltInFunction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return scPProxy.processor.ExecuteBuiltInFunction(tx, acntSrc, acntDst)
}

func (scPProxy *SCProcessorProxy) DeploySmartContract(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return scPProxy.processor.DeploySmartContract(tx, acntSrc)
}

func (scPProxy *SCProcessorProxy) ProcessIfError(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
	return scPProxy.processor.ProcessIfError(acntSnd, txHash, tx, returnCode, returnMessage, snapshot, gasLocked)
}

func (scPProxy *SCProcessorProxy) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	return scPProxy.processor.IsPayable(sndAddress, recvAddress)
}

func (scPProxy *SCProcessorProxy) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	return scPProxy.processor.ProcessSmartContractResult(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scPProxy *SCProcessorProxy) IsInterfaceNil() bool {
	return scPProxy == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (scPProxy *SCProcessorProxy) EpochConfirmed(epoch uint32, _ uint64) {
	scPProxy.mutRc.Lock()
	defer scPProxy.mutRc.Unlock()

	if epoch > scPProxy.args.EnableEpochs.SCProcessorV2EnableEpoch {
		if scPProxy.configuredProcessor != procV2 {
			// what to do with the possible error ?
			scPProxy.processor, _ = scPProxy.createProcessorV2()
		}
		return
	}

	if scPProxy.configuredProcessor != procV1 {
		// what to do with the possible error ?
		scPProxy.processor, _ = scPProxy.createProcessorV1()
	}
}
