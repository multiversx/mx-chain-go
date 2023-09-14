package processorV2

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// VMOutputAccountsProcessor will process the VMOutput from a regular of builtin function call
type VMOutputAccountsProcessor struct {
	sc       *scProcessor
	vmInput  *vmcommon.VMInput
	vmOutput *vmcommon.VMOutput
	tx       data.TransactionHandler
	txHash   []byte
}

// NewVMOutputAccountsProcessor creates a new VMOutputProcessor instance.
func NewVMOutputAccountsProcessor(
	sc *scProcessor,
	vmInput *vmcommon.VMInput,
	vmOutput *vmcommon.VMOutput,
	tx data.TransactionHandler,
	txHash []byte) *VMOutputAccountsProcessor {
	return &VMOutputAccountsProcessor{
		sc:       sc,
		vmInput:  vmInput,
		vmOutput: vmOutput,
		tx:       tx,
		txHash:   txHash,
	}
}

// Run runs the configured processing steps on vm output
func (oap *VMOutputAccountsProcessor) Run() (bool, []data.TransactionHandler, error) {
	outputAccounts := process.SortVMOutputInsideData(oap.vmOutput)
	indexedSCResults := make([]internalIndexedScr, 0, len(outputAccounts))

	sumOfAllDiff := big.NewInt(0)
	sumOfAllDiff.Sub(sumOfAllDiff, oap.tx.GetValue())

	createdAsyncCallback := false
	for _, outAcc := range outputAccounts {
		acc, err := oap.sc.scProcessorHelper.GetAccountFromAddress(outAcc.Address)
		if err != nil {
			return false, nil, err
		}

		createdAsyncCallback, indexedSCResults = oap.processOutputTransfersStep(outAcc, createdAsyncCallback, indexedSCResults)

		continueWithNextAccount, err := oap.computeSumOfAllDiffStep(acc, outAcc, sumOfAllDiff)
		if err != nil {
			return false, nil, err
		}
		if continueWithNextAccount {
			continue
		}

		err = oap.processStorageUpdatesStep(acc, outAcc)
		if err != nil {
			return false, nil, err
		}

		err = oap.updateSmartContractCodeStep(oap.vmOutput, acc, outAcc)
		if err != nil {
			return false, nil, err
		}

		err = oap.updateAccountNonceIfThereIsAChange(acc, outAcc)
		if err != nil {
			return false, nil, err
		}

		sumOfAllDiff, err = oap.updateAccountBalanceStep(acc, outAcc, sumOfAllDiff)
		if err != nil {
			return false, nil, err
		}
	}

	if sumOfAllDiff.Cmp(zero) != 0 {
		return false, nil, process.ErrOverallBalanceChangeFromSC
	}

	sortIndexedScrs(indexedSCResults)

	scResults := unwrapSCRs(indexedSCResults)

	return createdAsyncCallback, scResults, nil
}

func unwrapSCRs(indexedSCResults []internalIndexedScr) []data.TransactionHandler {
	scResults := make([]data.TransactionHandler, 0, len(indexedSCResults))
	for _, indexedSCResult := range indexedSCResults {
		scResults = append(scResults, indexedSCResult.result)
	}
	return scResults
}

func sortIndexedScrs(indexedSCResults []internalIndexedScr) {
	sort.Slice(indexedSCResults, func(i, j int) bool {
		return indexedSCResults[i].index < indexedSCResults[j].index
	})
}

func (oap *VMOutputAccountsProcessor) processOutputTransfersStep(outAcc *vmcommon.OutputAccount, createdAsyncCallback bool, scResults []internalIndexedScr) (bool, []internalIndexedScr) {
	tmpCreatedAsyncCallback, newScrs := oap.sc.createSmartContractResults(oap.vmInput, oap.vmOutput, outAcc, oap.tx, oap.txHash)
	createdAsyncCallback = createdAsyncCallback || tmpCreatedAsyncCallback
	if len(newScrs) != 0 {
		scResults = append(scResults, newScrs...)
	}
	return createdAsyncCallback, scResults
}

func (oap *VMOutputAccountsProcessor) computeSumOfAllDiffStep(
	acc state.UserAccountHandler,
	outAcc *vmcommon.OutputAccount,
	sumOfAllDiff *big.Int) (bool, error) {
	if !check.IfNil(acc) {
		return false, nil
	}
	if outAcc.BalanceDelta != nil {
		if outAcc.BalanceDelta.Cmp(zero) < 0 {
			return false, process.ErrNegativeBalanceDeltaOnCrossShardAccount
		}
		sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)
	}
	return true, nil
}

