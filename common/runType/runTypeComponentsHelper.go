package runType

import (
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"

	"github.com/multiversx/mx-chain-core-go/core"
)

// ReadInitialAccounts returns the genesis accounts from a file
func ReadInitialAccounts(filePath string) ([]genesis.InitialAccountHandler, error) {
	initialAccounts := make([]*data.InitialAccount, 0)
	err := core.LoadJsonFile(&initialAccounts, filePath)
	if err != nil {
		return nil, err
	}

	var accounts []genesis.InitialAccountHandler
	for _, ia := range initialAccounts {
		accounts = append(accounts, ia)
	}

	return accounts, nil
}
