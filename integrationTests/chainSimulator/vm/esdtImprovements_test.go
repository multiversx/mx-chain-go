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
	dataVm "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
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
// 1.check that the metadata for all tokens is saved on the system account
// 2. wait for DynamicEsdtFlag activation
// 3. transfer the tokens to another account
// 4. check that the metadata for all tokens is saved on the system account
// 5. make an updateTokenID@tokenID function call on the ESDTSystem SC for all token types
// 6. check that the metadata for all tokens is saved on the system account
// 7. transfer the tokens to another account
// 8. check that the metaData for the NFT was removed from the system account and moved to the user account
// 9. check that the metaData for the rest of the tokens is still present on the system account and not on the userAccount
// 10. do the test for both intra and cross shard txs
func TestChainSimulator_CheckNFTandSFTMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// logger.SetLogLevel("*:TRACE")

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(4)

	baseIssuingCost := "1000"

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
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(staking.OneEGLD, mintValue)

	address, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	log.Info("Initial setup: Create an NFT and an SFT (before the activation of DynamicEsdtFlag)")

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issueNonFungible"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString([]byte("ASD"))),
		},
		[]byte("@"),
	)

	tx := &transaction.Transaction{
		Nonce:     0,
		SndAddr:   address.Bytes,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	fmt.Println(txResult)
	fmt.Println(txResult.Logs.Events[0])
	fmt.Println(txResult.Logs.Events[0].Identifier)
	tokenID := txResult.Logs.Events[0].Topics[0]

	log.Info("Issued token id", "tokenID", string(tokenID))

	roles := [][]byte{
		[]byte(core.ESDTMetaDataRecreate),
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTBurn),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdateAttributes),
		[]byte(core.ESDTRoleNFTAddURI),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	tokenType := core.DynamicNFTESDT

	txDataField = bytes.Join(
		[][]byte{
			[]byte(core.ESDTSetTokenType),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString([]byte(tokenType))),
		},
		[]byte("@"),
	)

	tx = &transaction.Transaction{
		Nonce:     0,
		SndAddr:   core.ESDTSCAddress,
		RcvAddr:   address.Bytes,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	nonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	name := []byte(hex.EncodeToString([]byte("name")))
	hash := []byte(hex.EncodeToString([]byte("hash")))
	attributes := []byte(hex.EncodeToString([]byte("attributes")))
	uris := []byte(hex.EncodeToString([]byte("uri")))

	expUris := [][]byte{[]byte(hex.EncodeToString([]byte("uri")))}

	txDataField = bytes.Join(
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

	tx = &transaction.Transaction{
		Nonce:     1,
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

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	log.Info("Step 1. check that the metadata for all tokens is saved on the system account")

	retrievedMetaData := getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	require.Equal(t, nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expUris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))

	log.Info("Step 2. wait for DynamicEsdtFlag activation")

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Step 3. transfer the tokens to another account")

	address2, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	// tx = utils.CreateESDTNFTTransferTx(
	// 	2,
	// 	address.Bytes,
	// 	address2.Bytes,
	// 	tokenID,
	// 	1,
	// 	big.NewInt(1),
	// 	minGasPrice,
	// 	10_000_000,
	// 	"",
	// )
	// tx.Version = 1
	// tx.Signature = []byte("dummySig")
	// tx.ChainID = []byte(configs.ChainID)

	// txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	// require.Nil(t, err)
	// require.NotNil(t, txResult)

	// fmt.Println(txResult.Logs.Events[0])
	// fmt.Println(txResult.Logs.Events[0].Topics[0])
	// fmt.Println(txResult.Logs.Events[0].Topics[1])
	// fmt.Println(string(txResult.Logs.Events[0].Topics[1]))

	require.Equal(t, "success", txResult.Status.String())

	log.Info("Step 4. check that the metadata for all tokens is saved on the system account")

	retrievedMetaData = getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	require.Equal(t, nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expUris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))

	log.Info("Step 5. make an updateTokenID@tokenID function call on the ESDTSystem SC for all token types")

	txDataField = []byte("updateTokenID@" + hex.EncodeToString(tokenID))

	tx = &transaction.Transaction{
		Nonce:     2,
		SndAddr:   address.Bytes,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	fmt.Println(txResult.Logs.Events[0])
	fmt.Println(txResult.Logs.Events[0].Topics[0])
	fmt.Println(txResult.Logs.Events[0].Topics[1])
	fmt.Println(string(txResult.Logs.Events[0].Topics[1]))

	require.Equal(t, "success", txResult.Status.String())

	log.Info("Step 6. check that the metadata for all tokens is saved on the system account")

	retrievedMetaData = getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	require.Equal(t, nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expUris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))

	log.Info("Step 7. transfer the tokens to another account")

	tx = utils.CreateESDTNFTTransferTx(
		3,
		address.Bytes,
		address2.Bytes,
		tokenID,
		1,
		big.NewInt(1),
		minGasPrice,
		10_000_000,
		"",
	)
	tx.Version = 1
	tx.Signature = []byte("dummySig")
	tx.ChainID = []byte(configs.ChainID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	fmt.Println(txResult.Logs.Events[0])
	fmt.Println(string(txResult.Logs.Events[0].Topics[0]))
	fmt.Println(string(txResult.Logs.Events[0].Topics[1]))

	require.Equal(t, "success", txResult.Status.String())

	log.Info("Step 8. check that the metaData for the NFT was removed from the system account and moved to the user account")

	shardID2 := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address2.Bytes)

	retrievedMetaData = getMetaDataFromAcc(t, cs, address2.Bytes, tokenID, shardID2)

	require.Equal(t, nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expUris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))
}