func (oap *VMOutputAccountsProcessor) processStorageUpdatesStep(
	acc state.UserAccountHandler,
	outAcc *vmcommon.OutputAccount) error {
	for _, storeUpdate := range outAcc.StorageUpdates {
		if !process.IsAllowedToSaveUnderKey(storeUpdate.Offset) {
			log.Trace("storeUpdate is not allowed", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
			return process.ErrNotAllowedToWriteUnderProtectedKey
		}

		err := acc.SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
		if err != nil {
			log.Warn("saveKeyValue", "error", err)
			return err
		}
		log.Trace("storeUpdate", "acc", outAcc.Address, "key", storeUpdate.Offset, "data", storeUpdate.Data)
	}
	return nil
}

// updateSmartContractCode upgrades code for "direct" deployments & upgrades and for "indirect" deployments & upgrades
// It receives:
//
//	(1) the account as found in the State
//	(2) the account as returned in VM Output
//	(3) the transaction that, upon execution, produced the VM Output
func (oap *VMOutputAccountsProcessor) updateSmartContractCodeStep(
	vmOutput *vmcommon.VMOutput,
	stateAccount state.UserAccountHandler,
	outputAccount *vmcommon.OutputAccount,
) error {
	if len(outputAccount.Code) == 0 {
		return nil
	}
	if len(outputAccount.CodeMetadata) == 0 {
		return nil
	}
	if !core.IsSmartContractAddress(outputAccount.Address) {
		return nil
	}

	outputAccountCodeMetadataBytes, err := oap.sc.blockChainHook.FilterCodeMetadataForUpgrade(outputAccount.CodeMetadata)
	if err != nil {
		return err
	}

	// This check is desirable (not required though) since currently both Arwen and IELE send the code in the output account even for "regular" execution
	sameCode := bytes.Equal(outputAccount.Code, oap.sc.accounts.GetCode(stateAccount.GetCodeHash()))
	sameCodeMetadata := bytes.Equal(outputAccountCodeMetadataBytes, stateAccount.GetCodeMetadata())
	if sameCode && sameCodeMetadata {
		return nil
	}

	currentOwner := stateAccount.GetOwnerAddress()
	isCodeDeployerSet := len(outputAccount.CodeDeployerAddress) > 0
	isCodeDeployerOwner := bytes.Equal(currentOwner, outputAccount.CodeDeployerAddress) && isCodeDeployerSet

	noExistingCode := len(oap.sc.accounts.GetCode(stateAccount.GetCodeHash())) == 0
	noExistingOwner := len(currentOwner) == 0
	currentCodeMetadata := vmcommon.CodeMetadataFromBytes(stateAccount.GetCodeMetadata())
	newCodeMetadata := vmcommon.CodeMetadataFromBytes(outputAccountCodeMetadataBytes)
	isUpgradeable := currentCodeMetadata.Upgradeable
	isDeployment := noExistingCode && noExistingOwner
	isUpgrade := !isDeployment && isCodeDeployerOwner && isUpgradeable

	entry := &vmcommon.LogEntry{
		Address: stateAccount.AddressBytes(),
		Topics: [][]byte{
			outputAccount.Address, outputAccount.CodeDeployerAddress,
		},
	}

	if isDeployment {
		// At this point, we are under the condition "noExistingOwner"
		stateAccount.SetOwnerAddress(outputAccount.CodeDeployerAddress)
		stateAccount.SetCodeMetadata(outputAccountCodeMetadataBytes)
		stateAccount.SetCode(outputAccount.Code)
		log.Debug("updateSmartContractCode(): created",
			"address", oap.sc.pubkeyConv.SilentEncode(outputAccount.Address, log),
			"upgradeable", newCodeMetadata.Upgradeable)

		entry.Identifier = []byte(core.SCDeployIdentifier)
		vmOutput.Logs = append(vmOutput.Logs, entry)
		return nil
	}

	if isUpgrade {
		stateAccount.SetCodeMetadata(outputAccountCodeMetadataBytes)
		stateAccount.SetCode(outputAccount.Code)
		log.Debug("updateSmartContractCode(): upgraded",
			"address", oap.sc.pubkeyConv.SilentEncode(outputAccount.Address, log),
			"upgradeable", newCodeMetadata.Upgradeable)

		entry.Identifier = []byte(core.SCUpgradeIdentifier)
		vmOutput.Logs = append(vmOutput.Logs, entry)
		return nil
	}

	return nil
}

func (oap *VMOutputAccountsProcessor) updateAccountNonceIfThereIsAChange(
	acc state.UserAccountHandler,
	outAcc *vmcommon.OutputAccount) error {
	if outAcc.Nonce != acc.GetNonce() && outAcc.Nonce != 0 {
		if outAcc.Nonce < acc.GetNonce() {
			return process.ErrWrongNonceInVMOutput
		}

		nonceDifference := outAcc.Nonce - acc.GetNonce()
		acc.IncreaseNonce(nonceDifference)
	}
	return nil
}

func (oap *VMOutputAccountsProcessor) updateAccountBalanceStep(
	acc state.UserAccountHandler,
	outAcc *vmcommon.OutputAccount,
	sumOfAllDiff *big.Int) (*big.Int, error) {
	// if no change then continue to next account
	if outAcc.BalanceDelta == nil || outAcc.BalanceDelta.Cmp(zero) == 0 {
		err := oap.sc.accounts.SaveAccount(acc)
		if err != nil {
			return sumOfAllDiff, err
		}

		return sumOfAllDiff, nil
	}

	sumOfAllDiff = big.NewInt(0).Add(sumOfAllDiff, outAcc.BalanceDelta)

	err := acc.AddToBalance(outAcc.BalanceDelta)
	if err != nil {
		return sumOfAllDiff, err
	}

	err = oap.sc.accounts.SaveAccount(acc)
	if err != nil {
		return sumOfAllDiff, err
	}

	return sumOfAllDiff, nil
}
