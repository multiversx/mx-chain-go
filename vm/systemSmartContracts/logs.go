package systemSmartContracts

import (
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func (d *delegation) createLogEntryForDelegate(
	funcIdentifier string,
	callerAddr []byte,
	delegationValue *big.Int,
	globalFund *GlobalFundData,
	delegator *DelegatorData,
	dStatus *DelegationContractStatus,
	isNew bool,
) *vmcommon.LogEntry {
	numUsersWithCurrent := dStatus.NumUsers
	if isNew {
		numUsersWithCurrent++
	}

	activeFund := big.NewInt(0)
	if len(delegator.ActiveFund) != 0 {
		fund, err := d.getFund(delegator.ActiveFund)
		if err != nil {
			log.Warn("d.createLogEntryForDelegate cannot get fund", "error", err.Error())
		} else {
			activeFund = fund.Value
		}
	}

	numUsers := big.NewInt(0).SetUint64(numUsersWithCurrent)
	numActiveWithCurrentValue := big.NewInt(0).Add(globalFund.TotalActive, delegationValue)
	delegatorActiveWithCurrent := big.NewInt(0).Add(activeFund, delegationValue)
	return &vmcommon.LogEntry{
		Identifier: []byte(funcIdentifier),
		Address:    callerAddr,
		Topics: [][]byte{
			delegationValue.Bytes(), delegatorActiveWithCurrent.Bytes(), numUsers.Bytes(), numActiveWithCurrentValue.Bytes(),
		},
	}
}