func executeQuery(cs testsChainSimulator.ChainSimulator, shardID uint32, scAddress []byte, funcName string, args [][]byte) (*dataVm.VMOutputApi, error) {
	output, _, err := cs.GetNodeHandler(shardID).GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  funcName,
		Arguments: args,
	})
	return output, err
}

func getMetaDataFromAcc(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
) *esdt.MetaData {
	account, err := cs.GetNodeHandler(shardID).GetStateComponents().AccountsAdapter().LoadAccount(addressBytes)
	require.Nil(t, err)
	userAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)

	baseEsdtKeyPrefix := core.ProtectedKeyPrefix + core.ESDTKeyIdentifier
	key := append([]byte(baseEsdtKeyPrefix), token...)

	key2 := append(key, big.NewInt(0).SetUint64(1).Bytes()...)
	esdtDataBytes, _, err := userAccount.RetrieveValue(key2)
	require.Nil(t, err)

	esdtData := &esdt.ESDigitalToken{}
	err = cs.GetNodeHandler(shardID).GetCoreComponents().InternalMarshalizer().Unmarshal(esdtData, esdtDataBytes)
	require.Nil(t, err)
	require.NotNil(t, esdtData.TokenMetaData)

	return esdtData.TokenMetaData
}

func setAddressEsdtRoles(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	address dtos.WalletAddress,
	token []byte,
	roles [][]byte,
) {
	marshaller := cs.GetNodeHandler(0).GetCoreComponents().InternalMarshalizer()

	rolesKey := append([]byte(core.ProtectedKeyPrefix), append([]byte(core.ESDTRoleIdentifier), []byte(core.ESDTKeyIdentifier)...)...)
	rolesKey = append(rolesKey, token...)

	rolesData := &esdt.ESDTRoles{
		Roles: roles,
	}

	rolesDataBytes, err := marshaller.Marshal(rolesData)
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
}

// Test scenario
//
// Initial setup: Create fungible, NFT,  SFT and metaESDT tokens
// (after the activation of DynamicEsdtFlag)
//
// 1. check that the metaData for the NFT was saved in the user account and not on the system account
// 2. check that the metaData for the other token types is saved on the system account and not at the user account level
func TestChainSimulator_CreateTokensAfterActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// logger.SetLogLevel("*:TRACE")

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(2)

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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Initial setup: Create fungible, NFT,  SFT and metaESDT tokens (after the activation of DynamicEsdtFlag)")

	tokenID := []byte("ASD-d31313")

	roles := [][]byte{
		[]byte(core.ESDTMetaDataRecreate),
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTBurn),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdateAttributes),
		[]byte(core.ESDTRoleNFTAddURI),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	nonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	name := []byte(hex.EncodeToString([]byte("name")))
	hash := []byte(hex.EncodeToString([]byte("hash")))
	attributes := []byte(hex.EncodeToString([]byte("attributes")))
	uris := []byte(hex.EncodeToString([]byte("uri")))

	expUris := [][]byte{[]byte(hex.EncodeToString([]byte("uri")))}

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

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	log.Info("Step 1. check that the metaData for the NFT was saved in the user account and not on the system account")

	retrievedMetaData := getMetaDataFromAcc(t, cs, address.Bytes, tokenID, shardID)

	require.Equal(t, nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expUris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))
}

