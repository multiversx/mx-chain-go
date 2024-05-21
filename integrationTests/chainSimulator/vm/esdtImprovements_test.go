package vm

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"

	minGasPrice = 1000000000
)

var log = logger.GetOrCreate("integrationTests/chainSimulator/vm")

// Test scenario
//
// Initial setup: Create an NFT and an SFT (before the activation of DynamicEsdtFlag)
//
// 1. check that the metadata for nft and sfts are saved on the system account
// 2. wait for DynamicEsdtFlag activation
// 3. transfer the NFT and the SFT to another account
// 4. check that the metadata for nft is saved to the receiver account
// 5. check that the metadata for the sft is saved on the system account
// 6. repeat 3-5 for both intra and cross shard
func TestChainSimulator_CheckNFTMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(4)

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 numOfShards,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            3,
		MetaChainMinNodes:           3,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		NumNodesWaitingListMeta:     0,
		NumNodesWaitingListShard:    0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(staking.OneEGLD, mintValue)

	address, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	// set account esdt roles
	tokenID := []byte("ASD-d31313")

	roles := [][]byte{[]byte(core.ESDTMetaDataRecreate), []byte(core.ESDTRoleNFTCreate)}
	rolesKey := append([]byte(core.ProtectedKeyPrefix), append([]byte(core.ESDTRoleIdentifier), []byte(core.ESDTKeyIdentifier)...)...)
	rolesKey = append(rolesKey, tokenID...)

	rolesData := &esdt.ESDTRoles{
		Roles: roles,
	}

	rolesDataBytes, err := cs.GetNodeHandler(shardID).GetCoreComponents().InternalMarshalizer().Marshal(rolesData)
	require.Nil(t, err)

	keys := make(map[string]string)
	keys[hex.EncodeToString(rolesKey)] = hex.EncodeToString(rolesDataBytes)

	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: address.Bech32,
			Balance: "10000000000000000000000",
			Keys:    keys,
		},
	})
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	privateKey, _, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	nonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	name := []byte(hex.EncodeToString([]byte("name")))
	hash := []byte(hex.EncodeToString([]byte("hash")))
	attributes := []byte(hex.EncodeToString([]byte("attributes")))
	uris := []byte(hex.EncodeToString([]byte("uri")))

	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())), // quantity
			name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			hash,
			attributes,
			uris,
		},
		[]byte("@"),
	)

	tx := &transaction.Transaction{
		Nonce:     0,
		SndAddr:   address.Bytes,
		RcvAddr:   address.Bytes,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	account, err := cs.GetNodeHandler(shardID).GetStateComponents().AccountsAdapter().LoadAccount(core.SystemAccountAddress)
	require.Nil(t, err)
	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	baseEsdtKeyPrefix := core.ProtectedKeyPrefix + core.ESDTKeyIdentifier
	key := append([]byte(baseEsdtKeyPrefix), tokenID...)

	fmt.Println(userAccount)

	key2 := append(key, big.NewInt(0).SetUint64(1).Bytes()...)
	esdtDataBytes, _, err := userAccount.RetrieveValue(key2)
	require.Nil(t, err)
	esdtData := &esdt.ESDigitalToken{}

	err = cs.GetNodeHandler(shardID).GetCoreComponents().InternalMarshalizer().Unmarshal(esdtData, esdtDataBytes)
	require.Nil(t, err)

	require.NotNil(t, esdtData.TokenMetaData)
	fmt.Println(esdtData.TokenMetaData)

	expUris := [][]byte{[]byte(hex.EncodeToString([]byte("uri")))}

	retrievedMetaData := esdtData.TokenMetaData

	require.Equal(t, nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	// require.Equal(t, expectedMetaData.royalties, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Royalties)).Bytes())))
	require.Equal(t, hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expUris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))
}
