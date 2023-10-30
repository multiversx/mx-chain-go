package configs

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"path"
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
	Configs               *config.Configs
	ValidatorsPrivateKeys []crypto.PrivateKey
}

func CreateChainSimulatorConfigs(tb testing.TB, args ArgsChainSimulatorConfigs) ArgsConfigsSimulator {
	configs := testscommon.CreateTestConfigs(tb, args.OriginalConfigsPath)

	// empty genesis smart contracts file
	modifyFile(tb, configs.ConfigurationPathsHolder.SmartContracts, func(intput []byte) []byte {
		return []byte("[]")
	})

	// generate validatos key and nodesSetup.json
	privateKeys, publicKeys := generateValidatorsKeyAndUpdateFiles(tb, configs, args.NumOfShards, args.GenesisAddressWithStake)

	// update genesis.json
	modifyFile(tb, configs.ConfigurationPathsHolder.Genesis, func(i []byte) []byte {
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

		addressesBytes, err := json.Marshal(addresses)
		require.Nil(tb, err)

		return addressesBytes
	})

	// generate validators.pem
	configs.ConfigurationPathsHolder.ValidatorKey = path.Join(args.OriginalConfigsPath, "validatorKey.pem")
	generateValidatorsPem(tb, configs.ConfigurationPathsHolder.ValidatorKey, publicKeys, privateKeys)

	return ArgsConfigsSimulator{
		Configs:               configs,
		ValidatorsPrivateKeys: privateKeys,
	}
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

func generateValidatorsPem(tb testing.TB, validatorsFile string, publicKeys []crypto.PublicKey, privateKey []crypto.PrivateKey) {
	validatorPubKeyConverter, err := pubkeyConverter.NewHexPubkeyConverter(96)
	require.Nil(tb, err)

	buff := bytes.Buffer{}
	for idx := 0; idx < len(publicKeys); idx++ {
		publicKeyBytes, errA := publicKeys[idx].ToByteArray()
		require.Nil(tb, errA)

		pkString, errE := validatorPubKeyConverter.Encode(publicKeyBytes)
		require.Nil(tb, errE)

		privateKeyBytes, errP := privateKey[idx].ToByteArray()
		require.Nil(tb, errP)

		blk := pem.Block{
			Type:  "PRIVATE KEY for " + pkString,
			Bytes: []byte(hex.EncodeToString(privateKeyBytes)),
		}

		err = pem.Encode(&buff, &blk)
		require.Nil(tb, errE)
	}

	err = os.WriteFile(validatorsFile, buff.Bytes(), 0644)
	require.Nil(tb, err)
}

func modifyFile(tb testing.TB, fileName string, f func(i []byte) []byte) {
	input, err := os.ReadFile(fileName)
	require.Nil(tb, err)

	output := input
	if f != nil {
		output = f(input)
	}

	err = os.WriteFile(fileName, output, 0644)
	require.Nil(tb, err)
}
