package integrationTests

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/stakingcommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ProcessSCOutputAccounts will save account changes in accounts db from vmOutput
func ProcessSCOutputAccounts(vmOutput *vmcommon.VMOutput, accountsDB state.AccountsAdapter) error {
	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		acc := stakingcommon.LoadUserAccount(accountsDB, outAcc.Address)

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		for _, storeUpdate := range storageUpdates {
			err := acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
			if err != nil {
				return err
			}

			if outAcc.BalanceDelta != nil && outAcc.BalanceDelta.Cmp(zero) != 0 {
				err = acc.AddToBalance(outAcc.BalanceDelta)
				if err != nil {
					return err
				}
			}

			err = accountsDB.SaveAccount(acc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
