package runType

import (
	"github.com/multiversx/mx-chain-go/config"
	mainFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"

	"github.com/multiversx/mx-chain-core-go/core"
)

func CreateArgsRunTypeComponents(
	coreComponents mainFactory.CoreComponentsHandler,
	cryptoComponents mainFactory.CryptoComponentsHandler,
	configs config.Configs,
) (*ArgsRunTypeComponents, error) {
	initialAccounts := make([]*data.InitialAccount, 0)
	err := core.LoadJsonFile(&initialAccounts, configs.ConfigurationPathsHolder.Genesis)
	if err != nil {
		return nil, err
	}

	var accounts []genesis.InitialAccountHandler
	for _, ia := range initialAccounts {
		accounts = append(accounts, ia)
	}

	return &ArgsRunTypeComponents{
		CoreComponents:   coreComponents,
		CryptoComponents: cryptoComponents,
		Configs:          configs,
		InitialAccounts:  accounts,
	}, nil
}
