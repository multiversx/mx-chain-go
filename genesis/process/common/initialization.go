package common

import (
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
)

// UpdateSystemSCContractsCode will update system accounts contract codes, code, and owner, except for the sys account for each shard
func UpdateSystemSCContractsCode(contractMetadata []byte, userAccountsDB state.AccountsAdapter) error {
	contractsToUpdate := make([][]byte, 0)
	contractsToUpdate = append(contractsToUpdate, vm.StakingSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.ValidatorSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.GovernanceSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.ESDTSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.DelegationManagerSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.FirstDelegationSCAddress)

	for _, address := range contractsToUpdate {
		userAcc, err := getUserAccount(address, userAccountsDB)
		if err != nil {
			return err
		}

		userAcc.SetOwnerAddress(address)
		userAcc.SetCodeMetadata(contractMetadata)
		userAcc.SetCode(address)

		err = userAccountsDB.SaveAccount(userAcc)
		if err != nil {
			return err
		}
	}

	return nil
}

func getUserAccount(address []byte, userAccountsDB state.AccountsAdapter) (state.UserAccountHandler, error) {
	acc, err := userAccountsDB.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	userAcc, ok := acc.(state.UserAccountHandler)
	if !ok {
		return nil, errors.ErrWrongTypeAssertion
	}

	return userAcc, nil
}
