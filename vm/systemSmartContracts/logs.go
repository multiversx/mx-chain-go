package systemSmartContracts

import (
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func (d *delegation) createAndAddLogEntry(contractCallInput *vmcommon.ContractCallInput, topics ...[]byte) {
	entry := &vmcommon.LogEntry{
		Identifier: []byte(contractCallInput.Function),
		Address:    contractCallInput.CallerAddr,
		Topics:     topics,
	}

	d.eei.AddLogEntry(entry)
}

func (d *delegation) createAndAddLogEntryForWithdraw(
	contractCallInput *vmcommon.ContractCallInput,
	actualUserUnBond *big.Int,
	globalFund *GlobalFundData,
	delegator *DelegatorData,
	dStatus *DelegationContractStatus,
) {
	activeFund := d.getFundForLogEntry(delegator.ActiveFund)

	numUsers := big.NewInt(0).SetUint64(dStatus.NumUsers)
	d.createAndAddLogEntry(contractCallInput, actualUserUnBond.Bytes(), activeFund.Bytes(), numUsers.Bytes(), globalFund.TotalActive.Bytes())
}

func (d *delegation) createAndAddLogEntryForDelegate(
	contractCallInput *vmcommon.ContractCallInput,
	delegationValue *big.Int,
	globalFund *GlobalFundData,
	delegator *DelegatorData,
	dStatus *DelegationContractStatus,
	isNew bool,
) {
	numUsersWithCurrent := dStatus.NumUsers
	if isNew {
		numUsersWithCurrent++
	}

	activeFund := d.getFundForLogEntry(delegator.ActiveFund)
	numUsers := big.NewInt(0).SetUint64(numUsersWithCurrent)
	numActiveWithCurrentValue := big.NewInt(0).Add(globalFund.TotalActive, delegationValue)
	delegatorActiveWithCurrent := big.NewInt(0).Add(activeFund, delegationValue)

	d.createAndAddLogEntry(contractCallInput, delegationValue.Bytes(), delegatorActiveWithCurrent.Bytes(), numUsers.Bytes(), numActiveWithCurrentValue.Bytes())
}

func (d *delegation) getFundForLogEntry(activeFund []byte) *big.Int {
	if len(activeFund) == 0 {
		return big.NewInt(0)
	}

	fund, err := d.getFund(activeFund)
	if err != nil {
		log.Warn("d.getFundForLogEntry cannot get fund", "error", err.Error())

		return big.NewInt(0)
	}

	return fund.Value
}
