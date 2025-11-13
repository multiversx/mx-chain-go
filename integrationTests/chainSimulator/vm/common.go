package vm

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	DefaultPathToInitialConfig = "../../../../cmd/node/config/"

	MinGasPrice                            = 1000000000
	MaxNumOfBlockToGenerateWhenExecutingTx = 7
)

var (
	RoundDurationInMillis          = uint64(6000)
	SupernovaRoundDurationInMillis = uint64(600)
	RoundsPerEpoch                 = core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	SupernovaRoundsPerEpoch = core.OptionalUint64{
		HasValue: true,
		Value:    200,
	}
)

var OneEGLD = big.NewInt(1000000000000000000)

var Log = logger.GetOrCreate("integrationTests/chainSimulator/vm")

func CreateAddresses(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	isCrossShard bool,
) []dtos.WalletAddress {
	var shardIDs []uint32
	if !isCrossShard {
		shardIDs = []uint32{1, 1, 1}
	} else {
		shardIDs = []uint32{0, 1, 2}
	}

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(OneEGLD, mintValue)

	address, err := cs.GenerateAndMintWalletAddress(shardIDs[0], mintValue)
	require.Nil(t, err)

	address2, err := cs.GenerateAndMintWalletAddress(shardIDs[1], mintValue)
	require.Nil(t, err)

	address3, err := cs.GenerateAndMintWalletAddress(shardIDs[2], mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	return []dtos.WalletAddress{address, address2, address3}
}

func IssueTx(nonce uint64, sndAdr []byte, ticker []byte, baseIssuingCost string) *transaction.Transaction {
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issue"),
			[]byte(hex.EncodeToString([]byte("asdname1"))),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func IssueMetaESDTTx(nonce uint64, sndAdr []byte, ticker []byte, baseIssuingCost string) *transaction.Transaction {
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("registerMetaESDT"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString(ticker)),
			[]byte(hex.EncodeToString(big.NewInt(10).Bytes())),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func IssueNonFungibleTx(nonce uint64, sndAdr []byte, ticker []byte, baseIssuingCost string) *transaction.Transaction {
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issueNonFungible"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString(ticker)),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func IssueSemiFungibleTx(nonce uint64, sndAdr []byte, ticker []byte, baseIssuingCost string) *transaction.Transaction {
	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issueSemiFungible"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString(ticker)),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func ChangeToDynamicTx(nonce uint64, sndAdr []byte, tokenID []byte) *transaction.Transaction {
	txDataField := []byte("changeToDynamic@" + hex.EncodeToString(tokenID))

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func UpdateTokenIDTx(nonce uint64, sndAdr []byte, tokenID []byte) *transaction.Transaction {
	txDataField := []byte("updateTokenID@" + hex.EncodeToString(tokenID))

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func EsdtNftCreateTx(
	nonce uint64,
	sndAdr []byte,
	tokenID []byte,
	metaData *txsFee.MetaData,
	quantity int64,
) *transaction.Transaction {
	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.BuiltInFunctionESDTNFTCreate),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(quantity).Bytes())), // quantity
			metaData.Name,
			metaData.Royalties,
			metaData.Hash,
			metaData.Attributes,
			metaData.Uris[0],
			metaData.Uris[1],
			metaData.Uris[2],
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   sndAdr,
		GasLimit:  10_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func ModifyCreatorTx(
	nonce uint64,
	sndAdr []byte,
	tokenID []byte,
) *transaction.Transaction {
	txDataField := bytes.Join(
		[][]byte{
			[]byte(core.ESDTModifyCreator),
			[]byte(hex.EncodeToString(tokenID)),
			[]byte(hex.EncodeToString(big.NewInt(1).Bytes())),
		},
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAdr,
		RcvAddr:   sndAdr,
		GasLimit:  10_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func GetESDTDataFromAcc(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
) *esdt.ESDigitalToken {
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

	return esdtData
}

func GetMetaDataFromAcc(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
) *esdt.MetaData {
	esdtData := GetESDTDataFromAcc(t, cs, addressBytes, token, shardID)

	require.NotNil(t, esdtData.TokenMetaData)

	return esdtData.TokenMetaData
}

func SetSpecialRoleTx(
	nonce uint64,
	sndAddr []byte,
	address []byte,
	token []byte,
	roles [][]byte,
) *transaction.Transaction {
	txDataBytes := [][]byte{
		[]byte("setSpecialRole"),
		[]byte(hex.EncodeToString(token)),
		[]byte(hex.EncodeToString(address)),
	}

	for _, role := range roles {
		txDataBytes = append(txDataBytes, []byte(hex.EncodeToString(role)))
	}

	txDataField := bytes.Join(
		txDataBytes,
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAddr,
		RcvAddr:   vm.ESDTSCAddress,
		GasLimit:  60_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
}

func SetAddressEsdtRoles(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	nonce uint64,
	address dtos.WalletAddress,
	token []byte,
	roles [][]byte,
) {
	tx := SetSpecialRoleTx(nonce, address.Bytes, address.Bytes, token, roles)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())
}

type IssueTxFunc func(uint64, []byte, []byte, string) *transaction.Transaction

func EsdtNFTTransferTx(nonce uint64, sndAdr, rcvAddr, token []byte) *transaction.Transaction {
	tx := utils.CreateESDTNFTTransferTx(
		nonce,
		sndAdr,
		rcvAddr,
		token,
		1,
		big.NewInt(1),
		MinGasPrice,
		10_000_000,
		"",
	)
	tx.Version = 1
	tx.Signature = []byte("dummySig")
	tx.ChainID = []byte(configs.ChainID)

	return tx
}

func CheckMetaDataNotInAcc(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
) {
	esdtData := GetESDTDataFromAcc(t, cs, addressBytes, token, shardID)

	require.Nil(t, esdtData.TokenMetaData)
}

func MultiESDTNFTTransferTx(nonce uint64, sndAdr, rcvAddr []byte, tokens [][]byte) *transaction.Transaction {
	transferData := make([]*utils.TransferESDTData, 0)

	for _, tokenID := range tokens {
		transferData = append(transferData, &utils.TransferESDTData{
			Token: tokenID,
			Nonce: 1,
			Value: big.NewInt(1),
		})
	}

	tx := utils.CreateMultiTransferTX(
		nonce,
		sndAdr,
		rcvAddr,
		MinGasPrice,
		10_000_000,
		transferData...,
	)
	tx.Version = 1
	tx.Signature = []byte("dummySig")
	tx.ChainID = []byte(configs.ChainID)

	return tx
}

func CheckMetaData(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	token []byte,
	shardID uint32,
	expectedMetaData *txsFee.MetaData,
) {
	retrievedMetaData := GetMetaDataFromAcc(t, cs, addressBytes, token, shardID)

	require.Equal(t, expectedMetaData.Nonce, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Nonce)).Bytes())))
	require.Equal(t, expectedMetaData.Name, []byte(hex.EncodeToString(retrievedMetaData.Name)))
	require.Equal(t, expectedMetaData.Royalties, []byte(hex.EncodeToString(big.NewInt(int64(retrievedMetaData.Royalties)).Bytes())))
	require.Equal(t, expectedMetaData.Hash, []byte(hex.EncodeToString(retrievedMetaData.Hash)))
	for i, uri := range expectedMetaData.Uris {
		require.Equal(t, uri, []byte(hex.EncodeToString(retrievedMetaData.URIs[i])))
	}
	require.Equal(t, expectedMetaData.Attributes, []byte(hex.EncodeToString(retrievedMetaData.Attributes)))
}

func CheckReservedField(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	addressBytes []byte,
	tokenID []byte,
	shardID uint32,
	expectedReservedField []byte,
) {
	esdtData := GetESDTDataFromAcc(t, cs, addressBytes, tokenID, shardID)
	require.Equal(t, expectedReservedField, esdtData.Reserved)
}

func GetTestChainSimulatorWithDynamicNFTEnabled(t *testing.T, baseIssuingCost string) (testsChainSimulator.ChainSimulator, int32) {
	activationEpochForDynamicNFT := uint32(2)

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          RoundDurationInMillis,
		SupernovaRoundDurationInMillis: SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListMeta:        0,
		NumNodesWaitingListShard:       0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpochForDynamicNFT
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = integrationTests.UnreachableEpoch // TODO: handle supernova activation with transition
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpochForDynamicNFT))
	require.Nil(t, err)

	return cs, int32(activationEpochForDynamicNFT)
}

