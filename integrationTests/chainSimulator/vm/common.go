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
	chainSimulator2 "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/state"
	vm2 "github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	DefaultPathToInitialConfig = "../../../cmd/node/config/"

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

	OneEGLD = big.NewInt(1000000000000000000)

	Log = logger.GetOrCreate("integrationTests/chainSimulator/vm")
)

func TransferAndCheckTokensMetaData(t *testing.T, isCrossShard bool, isMultiTransfer bool) {
	activationEpoch := uint32(4)

	baseIssuingCost := "1000"

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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = activationEpoch
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	addrs := CreateAddresses(t, cs, isCrossShard)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch) - 1)
	require.Nil(t, err)

	Log.Info("Initial setup: Create NFT, SFT and metaESDT tokens (before the activation of DynamicEsdtFlag)")

	// issue metaESDT
	metaESDTTicker := []byte("METATICKER")
	nonce := uint64(0)
	tx := IssueMetaESDTTx(nonce, addrs[0].Bytes, metaESDTTicker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	metaESDTTokenID := txResult.Logs.Events[0].Topics[0]

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
	}
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], metaESDTTokenID, roles)
	nonce++

	rolesTransfer := [][]byte{[]byte(core.ESDTRoleTransfer)}
	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, metaESDTTokenID, rolesTransfer)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued metaESDT token id", "tokenID", string(metaESDTTokenID))

	// issue NFT
	nftTicker := []byte("NFTTICKER")
	tx = IssueNonFungibleTx(nonce, addrs[0].Bytes, nftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	nftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], nftTokenID, roles)
	nonce++

	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, nftTokenID, rolesTransfer)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued NFT token id", "tokenID", string(nftTokenID))

	// issue SFT
	sftTicker := []byte("SFTTICKER")
	tx = IssueSemiFungibleTx(nonce, addrs[0].Bytes, sftTicker, baseIssuingCost)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sftTokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[0], sftTokenID, roles)
	nonce++

	tx = SetSpecialRoleTx(nonce, addrs[0].Bytes, addrs[1].Bytes, sftTokenID, rolesTransfer)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Issued SFT token id", "tokenID", string(sftTokenID))

	nftMetaData := txsFee.GetDefaultMetaData()
	nftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData := txsFee.GetDefaultMetaData()
	sftMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	esdtMetaData := txsFee.GetDefaultMetaData()
	esdtMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tokenIDs := [][]byte{
		nftTokenID,
		sftTokenID,
		metaESDTTokenID,
	}

	tokensMetadata := []*txsFee.MetaData{
		nftMetaData,
		sftMetaData,
		esdtMetaData,
	}

	for i := range tokenIDs {
		tx = EsdtNftCreateTx(nonce, addrs[0].Bytes, tokenIDs[i], tokensMetadata[i], 1)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("Step 1. check that the metadata for all tokens is saved on the system account")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)

	CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
	CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	CheckMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)

	Log.Info("Step 2. wait for DynamicEsdtFlag activation")

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	Log.Info("Step 3. transfer the tokens to another account")

	if isMultiTransfer {
		tx = MultiESDTNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, tokenIDs)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())

		nonce++
	} else {
		for _, tokenID := range tokenIDs {
			Log.Info("transfering token id", "tokenID", tokenID)

			tx = EsdtNFTTransferTx(nonce, addrs[0].Bytes, addrs[1].Bytes, tokenID)
			txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
			require.Nil(t, err)
			require.NotNil(t, txResult)
			require.Equal(t, "success", txResult.Status.String())

			nonce++
		}
	}

	Log.Info("Step 4. check that the metadata for all tokens is saved on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[1].Bytes)

	CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
	CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	CheckMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)

	Log.Info("Step 5. make an updateTokenID@tokenID function call on the ESDTSystem SC for all token types")

	for _, tokenID := range tokenIDs {
		tx = UpdateTokenIDTx(nonce, addrs[0].Bytes, tokenID)

		Log.Info("updating token id", "tokenID", tokenID)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)
		require.Equal(t, "success", txResult.Status.String())

		nonce++
	}

	Log.Info("Step 6. check that the metadata for all tokens is saved on the system account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	CheckMetaData(t, cs, core.SystemAccountAddress, nftTokenID, shardID, nftMetaData)
	CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	CheckMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)

	Log.Info("Step 7. transfer the tokens to another account")

	nonce = uint64(0)
	if isMultiTransfer {
		tx = MultiESDTNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, tokenIDs)

		txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.NotNil(t, txResult)

		require.Equal(t, "success", txResult.Status.String())
	} else {
		for _, tokenID := range tokenIDs {
			Log.Info("transfering token id", "tokenID", tokenID)

			tx = EsdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, tokenID)
			txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
			require.Nil(t, err)
			require.NotNil(t, txResult)
			require.Equal(t, "success", txResult.Status.String())

			nonce++
		}
	}

	Log.Info("Step 8. check that the metaData for the NFT was removed from the system account and moved to the user account")

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	shardID = cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)

	CheckMetaData(t, cs, addrs[2].Bytes, nftTokenID, shardID, nftMetaData)
	CheckMetaDataNotInAcc(t, cs, core.SystemAccountAddress, nftTokenID, shardID)

	Log.Info("Step 9. check that the metaData for the rest of the tokens is still present on the system account and not on the userAccount")

	CheckMetaData(t, cs, core.SystemAccountAddress, sftTokenID, shardID, sftMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[2].Bytes, sftTokenID, shardID)

	CheckMetaData(t, cs, core.SystemAccountAddress, metaESDTTokenID, shardID, esdtMetaData)
	CheckMetaDataNotInAcc(t, cs, addrs[2].Bytes, metaESDTTokenID, shardID)
}

