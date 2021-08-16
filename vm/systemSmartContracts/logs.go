package systemSmartContracts

import (
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func (d *delegation) createAndAddLogEntry(function []byte, called []byte, topics ...[]byte) {
	entry := &vmcommon.LogEntry{
		Identifier: function,
		Address:    called,
		Topics:     topics,
	}

	d.eei.AddLogEntry(entry)
}

func (d *delegation) createAndAddLogEntryForWithdraw(
	funcIdentifier string,
	callerAddr []byte,
	delegationValue *big.Int,
	globalFund *GlobalFundData,
	delegator *DelegatorData,
	dStatus *DelegationContractStatus,
) {
	activeFund := big.NewInt(0)
	fund, err := d.getFund(delegator.ActiveFund)
	if err != nil {
		log.Warn("d.createLogEntryForWithdraw cannot get fund", "error", err.Error())
	} else {
		activeFund = fund.Value
	}

	numUsers := big.NewInt(0).SetUint64(dStatus.NumUsers)
	d.createAndAddLogEntry([]byte(funcIdentifier), callerAddr, delegationValue.Bytes(), activeFund.Bytes(), numUsers.Bytes(), globalFund.TotalActive.Bytes())
}

func (d *delegation) createAndAddLogEntryForDelegate(
	funcIdentifier string,
	callerAddr []byte,
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

	activeFund := big.NewInt(0)
	fund, err := d.getFund(delegator.ActiveFund)
	if err != nil {
		log.Warn("d.createLogEntryForDelegate cannot get fund", "error", err.Error())
	} else {
		activeFund = fund.Value
	}

	numUsers := big.NewInt(0).SetUint64(numUsersWithCurrent)
	numActiveWithCurrentValue := big.NewInt(0).Add(globalFund.TotalActive, delegationValue)
	delegatorActiveWithCurrent := big.NewInt(0).Add(activeFund, delegationValue)

	d.createAndAddLogEntry([]byte(funcIdentifier), callerAddr, delegationValue.Bytes(), delegatorActiveWithCurrent.Bytes(), numUsers.Bytes(), numActiveWithCurrentValue.Bytes())
}
