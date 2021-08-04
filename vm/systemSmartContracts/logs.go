package systemSmartContracts

import (
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func createLogEntryForDelegate(
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

	numUsers := big.NewInt(0).SetUint64(numUsersWithCurrent)
	numActiveWithCurrentValue := big.NewInt(0).Add(globalFund.TotalActive, delegationValue)
	delegatorActiveWithCurrent := big.NewInt(0).Add(big.NewInt(0).SetBytes(delegator.ActiveFund), delegationValue)
	return &vmcommon.LogEntry{
		Identifier: []byte(funcIdentifier),
		Address:    callerAddr,
		Topics: [][]byte{
			delegationValue.Bytes(), delegatorActiveWithCurrent.Bytes(), numUsers.Bytes(), numActiveWithCurrentValue.Bytes(),
		},
	}
}
