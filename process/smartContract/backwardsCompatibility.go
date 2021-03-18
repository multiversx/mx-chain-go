package smartContract

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
)

func (sc *scProcessor) updateDeveloperRewardsV1(
	tx data.TransactionHandler,
	vmOutput *vmcommon.VMOutput,
	builtInFuncGasUsed uint64,
) error {
	usedGasByMainSC := tx.GetGasLimit() - vmOutput.GasRemaining
	if !sc.isSelfShard(tx.GetSndAddr()) {
		usedGasByMainSC -= sc.economicsFee.ComputeGasLimit(tx)
	}
	usedGasByMainSC -= builtInFuncGasUsed

	for _, outAcc := range vmOutput.OutputAccounts {
		if bytes.Equal(tx.GetRcvAddr(), outAcc.Address) {
			continue
		}

		sentGas := uint64(0)
		for _, outTransfer := range outAcc.OutputTransfers {
			sentGas += outTransfer.GasLimit
		}
		usedGasByMainSC -= sentGas
		usedGasByMainSC -= outAcc.GasUsed

		if outAcc.GasUsed > 0 && sc.isSelfShard(outAcc.Address) {
			err := sc.addToDevRewardsV1(outAcc.Address, outAcc.GasUsed, tx.GetGasPrice())
			if err != nil {
				return err
			}
		}
	}

	err := sc.addToDevRewardsV1(tx.GetRcvAddr(), usedGasByMainSC, tx.GetGasPrice())
	if err != nil {
		return err
	}

	return nil
}

func (sc *scProcessor) addToDevRewardsV1(address []byte, gasUsed uint64, gasPrice uint64) error {
	if core.IsEmptyAddress(address) || !core.IsSmartContractAddress(address) {
		return nil
	}

	consumedFee := core.SafeMul(gasPrice, gasUsed)

	var devRwd *big.Int
	if sc.flagStakingV2.IsSet(){
		devRwd = core.GetIntTrimmedPercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())
	} else {
		devRwd = core.GetApproximatePercentageOfValue(consumedFee, sc.economicsFee.DeveloperPercentage())
	}

	userAcc, err := sc.getAccountFromAddress(address)
	if err != nil {
		return err
	}

	if check.IfNil(userAcc) {
		return nil
	}

	userAcc.AddToDeveloperReward(devRwd)
	err = sc.accounts.SaveAccount(userAcc)
	if err != nil {
		return err
	}

	return nil
}
