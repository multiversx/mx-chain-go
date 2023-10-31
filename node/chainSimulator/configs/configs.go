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
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

type ArgsChainSimulatorConfigs struct {
	NumOfShards               uint32
	OriginalConfigsPath       string
	GenesisAddressWithStake   string
	GenesisAddressWithBalance string
}

type ArgsConfigsSimulator struct {
	GasScheduleFilename   string
	Configs               *config.Configs
	ValidatorsPrivateKeys []crypto.PrivateKey
}

func CreateChainSimulatorConfigs(args ArgsChainSimulatorConfigs) (*ArgsConfigsSimulator, error) {
	configs, err := testscommon.CreateTestConfigs(args.OriginalConfigsPath)
	if err != nil {
		return nil, err
	}

	// empty genesis smart contracts file
	err = modifyFile(configs.ConfigurationPathsHolder.SmartContracts, func(intput []byte) ([]byte, error) {
		return []byte("[]"), nil
	})
	if err != nil {
		return nil, err
	}

	// generate validatos key and nodesSetup.json
	privateKeys, publicKeys := generateValidatorsKeyAndUpdateFiles(nil, configs, args.NumOfShards, args.GenesisAddressWithStake)

	// update genesis.json
	err = modifyFile(configs.ConfigurationPathsHolder.Genesis, func(i []byte) ([]byte, error) {
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

		return addressesBytes, nil
	})
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

	return &ArgsConfigsSimulator{
		Configs:               configs,
		ValidatorsPrivateKeys: privateKeys,
		GasScheduleFilename:   gasScheduleName,
	}, nil
}

func generateValidatorsKeyAndUpdateFiles(tb testing.TB, configs *config.Configs, numOfShards uint32, address string) ([]crypto.PrivateKey, []crypto.PublicKey) {
	blockSigningGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	nodesSetupFile := configs.ConfigurationPathsHolder.Nodes
	nodes := &sharding.NodesSetup{}
	err := core.LoadJsonFile(nodes, nodesSetupFile)
	require.Nil(tb, err)

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
		require.Nil(tb, errB)

		nodes.InitialNodes = append(nodes.InitialNodes, &sharding.InitialNode{
			PubKey:  hex.EncodeToString(pkBytes),
			Address: address,
		})
	}

	marshaledNodes, err := json.Marshal(nodes)
	require.Nil(tb, err)

	err = os.WriteFile(nodesSetupFile, marshaledNodes, 0644)
	require.Nil(tb, err)

	return privateKeys, publicKeys
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

func modifyFile(fileName string, f func(i []byte) ([]byte, error)) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	output := input
	if f != nil {
		output, err = f(input)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(fileName, output, 0644)
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