func CreateAddresses(
	t *testing.T,
	cs chainSimulator2.ChainSimulator,
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

func CheckMetaData(
	t *testing.T,
	cs chainSimulator2.ChainSimulator,
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
	cs chainSimulator2.ChainSimulator,
	addressBytes []byte,
	tokenID []byte,
	shardID uint32,
	expectedReservedField []byte,
) {
	esdtData := GetESDTDataFromAcc(t, cs, addressBytes, tokenID, shardID)
	require.Equal(t, expectedReservedField, esdtData.Reserved)
}

func CheckMetaDataNotInAcc(
	t *testing.T,
	cs chainSimulator2.ChainSimulator,
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
		RcvAddr:   vm2.ESDTSCAddress,
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
		RcvAddr:   vm2.ESDTSCAddress,
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
	cs chainSimulator2.ChainSimulator,
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
	cs chainSimulator2.ChainSimulator,
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
		RcvAddr:   vm2.ESDTSCAddress,
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
	cs chainSimulator2.ChainSimulator,
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

func TestChainSimulatorChangeMetaData(t *testing.T, issueFn IssueTxFunc) {
	baseIssuingCost := "1000"

	cs, _ := GetTestChainSimulatorWithDynamicNFTEnabled(t, baseIssuingCost)
	defer cs.Close()

	addrs := CreateAddresses(t, cs, true)

	Log.Info("Initial setup: Create token and send in another shard")

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTAddQuantity),
	}

	ticker := []byte("TICKER")
	nonce := uint64(0)
	tx := issueFn(nonce, addrs[1].Bytes, ticker, baseIssuingCost)
	nonce++

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	tokenID := txResult.Logs.Events[0].Topics[0]
	SetAddressEsdtRoles(t, cs, nonce, addrs[1], tokenID, roles)
	nonce++

	Log.Info("Issued token id", "tokenID", string(tokenID))

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	tx = EsdtNftCreateTx(nonce, addrs[1].Bytes, tokenID, metaData, 2)
	nonce++

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	tx = ChangeToDynamicTx(nonce, addrs[1].Bytes, tokenID)
	nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("Send to separate shards")

	tx = EsdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[2].Bytes, tokenID)
	nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	tx = EsdtNFTTransferTx(nonce, addrs[1].Bytes, addrs[0].Bytes, tokenID)
	nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)

	require.Equal(t, "success", txResult.Status.String())

	roles = [][]byte{
		[]byte(core.ESDTRoleTransfer),
		[]byte(core.ESDTRoleNFTUpdate),
	}
	tx = SetSpecialRoleTx(nonce, addrs[1].Bytes, addrs[0].Bytes, tokenID, roles)

	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	Log.Info("Step 1. change the sft meta data in one shard")

	sftMetaData2 := txsFee.GetDefaultMetaData()
	sftMetaData2.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))

	sftMetaData2.Name = []byte(hex.EncodeToString([]byte("name2")))
	sftMetaData2.Hash = []byte(hex.EncodeToString([]byte("hash2")))
	sftMetaData2.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))

	tx = EsdtMetaDataUpdateTx(tokenID, sftMetaData2, 0, addrs[0].Bytes)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	Log.Info("Step 2. check that the newest metadata is saved")

	shardID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[0].Bytes)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, shardID, sftMetaData2)

	shard2ID := cs.GetNodeHandler(0).GetProcessComponents().ShardCoordinator().ComputeId(addrs[2].Bytes)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, shard2ID, metaData)

	Log.Info("Step 3. create new wallet is shard 2")

	mintValue := big.NewInt(10)
	mintValue = mintValue.Mul(OneEGLD, mintValue)
	newShard2Addr, err := cs.GenerateAndMintWalletAddress(2, mintValue)
	require.Nil(t, err)
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	Log.Info("Step 4. send updated token to shard 2 ")

	tx = EsdtNFTTransferTx(1, addrs[0].Bytes, newShard2Addr.Bytes, tokenID)
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	Log.Info("Step 5. check meta data in shard 2 is updated to latest version ")

	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, shard2ID, sftMetaData2)
	CheckMetaData(t, cs, core.SystemAccountAddress, tokenID, shardID, sftMetaData2)

}