// Test scenario
//
// Initial setup: Create NFT
//
// Call ESDTMetaDataRecreate to rewrite the meta data for the nft
// (The sender must have the ESDTMetaDataRecreate role)
func TestChainSimulator_NFT_ESDTMetaDataRecreate(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(2)

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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Initial setup: Create NFT")

	tokenID := []byte("ASD-d31313")

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTMetaDataRecreate),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	name := []byte(hex.EncodeToString([]byte("name")))
	hash := []byte(hex.EncodeToString([]byte("hash")))
	attributes := []byte(hex.EncodeToString([]byte("attributes")))
	uris := [][]byte{[]byte(hex.EncodeToString([]byte("uri")))}

	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())), // quantity
			name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			hash,
			attributes,
			uris[0],
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

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	log.Info("Call ESDTMetaDataRecreate to rewrite the meta data for the nft")

	nonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	name = []byte(hex.EncodeToString([]byte("name2")))
	hash = []byte(hex.EncodeToString([]byte("hash2")))
	attributes = []byte(hex.EncodeToString([]byte("attributes2")))

	txDataField = bytes.Join(
		[][]byte{
			[]byte(core.ESDTMetaDataRecreate),
			[]byte(hex.EncodeToString(tokenID)),
			nonce,
			name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			hash,
			attributes,
			uris[0],
		},
		[]byte("@"),
	)

	tx = &transaction.Transaction{
		Nonce:     1,
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

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	retrievedMetaData := getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	require.Equal(t, nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range uris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))
}

// Test scenario
//
// Initial setup: Create NFT
//
// Call ESDTMetaDataUpdate to update some of the meta data parameters
// (The sender must have the ESDTRoleNFTUpdate role)
func TestChainSimulator_NFT_ESDTMetaDataUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(2)

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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Initial setup: Create  NFT")

	tokenID := []byte("ASD-d31313")

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	name := []byte(hex.EncodeToString([]byte("name")))
	hash := []byte(hex.EncodeToString([]byte("hash")))
	attributes := []byte(hex.EncodeToString([]byte("attributes")))
	uris := [][]byte{[]byte(hex.EncodeToString([]byte("uri")))}

	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())), // quantity
			name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			hash,
			attributes,
			uris[0],
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

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	log.Info("Call ESDTMetaDataRecreate to rewrite the meta data for the nft")

	nonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	name = []byte(hex.EncodeToString([]byte("name2")))
	hash = []byte(hex.EncodeToString([]byte("hash2")))
	attributes = []byte(hex.EncodeToString([]byte("attributes2")))

	txDataField = bytes.Join(
		[][]byte{
			[]byte(core.ESDTMetaDataUpdate),
			[]byte(hex.EncodeToString(tokenID)),
			nonce,
			name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			hash,
			attributes,
			uris[0],
		},
		[]byte("@"),
	)

	tx = &transaction.Transaction{
		Nonce:     1,
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

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	retrievedMetaData := getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	require.Equal(t, nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range uris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))
}

