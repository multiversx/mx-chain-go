package runType

import (
	"github.com/multiversx/mx-chain-go/config"
	mainFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"

	"github.com/multiversx/mx-chain-core-go/core"
)

// CreateArgsRunTypeComponents creates the args for run type component
func CreateArgsRunTypeComponents(
	coreComponents mainFactory.CoreComponentsHandler,
	cryptoComponents mainFactory.CryptoComponentsHandler,
	configs config.Configs,
) (*runType.ArgsRunTypeComponents, error) {
	initialAccounts := make([]*data.InitialAccount, 0)
	err := core.LoadJsonFile(&initialAccounts, configs.ConfigurationPathsHolder.Genesis)
	if err != nil {
		return nil, err
	}

	var accounts []genesis.InitialAccountHandler
	for _, ia := range initialAccounts {
		accounts = append(accounts, ia)
	}

	return &runType.ArgsRunTypeComponents{
		CoreComponents:   coreComponents,
		CryptoComponents: cryptoComponents,
		Configs:          configs,
		InitialAccounts:  accounts,
	}, nil
}

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
