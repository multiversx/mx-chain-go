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
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
)

const (
	// ChainID contains the chain id
	ChainID = "chain"
)

// ArgsChainSimulatorConfigs holds all the components needed to create the chain simulator configs
type ArgsChainSimulatorConfigs struct {
	NumOfShards               uint32
	OriginalConfigsPath       string
	GenesisAddressWithStake   string
	GenesisAddressWithBalance string
	GenesisTimeStamp          int64
	RoundDurationInMillis     uint64
}

// ArgsConfigsSimulator holds the configs for the chain simulator
type ArgsConfigsSimulator struct {
	GasScheduleFilename   string
	Configs               *config.Configs
	ValidatorsPrivateKeys []crypto.PrivateKey
	ValidatorsPublicKeys  map[uint32][]byte
}

// CreateChainSimulatorConfigs will create the chain simulator configs
func CreateChainSimulatorConfigs(args ArgsChainSimulatorConfigs) (*ArgsConfigsSimulator, error) {
	configs, err := testscommon.CreateTestConfigs(args.OriginalConfigsPath)
	if err != nil {
		return nil, err
	}

	configs.GeneralConfig.GeneralSettings.ChainID = ChainID

	// empty genesis smart contracts file
	err = os.WriteFile(configs.ConfigurationPathsHolder.SmartContracts, []byte("[]"), os.ModePerm)
	if err != nil {
		return nil, err
	}

	// generate validators key and nodesSetup.json
	privateKeys, publicKeys, err := generateValidatorsKeyAndUpdateFiles(
		configs,
		args.NumOfShards,
		args.GenesisAddressWithStake,
		args.GenesisTimeStamp,
		args.RoundDurationInMillis,
	)
	if err != nil {
		return nil, err
	}

	// update genesis.json
	addresses := make([]data.InitialAccount, 0)
	// 10_000 egld
	bigValue, _ := big.NewInt(0).SetString("10000000000000000000000", 0)
	addresses = append(addresses, data.InitialAccount{
		Address:      args.GenesisAddressWithStake,
		StakingValue: bigValue,
		Supply:       bigValue,
	})

	bigValueAddr, _ := big.NewInt(0).SetString("19990000000000000000000000", 10)
	addresses = append(addresses, data.InitialAccount{
		Address: args.GenesisAddressWithBalance,
		Balance: bigValueAddr,
		Supply:  bigValueAddr,
	})

	addressesBytes, errM := json.Marshal(addresses)
	if errM != nil {
		return nil, errM
	}

	err = os.WriteFile(configs.ConfigurationPathsHolder.Genesis, addressesBytes, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// generate validators.pem
	configs.ConfigurationPathsHolder.ValidatorKey = path.Join(args.OriginalConfigsPath, "validatorKey.pem")
	err = generateValidatorsPem(configs.ConfigurationPathsHolder.ValidatorKey, publicKeys, privateKeys)
	if err != nil {
		return nil, err
	}

	gasScheduleName, err := GetLatestGasScheduleFilename(configs.ConfigurationPathsHolder.GasScheduleDirectoryName)
	if err != nil {
		return nil, err
	}

	configs.GeneralConfig.SmartContractsStorage.DB.Type = string(storageunit.MemoryDB)
	configs.GeneralConfig.SmartContractsStorageForSCQuery.DB.Type = string(storageunit.MemoryDB)
	configs.GeneralConfig.SmartContractsStorageSimulate.DB.Type = string(storageunit.MemoryDB)

	// enable db lookup extension
	configs.GeneralConfig.DbLookupExtensions.Enabled = true

	publicKeysBytes := make(map[uint32][]byte)
	publicKeysBytes[core.MetachainShardId], err = publicKeys[0].ToByteArray()
	if err != nil {
		return nil, err
	}

	for idx := uint32(1); idx < uint32(len(publicKeys)); idx++ {
		publicKeysBytes[idx], err = publicKeys[idx].ToByteArray()
		if err != nil {
			return nil, err
		}
	}

	return &ArgsConfigsSimulator{
		Configs:               configs,
		ValidatorsPrivateKeys: privateKeys,
		GasScheduleFilename:   gasScheduleName,
		ValidatorsPublicKeys:  publicKeysBytes,
	}, nil
}

func generateValidatorsKeyAndUpdateFiles(
	configs *config.Configs,
	numOfShards uint32,
	address string,
	genesisTimeStamp int64,
	roundDurationInMillis uint64,
) ([]crypto.PrivateKey, []crypto.PublicKey, error) {
	blockSigningGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	nodesSetupFile := configs.ConfigurationPathsHolder.Nodes
	nodes := &sharding.NodesSetup{}
	err := core.LoadJsonFile(nodes, nodesSetupFile)
	if err != nil {
		return nil, nil, err
	}

	nodes.RoundDuration = roundDurationInMillis
	nodes.StartTime = genesisTimeStamp
	nodes.ConsensusGroupSize = 1
	nodes.MinNodesPerShard = 1
	nodes.MetaChainMinNodes = 1
	nodes.MetaChainConsensusGroupSize = 1
	nodes.InitialNodes = make([]*sharding.InitialNode, 0)

	privateKeys := make([]crypto.PrivateKey, 0, numOfShards+1)
	publicKeys := make([]crypto.PublicKey, 0, numOfShards+1)
	for idx := uint32(0); idx < numOfShards+1; idx++ {
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
