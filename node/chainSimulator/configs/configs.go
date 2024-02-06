package configs

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	shardingCore "github.com/multiversx/mx-chain-core-go/core/sharding"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialStakedEgldPerNode = big.NewInt(0).Mul(oneEgld, big.NewInt(2500))
var initialSupply = big.NewInt(0).Mul(oneEgld, big.NewInt(20000000)) // 20 million EGLD
const (
	// ChainID contains the chain id
	ChainID = "chain"

	shardIDWalletWithStake   = 0
	allValidatorsPemFileName = "allValidatorsKeys.pem"
)

// ArgsChainSimulatorConfigs holds all the components needed to create the chain simulator configs
type ArgsChainSimulatorConfigs struct {
	NumOfShards           uint32
	OriginalConfigsPath   string
	GenesisTimeStamp      int64
	RoundDurationInMillis uint64
	TempDir               string
	MinNodesPerShard      uint32
	MetaChainMinNodes     uint32
	RoundsPerEpoch        core.OptionalUint64
	AlterConfigsFunction  func(cfg *config.Configs)
}

// ArgsConfigsSimulator holds the configs for the chain simulator
type ArgsConfigsSimulator struct {
	GasScheduleFilename   string
	Configs               config.Configs
	ValidatorsPrivateKeys []crypto.PrivateKey
	InitialWallets        *dtos.InitialWalletKeys
}

// CreateChainSimulatorConfigs will create the chain simulator configs
func CreateChainSimulatorConfigs(args ArgsChainSimulatorConfigs) (*ArgsConfigsSimulator, error) {
	configs, err := testscommon.CreateTestConfigs(args.TempDir, args.OriginalConfigsPath)
	if err != nil {
		return nil, err
	}

	if args.AlterConfigsFunction != nil {
		args.AlterConfigsFunction(configs)
	}

	configs.GeneralConfig.GeneralSettings.ChainID = ChainID

	// empty genesis smart contracts file
	err = os.WriteFile(configs.ConfigurationPathsHolder.SmartContracts, []byte("[]"), os.ModePerm)
	if err != nil {
		return nil, err
	}

	// update genesis.json
	initialWallets, err := generateGenesisFile(args, configs)
	if err != nil {
		return nil, err
	}

	// generate validators key and nodesSetup.json
	privateKeys, publicKeys, err := generateValidatorsKeyAndUpdateFiles(
		configs,
		initialWallets.InitialWalletWithStake.Address,
		args,
	)
	if err != nil {
		return nil, err
	}

	configs.ConfigurationPathsHolder.AllValidatorKeys = path.Join(args.OriginalConfigsPath, allValidatorsPemFileName)
	err = generateValidatorsPem(configs.ConfigurationPathsHolder.AllValidatorKeys, publicKeys, privateKeys)
	if err != nil {
		return nil, err
	}

	configs.GeneralConfig.SmartContractsStorage.DB.Type = string(storageunit.MemoryDB)
	configs.GeneralConfig.SmartContractsStorageForSCQuery.DB.Type = string(storageunit.MemoryDB)
	configs.GeneralConfig.SmartContractsStorageSimulate.DB.Type = string(storageunit.MemoryDB)

	maxNumNodes := uint64(args.MinNodesPerShard*args.NumOfShards+args.MetaChainMinNodes) + 2*uint64(args.NumOfShards+1)
	configs.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake = maxNumNodes
	numMaxNumNodesEnableEpochs := len(configs.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch)
	for idx := 0; idx < numMaxNumNodesEnableEpochs-1; idx++ {
		configs.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[idx].MaxNumNodes = uint32(maxNumNodes)
	}

	configs.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[numMaxNumNodesEnableEpochs-1].EpochEnable = configs.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch
	prevEntry := configs.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[numMaxNumNodesEnableEpochs-2]
	configs.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[numMaxNumNodesEnableEpochs-1].NodesToShufflePerShard = prevEntry.NodesToShufflePerShard
	configs.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[numMaxNumNodesEnableEpochs-1].MaxNumNodes = prevEntry.MaxNumNodes - (args.NumOfShards+1)*prevEntry.NodesToShufflePerShard

	// set compatible trie configs
	configs.GeneralConfig.StateTriesConfig.SnapshotsEnabled = false

	// enable db lookup extension
	configs.GeneralConfig.DbLookupExtensions.Enabled = true

	if args.RoundsPerEpoch.HasValue {
		configs.GeneralConfig.EpochStartConfig.RoundsPerEpoch = int64(args.RoundsPerEpoch.Value)
	}

	gasScheduleName, err := GetLatestGasScheduleFilename(configs.ConfigurationPathsHolder.GasScheduleDirectoryName)
	if err != nil {
		return nil, err
	}

	return &ArgsConfigsSimulator{
		Configs:               *configs,
		ValidatorsPrivateKeys: privateKeys,
		GasScheduleFilename:   gasScheduleName,
		InitialWallets:        initialWallets,
	}, nil
}