func CheckTokenRoles(t *testing.T, returnData [][]byte, expectedRoles [][]byte) {
	for _, expRole := range expectedRoles {
		found := false

		for _, item := range returnData {
			if bytes.Equal(expRole, item) {
				found = true
			}
		}

		require.True(t, found)
	}
}

func GetTestChainSimulatorWithDynamicNFTEnabled(t *testing.T, baseIssuingCost string) (chainSimulator2.ChainSimulator, int32) {
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

func GetTestChainSimulatorWithSaveToSystemAccountDisabled(t *testing.T, baseIssuingCost string) (chainSimulator2.ChainSimulator, int32) {
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

func CreateTokenUpdateTokenIDAndTransfer(
	t *testing.T,
	cs chainSimulator2.ChainSimulator,
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

func UnsetSpecialRole(
	nonce uint64,
	sndAddr []byte,
	address []byte,
	token []byte,
	role []byte,
) *transaction.Transaction {
	txDataBytes := [][]byte{
		[]byte("unSetSpecialRole"),
		[]byte(hex.EncodeToString(token)),
		[]byte(hex.EncodeToString(address)),
		[]byte(hex.EncodeToString(role)),
	}

	txDataField := bytes.Join(
		txDataBytes,
		[]byte("@"),
	)

	return &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   sndAddr,
		RcvAddr:   vm2.ESDTSCAddress,
		GasLimit:  60_000_000,
		GasPrice:  MinGasPrice,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     big.NewInt(0),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
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

func TransferSpecialRoleToAddr(
	t *testing.T,
	cs chainSimulator2.ChainSimulator,
	nonce uint64,
	tokenID []byte,
	sndAddr []byte,
	dstAddr []byte,
	role []byte,
) uint64 {
	tx := UnsetSpecialRole(nonce, sndAddr, sndAddr, tokenID, role)
	nonce++
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	tx = SetSpecialRoleTx(nonce, sndAddr, dstAddr, tokenID, [][]byte{role})
	nonce++
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	return nonce
}