func GetTestChainSimulatorWithSaveToSystemAccountDisabled(t *testing.T, baseIssuingCost string) (testsChainSimulator.ChainSimulator, int32) {
	activationEpochForSaveToSystemAccount := uint32(4)
	activationEpochForDynamicNFT := uint32(6)

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            DefaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          RoundDurationInMillis,
		SupernovaRoundDurationInMillis: SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        SupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListMeta:        0,
		NumNodesWaitingListShard:       0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.OptimizeNFTStoreEnableEpoch = activationEpochForSaveToSystemAccount
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpochForDynamicNFT
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpochForSaveToSystemAccount) - 2)
	require.Nil(t, err)

	return cs, int32(activationEpochForDynamicNFT)
}

func EsdtMetaDataUpdateTx(tokenID []byte, metaData *txsFee.MetaData, nonce uint64, address []byte) *transaction.Transaction {
	txData := [][]byte{
		[]byte(core.ESDTMetaDataUpdate),
		[]byte(hex.EncodeToString(tokenID)),
		metaData.Nonce,
		metaData.Name,
		metaData.Royalties,
		metaData.Hash,
		metaData.Attributes,
	}
	if len(metaData.Uris) > 0 {
		txData = append(txData, metaData.Uris...)
	} else {
		txData = append(txData, nil)
	}

	txDataField := bytes.Join(
		txData,
		[]byte("@"),
	)

	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   address,
		RcvAddr:   address,
		GasLimit:  10_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	return tx
}

func CreateTokenUpdateTokenIDAndTransfer(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	originAddress []byte,
	targetAddress []byte,
	tokenID []byte,
	metaData *txsFee.MetaData,
	epochForDynamicNFT int32,
	walletWithRoles dtos.WalletAddress,
) {
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, 1, walletWithRoles, tokenID, roles)

	tx := EsdtNftCreateTx(2, originAddress, tokenID, metaData, 1)

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	Log.Info("check that the metadata is saved on the user account")
	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(originAddress)
	CheckMetaData(t, cs, originAddress, tokenID, shardID, metaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, tokenID, shardID)

	err = cs.GenerateBlocksUntilEpochIsReached(epochForDynamicNFT)
	require.Nil(t, err)

	tx = UpdateTokenIDTx(3, originAddress, tokenID)

	Log.Info("updating token id", "tokenID", tokenID)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("transferring token id", "tokenID", tokenID)

	tx = EsdtNFTTransferTx(4, originAddress, targetAddress, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
}