func generateGenesisFile(args ArgsChainSimulatorConfigs, configs *config.Configs) (*dtos.InitialWalletKeys, error) {
	addressConverter, err := factory.NewPubkeyConverter(configs.GeneralConfig.AddressPubkeyConverter)
	if err != nil {
		return nil, err
	}

	initialWalletKeys := &dtos.InitialWalletKeys{
		ShardWallets: make(map[uint32]*dtos.WalletKey),
	}

	initialAddressWithStake, err := generateWalletKeyForShard(shardIDWalletWithStake, args.NumOfShards, addressConverter)
	if err != nil {
		return nil, err
	}

	initialWalletKeys.InitialWalletWithStake = initialAddressWithStake

	addresses := make([]data.InitialAccount, 0)
	stakedValue := big.NewInt(0).Set(initialStakedEgldPerNode)
	numOfNodes := args.MinNodesPerShard*args.NumOfShards + args.MetaChainMinNodes
	stakedValue = stakedValue.Mul(stakedValue, big.NewInt(int64(numOfNodes))) // 2500 EGLD * number of nodes
	addresses = append(addresses, data.InitialAccount{
		Address:      initialAddressWithStake.Address,
		StakingValue: stakedValue,
		Supply:       stakedValue,
	})

	// generate an address for every shard
	initialBalance := big.NewInt(0).Set(initialSupply)
	initialBalance = initialBalance.Sub(initialBalance, stakedValue)

	walletBalance := big.NewInt(0).Set(initialBalance)
	walletBalance.Div(walletBalance, big.NewInt(int64(args.NumOfShards)))

	// remainder = balance % numTotalWalletKeys
	remainder := big.NewInt(0).Set(initialBalance)
	remainder.Mod(remainder, big.NewInt(int64(args.NumOfShards)))

	for shardID := uint32(0); shardID < args.NumOfShards; shardID++ {
		walletKey, errG := generateWalletKeyForShard(shardID, args.NumOfShards, addressConverter)
		if errG != nil {
			return nil, errG
		}

		addresses = append(addresses, data.InitialAccount{
			Address: walletKey.Address,
			Balance: big.NewInt(0).Set(walletBalance),
			Supply:  big.NewInt(0).Set(walletBalance),
		})

		initialWalletKeys.ShardWallets[shardID] = walletKey
	}

	addresses[1].Balance.Add(walletBalance, remainder)
	addresses[1].Supply.Add(walletBalance, remainder)

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

func generateValidatorsKeyAndUpdateFiles(
	configs *config.Configs,
	address string,
	args ArgsChainSimulatorConfigs,
) ([]crypto.PrivateKey, []crypto.PublicKey, error) {
	blockSigningGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	nodesSetupFile := configs.ConfigurationPathsHolder.Nodes
	nodes := &sharding.NodesSetup{}
	err := core.LoadJsonFile(nodes, nodesSetupFile)
	if err != nil {
		return nil, nil, err
	}

	nodes.RoundDuration = args.RoundDurationInMillis
	nodes.StartTime = args.GenesisTimeStamp

	nodes.ConsensusGroupSize = 1
	nodes.MetaChainConsensusGroupSize = 1

	nodes.MinNodesPerShard = args.MinNodesPerShard
	nodes.MetaChainMinNodes = args.MetaChainMinNodes

	nodes.InitialNodes = make([]*sharding.InitialNode, 0)
	privateKeys := make([]crypto.PrivateKey, 0)
	publicKeys := make([]crypto.PublicKey, 0)
	// generate meta keys
	for idx := uint32(0); idx < args.MetaChainMinNodes; idx++ {
		sk, pk := blockSigningGenerator.GeneratePair()
		privateKeys = append(privateKeys, sk)
		publicKeys = append(publicKeys, pk)

		pkBytes, errB := pk.ToByteArray()
		if errB != nil {
			return nil, nil, errB
		}

		nodes.InitialNodes = append(nodes.InitialNodes, &sharding.InitialNode{
			PubKey:  hex.EncodeToString(pkBytes),
			Address: address,
		})
	}

	// generate shard keys
	for idx1 := uint32(0); idx1 < args.NumOfShards; idx1++ {
		for idx2 := uint32(0); idx2 < args.MinNodesPerShard; idx2++ {
			sk, pk := blockSigningGenerator.GeneratePair()
			privateKeys = append(privateKeys, sk)
			publicKeys = append(publicKeys, pk)

			pkBytes, errB := pk.ToByteArray()
			if errB != nil {
				return nil, nil, errB
			}

			nodes.InitialNodes = append(nodes.InitialNodes, &sharding.InitialNode{
				PubKey:  hex.EncodeToString(pkBytes),
				Address: address,
			})
		}
	}

	marshaledNodes, err := json.Marshal(nodes)
	if err != nil {
		return nil, nil, err
	}

	err = os.WriteFile(nodesSetupFile, marshaledNodes, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}

	return privateKeys, publicKeys, nil
}

func generateValidatorsPem(validatorsFile string, publicKeys []crypto.PublicKey, privateKey []crypto.PrivateKey) error {
	validatorPubKeyConverter, err := pubkeyConverter.NewHexPubkeyConverter(96)
	if err != nil {
		return err
	}

	buff := bytes.Buffer{}
	for idx := 0; idx < len(publicKeys); idx++ {
		publicKeyBytes, errA := publicKeys[idx].ToByteArray()
		if errA != nil {
			return errA
		}

		pkString, errE := validatorPubKeyConverter.Encode(publicKeyBytes)
		if errE != nil {
			return errE
		}

		privateKeyBytes, errP := privateKey[idx].ToByteArray()
		if errP != nil {
			return errP
		}

		blk := pem.Block{
			Type:  "PRIVATE KEY for " + pkString,
			Bytes: []byte(hex.EncodeToString(privateKeyBytes)),
		}

		err = pem.Encode(&buff, &blk)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(validatorsFile, buff.Bytes(), 0644)
}

// GetLatestGasScheduleFilename will parse the provided path and get the latest gas schedule filename
func GetLatestGasScheduleFilename(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}

	extension := ".toml"
	versionMarker := "V"

	highestVersion := 0
	filename := ""
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		splt := strings.Split(name, versionMarker)
		if len(splt) != 2 {
			continue
		}

		versionAsString := splt[1][:len(splt[1])-len(extension)]
		number, errConversion := strconv.Atoi(versionAsString)
		if errConversion != nil {
			continue
		}

		if number > highestVersion {
			highestVersion = number
			filename = name
		}
	}

	return path.Join(directory, filename), nil
}

func generateWalletKeyForShard(shardID, numOfShards uint32, converter core.PubkeyConverter) (*dtos.WalletKey, error) {
	walletSuite := ed25519.NewEd25519()
	walletKeyGenerator := signing.NewKeyGenerator(walletSuite)

	for {
		sk, pk := walletKeyGenerator.GeneratePair()

		pubKeyBytes, err := pk.ToByteArray()
		if err != nil {
			return nil, err
		}

		addressShardID := shardingCore.ComputeShardID(pubKeyBytes, numOfShards)
		if addressShardID != shardID {
			continue
		}

		privateKeyBytes, err := sk.ToByteArray()
		if err != nil {
			return nil, err
		}

		address, err := converter.Encode(pubKeyBytes)
		if err != nil {
			return nil, err
		}

		return &dtos.WalletKey{
			Address:       address,
			PrivateKeyHex: hex.EncodeToString(privateKeyBytes),
		}, nil
	}
}
