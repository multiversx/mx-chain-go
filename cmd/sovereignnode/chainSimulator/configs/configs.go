package configs

import (
	"encoding/json"
	"math/big"
	"os"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	chainSimulatorConfigs "github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

var initialStakedEgldPerNode = big.NewInt(0).Mul(chainSimulator.OneEGLD, big.NewInt(2500))
var initialEgldPerNode = big.NewInt(0).Mul(chainSimulator.OneEGLD, big.NewInt(5000))
var initialSupply = big.NewInt(0).Mul(chainSimulator.OneEGLD, big.NewInt(20000000)) // 20 million EGLD

// GenerateSovereignGenesisFile will generate sovereign initial wallet keys
func GenerateSovereignGenesisFile(args chainSimulatorConfigs.ArgsChainSimulatorConfigs, configs *config.Configs) (*dtos.InitialWalletKeys, error) {
	addressConverter, err := factory.NewPubkeyConverter(configs.GeneralConfig.AddressPubkeyConverter)
	if err != nil {
		return nil, err
	}

	initialWalletKeys := &dtos.InitialWalletKeys{
		BalanceWallets: make(map[uint32]*dtos.WalletKey),
		StakeWallets:   make([]*dtos.WalletKey, 0),
	}
	addresses := make([]data.InitialAccount, 0)
	numOfNodes := int(args.NumNodesWaitingListShard + args.MinNodesPerShard)

	totalStakedValue := big.NewInt(0).Set(initialStakedEgldPerNode)
	totalStakedValue.Mul(totalStakedValue, big.NewInt(int64(numOfNodes)))

	initialBalance := big.NewInt(0).Set(initialSupply)
	initialBalance.Sub(initialBalance, totalStakedValue)

	for i := 0; i < numOfNodes; i++ {
		initialEgld := big.NewInt(0).Set(initialEgldPerNode)

		if i == numOfNodes-1 {
			remainingSupply := big.NewInt(0).Set(initialBalance)
			allBalances := big.NewInt(0).Set(initialEgld)
			allBalances.Mul(allBalances, big.NewInt(int64(numOfNodes)))
			remainingSupply.Sub(remainingSupply, allBalances)
			if remainingSupply.Cmp(big.NewInt(0)) > 0 {
				initialEgld.Add(initialEgld, remainingSupply)
			}
		}

		walletKey, errG := chainSimulatorConfigs.GenerateWalletKeyForShard(core.SovereignChainShardId, args.NumOfShards, addressConverter)
		if errG != nil {
			return nil, errG
		}

		supply := big.NewInt(0).Set(initialEgld)
		supply.Add(supply, big.NewInt(0).Set(initialStakedEgldPerNode))

		addresses = append(addresses, data.InitialAccount{
			Address:      walletKey.Address.Bech32,
			Balance:      initialEgld,
			Supply:       supply,
			StakingValue: big.NewInt(0).Set(initialStakedEgldPerNode),
		})

		initialWalletKeys.StakeWallets = append(initialWalletKeys.StakeWallets, walletKey)
		initialWalletKeys.BalanceWallets[uint32(i)] = walletKey
	}

	addressesBytes, errM := json.Marshal(addresses)
	if errM != nil {
		return nil, errM
	}

	err = os.WriteFile(configs.ConfigurationPathsHolder.Genesis, addressesBytes, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return initialWalletKeys, nil
}