// Test scenario
//
// Initial setup: Create NFT
//
// Call ESDTModifyCreator and check that the creator was modified
// (The sender must have the ESDTRoleModifyCreator role)
func TestChainSimulator_NFT_ESDTModifyCreator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(2)

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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Initial setup: Create NFT")

	tokenID := []byte("ASD-d31313")

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	name := []byte(hex.EncodeToString([]byte("name")))
	hash := []byte(hex.EncodeToString([]byte("hash")))
	attributes := []byte(hex.EncodeToString([]byte("attributes")))
	uris := [][]byte{[]byte(hex.EncodeToString([]byte("uri")))}

	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())), // quantity
			name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			hash,
			attributes,
			uris[0],
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

	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	log.Info("Call ESDTModifyCreator and check that the creator was modified")

	newCreatorAddress, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	roles = [][]byte{
		[]byte(core.ESDTRoleModifyCreator),
	}
	setAddressEsdtRoles(t, cs, newCreatorAddress, tokenID, roles)

	nonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	txDataField = bytes.Join(
		[][]byte{
			[]byte(core.ESDTModifyCreator),
			[]byte(hex.EncodeToString(tokenID)),
			nonce,
		},
		[]byte("@"),
	)

	tx = &transaction.Transaction{
		Nonce:     0,
		SndAddr:   newCreatorAddress.Bytes,
		RcvAddr:   newCreatorAddress.Bytes,
		GasLimit:  10_000_000,
		GasPrice:  minGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	retrievedMetaData := getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	require.Equal(t, newCreatorAddress.Bytes, retrievedMetaData.Creator)
}

// Test scenario
//
// Initial setup: Create NFT
//
// Call ESDTSetNewURIs and check that the new URIs were set for the NFT
// (The sender must have the ESDTRoleSetNewURI role)
func TestChainSimulator_NFT_ESDTSetNewURIs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(2)

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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Initial setup: Create NFT")

	tokenID := []byte("ASD-d31313")

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	name := []byte(hex.EncodeToString([]byte("name")))
	hash := []byte(hex.EncodeToString([]byte("hash")))
	attributes := []byte(hex.EncodeToString([]byte("attributes")))
	uris := [][]byte{[]byte(hex.EncodeToString([]byte("uri")))}

	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())), // quantity
			name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			hash,
			attributes,
			uris[0],
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

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	log.Info("Call ESDTSetNewURIs and check that the new URIs were set for the NFT")

	roles = [][]byte{
		[]byte(core.ESDTRoleSetNewURI),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	nonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	uris = [][]byte{
		[]byte(hex.EncodeToString([]byte("uri0"))),
		[]byte(hex.EncodeToString([]byte("uri1"))),
		[]byte(hex.EncodeToString([]byte("uri2"))),
	}

	expUris := [][]byte{
		[]byte("uri0"),
		[]byte("uri1"),
		[]byte("uri2"),
	}

	txDataField = bytes.Join(
		[][]byte{
			[]byte(core.ESDTSetNewURIs),
			[]byte(hex.EncodeToString(tokenID)),
			nonce,
			uris[0],
			uris[1],
			uris[2],
		},
		[]byte("@"),
	)

	tx = &transaction.Transaction{
		Nonce:     1,
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

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	retrievedMetaData := getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	require.Equal(t, expUris, retrievedMetaData.URIs)
}

// Test scenario
//
// Initial setup: Create NFT
//
// Call ESDTModifyRoyalties and check that the royalties were changed
// (The sender must have the ESDTRoleModifyRoyalties role)
func TestChainSimulator_NFT_ESDTModifyRoyalties(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	activationEpoch := uint32(2)

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

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	log.Info("Initial setup: Create NFT")

	tokenID := []byte("ASD-d31313")

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	name := []byte(hex.EncodeToString([]byte("name")))
	hash := []byte(hex.EncodeToString([]byte("hash")))
	attributes := []byte(hex.EncodeToString([]byte("attributes")))
	uris := [][]byte{[]byte(hex.EncodeToString([]byte("uri")))}

	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())), // quantity
			name,
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			hash,
			attributes,
			uris[0],
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

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	log.Info("Call ESDTModifyRoyalties and check that the royalties were changed")

	roles = [][]byte{
		[]byte(core.ESDTRoleModifyRoyalties),
	}
	setAddressEsdtRoles(t, cs, address, tokenID, roles)

	nonce := []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	royalties := []byte(hex.EncodeToString(big.NewInt(20).Bytes()))

	txDataField = bytes.Join(
		[][]byte{
			[]byte(core.ESDTModifyRoyalties),
			[]byte(hex.EncodeToString(tokenID)),
			nonce,
			royalties,
		},
		[]byte("@"),
	)

	tx = &transaction.Transaction{
		Nonce:     1,
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

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	retrievedMetaData := getMetaDataFromAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	require.Equal(t, uint32(big.NewInt(20).Uint64()), retrievedMetaData.Royalties)
}
