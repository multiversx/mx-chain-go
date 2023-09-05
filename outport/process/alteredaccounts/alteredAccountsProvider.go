package alteredaccounts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts/shared"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var (
	log = logger.GetOrCreate("outport/process/alteredaccounts")
)

type markedAlteredAccountToken struct {
	identifier  string
	nonce       uint64
	isNFTCreate bool
}

type markedAlteredAccount struct {
	balanceChanged bool
	isSender       bool
	tokens         map[string]*markedAlteredAccountToken
}

// ArgsAlteredAccountsProvider holds the arguments needed for creating a new instance of alteredAccountsProvider
type ArgsAlteredAccountsProvider struct {
	ShardCoordinator       sharding.Coordinator
	AddressConverter       core.PubkeyConverter
	AccountsDB             state.AccountsAdapter
	EsdtDataStorageHandler vmcommon.ESDTNFTStorageHandler
}

type alteredAccountsProvider struct {
	shardCoordinator       sharding.Coordinator
	addressConverter       core.PubkeyConverter
	accountsDB             state.AccountsAdapter
	tokensProc             *tokensProcessor
	esdtDataStorageHandler vmcommon.ESDTNFTStorageHandler
	mutExtractAccounts     sync.Mutex
}

// NewAlteredAccountsProvider returns a new instance of alteredAccountsProvider
func NewAlteredAccountsProvider(args ArgsAlteredAccountsProvider) (*alteredAccountsProvider, error) {
	err := checkArgAlteredAccountsProvider(args)
	if err != nil {
		return nil, err
	}

	return &alteredAccountsProvider{
		shardCoordinator:       args.ShardCoordinator,
		addressConverter:       args.AddressConverter,
		accountsDB:             args.AccountsDB,
		tokensProc:             newTokensProcessor(args.ShardCoordinator),
		esdtDataStorageHandler: args.EsdtDataStorageHandler,
	}, nil
}

// ExtractAlteredAccountsFromPool will extract and return altered accounts from the pool
func (aap *alteredAccountsProvider) ExtractAlteredAccountsFromPool(txPool *outportcore.TransactionPool, options shared.AlteredAccountsOptions) (map[string]*alteredAccount.AlteredAccount, error) {
	if err := options.Verify(); err != nil {
		return nil, err
	}

	aap.mutExtractAccounts.Lock()
	defer aap.mutExtractAccounts.Unlock()

	if txPool == nil {
		log.Warn("alteredAccountsProvider: ExtractAlteredAccountsFromPool", "txPool is nil", "will return")
		return map[string]*alteredAccount.AlteredAccount{}, nil
	}

	markedAccounts := make(map[string]*markedAlteredAccount)
	aap.extractAddressesWithBalanceChange(txPool, markedAccounts)
	err := aap.tokensProc.extractESDTAccounts(txPool, markedAccounts)
	if err != nil {
		return nil, err
	}

	return aap.fetchDataForMarkedAccounts(markedAccounts, options)
}

func (aap *alteredAccountsProvider) fetchDataForMarkedAccounts(markedAccounts map[string]*markedAlteredAccount, options shared.AlteredAccountsOptions) (map[string]*alteredAccount.AlteredAccount, error) {
	alteredAccounts := make(map[string]*alteredAccount.AlteredAccount)
	var err error
	for address, markedAccount := range markedAccounts {
		err = aap.processMarkedAccountData(address, markedAccount, alteredAccounts, options)
		if err != nil {
			return nil, err
		}
	}

	return alteredAccounts, nil
}

func (aap *alteredAccountsProvider) processMarkedAccountData(
	addressStr string,
	markedAccount *markedAlteredAccount,
	alteredAccounts map[string]*alteredAccount.AlteredAccount,
	options shared.AlteredAccountsOptions,
) error {
	addressBytes := []byte(addressStr)
	encodedAddress := aap.addressConverter.SilentEncode(addressBytes, log)

	userAccount, err := aap.loadUserAccount(addressBytes, options)
	if err != nil {
		return fmt.Errorf("%w while loading account when computing altered accounts. address: %s", err, encodedAddress)
	}

	alteredAccounts[encodedAddress] = aap.getAlteredAccountFromUserAccounts(encodedAddress, userAccount)

	if options.WithAdditionalOutportData {
		aap.addAdditionalDataInAlteredAccount(alteredAccounts[encodedAddress], userAccount, markedAccount)
	}

	for _, tokenData := range markedAccount.tokens {
		err = aap.addTokensDataForMarkedAccount(encodedAddress, userAccount, tokenData, alteredAccounts, options)
		if err != nil {
			return fmt.Errorf("%w while fetching token data when computing altered accounts", err)
		}
	}

	return nil
}

func (aap *alteredAccountsProvider) addAdditionalDataInAlteredAccount(alteredAcc *alteredAccount.AlteredAccount, userAccount state.UserAccountHandler, markedAccount *markedAlteredAccount) {
	alteredAcc.AdditionalData = &alteredAccount.AdditionalAccountData{
		IsSender:       markedAccount.isSender,
		BalanceChanged: markedAccount.balanceChanged,
		UserName:       string(userAccount.GetUserName()),
	}

	ownerAddressBytes := userAccount.GetOwnerAddress()
	if core.IsSmartContractAddress(userAccount.AddressBytes()) && len(ownerAddressBytes) == aap.addressConverter.Len() {
		alteredAcc.AdditionalData.CurrentOwner = aap.addressConverter.SilentEncode(ownerAddressBytes, log)
	}
	developerRewards := userAccount.GetDeveloperReward()
	if developerRewards != nil {
		alteredAcc.AdditionalData.DeveloperRewards = developerRewards.String()
	}
}

func (aap *alteredAccountsProvider) getAlteredAccountFromUserAccounts(userEncodedAddress string, userAccount state.UserAccountHandler) *alteredAccount.AlteredAccount {
	return &alteredAccount.AlteredAccount{
		Address: userEncodedAddress,
		Balance: userAccount.GetBalance().String(),
		Nonce:   userAccount.GetNonce(),
	}
}

func (aap *alteredAccountsProvider) loadUserAccount(addressBytes []byte, options shared.AlteredAccountsOptions) (state.UserAccountHandler, error) {
	var account vmcommon.AccountHandler
	var err error

	if !options.WithCustomAccountsRepository {
		account, err = aap.accountsDB.LoadAccount(addressBytes)
	} else {
		account, _, err = options.AccountsRepository.GetAccountWithBlockInfo(addressBytes, options.AccountQueryOptions)
	}

	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, errCannotCastToUserAccountHandler
	}

	return userAccount, nil
}

func (aap *alteredAccountsProvider) addTokensDataForMarkedAccount(
	encodedAddress string,
	userAccount state.UserAccountHandler,
	markedAccountToken *markedAlteredAccountToken,
	alteredAccounts map[string]*alteredAccount.AlteredAccount,
	options shared.AlteredAccountsOptions,
) error {
	nonce := markedAccountToken.nonce
	tokenID := markedAccountToken.identifier

	storageKey := []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier)
	storageKey = append(storageKey, []byte(tokenID)...)

	userAccountVmCommon, ok := userAccount.(vmcommon.UserAccountHandler)
	if !ok {
		return fmt.Errorf("%w for address %s", errCannotCastToVmCommonUserAccountHandler, encodedAddress)
	}

	esdtToken, _, err := aap.esdtDataStorageHandler.GetESDTNFTTokenOnDestination(userAccountVmCommon, storageKey, nonce)
	if err != nil {
		return err
	}
	if esdtToken == nil {
		log.Warn("alteredAccountsProvider: nil esdt/nft token", "address", encodedAddress, "token ID", tokenID, "nonce", nonce)
		return nil
	}

	accountTokenData := &alteredAccount.AccountTokenData{
		Identifier: tokenID,
		Balance:    esdtToken.Value.String(),
		Nonce:      nonce,
		Properties: hex.EncodeToString(esdtToken.Properties),
		MetaData:   aap.convertMetaData(esdtToken.TokenMetaData),
	}
	if options.WithAdditionalOutportData {
		accountTokenData.AdditionalData = &alteredAccount.AdditionalAccountTokenData{
			IsNFTCreate: markedAccountToken.isNFTCreate,
		}
	}

	alteredAcc := alteredAccounts[encodedAddress]
	alteredAcc.Tokens = append(alteredAccounts[encodedAddress].Tokens, accountTokenData)

	return nil
}

func (aap *alteredAccountsProvider) convertMetaData(metaData *esdt.MetaData) *alteredAccount.TokenMetaData {
	if metaData == nil {
		return nil
	}

	metaDataCreatorAddr := aap.addressConverter.SilentEncode(metaData.Creator, log)
	return &alteredAccount.TokenMetaData{
		Nonce:      metaData.Nonce,
		Name:       string(metaData.Name),
		Creator:    metaDataCreatorAddr,
		Royalties:  metaData.Royalties,
		Hash:       metaData.Hash,
		URIs:       metaData.URIs,
		Attributes: metaData.Attributes,
	}
}

func (aap *alteredAccountsProvider) extractAddressesWithBalanceChange(
	txPool *outportcore.TransactionPool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	selfShardID := aap.shardCoordinator.SelfId()

	txs := txsMapToTxHandlerSlice(txPool.Transactions)
	scrs := scrsMapToTxHandlerSlice(txPool.SmartContractResults)
	rewards := rewardsMapToTxHandlerSlice(txPool.Rewards)
	invalidTxs := txsMapToTxHandlerSlice(txPool.InvalidTxs)

	aap.extractAddressesFromTxsHandlers(selfShardID, txs, markedAlteredAccounts, process.MoveBalance)
	aap.extractAddressesFromTxsHandlers(selfShardID, scrs, markedAlteredAccounts, process.SCInvoking)
	aap.extractAddressesFromTxsHandlers(selfShardID, rewards, markedAlteredAccounts, process.RewardTx)
	aap.extractAddressesFromTxsHandlers(selfShardID, invalidTxs, markedAlteredAccounts, process.InvalidTransaction)
}

func txsMapToTxHandlerSlice(txs map[string]*outportcore.TxInfo) []data.TransactionHandler {
	ret := make([]data.TransactionHandler, len(txs))

	idx := 0
	for _, tx := range txs {
		ret[idx] = tx.Transaction
		idx++
	}

	return ret
}

func scrsMapToTxHandlerSlice(scrs map[string]*outportcore.SCRInfo) []data.TransactionHandler {
	ret := make([]data.TransactionHandler, len(scrs))

	idx := 0
	for _, scr := range scrs {
		ret[idx] = scr.SmartContractResult
		idx++
	}

	return ret
}
func rewardsMapToTxHandlerSlice(rewards map[string]*outportcore.RewardInfo) []data.TransactionHandler {
	ret := make([]data.TransactionHandler, len(rewards))

	idx := 0
	for _, reward := range rewards {
		ret[idx] = reward.Reward
		idx++
	}

	return ret
}

func (aap *alteredAccountsProvider) extractAddressesFromTxsHandlers(
	selfShardID uint32,
	txsHandlers []data.TransactionHandler,
	markedAlteredAccounts map[string]*markedAlteredAccount,
	txType process.TransactionType,
) {
	for _, txHandler := range txsHandlers {
		senderAddress := txHandler.GetSndAddr()
		receiverAddress := txHandler.GetRcvAddr()

		senderShardID := aap.shardCoordinator.ComputeId(senderAddress)
		receiverShardID := aap.shardCoordinator.ComputeId(receiverAddress)

		if senderShardID == selfShardID && len(senderAddress) > 0 {
			aap.addAddressWithBalanceChangeInMap(senderAddress, markedAlteredAccounts, true)
		}

		txValue := txHandler.GetValue()
		if txValue == nil {
			txValue = big.NewInt(0)
		}

		balanceChanged := txValue.Cmp(big.NewInt(0)) > 0
		isValid := txType != process.InvalidTransaction
		isOnCurrentShard := receiverShardID == selfShardID
		addReceiver := isValid && isOnCurrentShard && balanceChanged && len(receiverAddress) > 0
		if addReceiver {
			aap.addAddressWithBalanceChangeInMap(receiverAddress, markedAlteredAccounts, false)
		}
	}
}

func (aap *alteredAccountsProvider) addAddressWithBalanceChangeInMap(
	address []byte,
	markedAlteredAccounts map[string]*markedAlteredAccount,
	isSender bool,
) {
	isValidAddress := len(address) == aap.addressConverter.Len()
	if !isValidAddress {
		return
	}

	addressStr := string(address)
	_, addressAlreadySelected := markedAlteredAccounts[addressStr]
	if addressAlreadySelected {
		markedAlteredAccounts[addressStr].isSender = markedAlteredAccounts[addressStr].isSender || isSender
		markedAlteredAccounts[addressStr].balanceChanged = true
		return
	}

	markedAlteredAccounts[addressStr] = &markedAlteredAccount{
		isSender:       isSender,
		balanceChanged: true,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (aap *alteredAccountsProvider) IsInterfaceNil() bool {
	return aap == nil
}

func checkArgAlteredAccountsProvider(args ArgsAlteredAccountsProvider) error {
	if check.IfNil(args.ShardCoordinator) {
		return errNilShardCoordinator
	}
	if check.IfNil(args.AddressConverter) {
		return ErrNilPubKeyConverter
	}
	if check.IfNil(args.AccountsDB) {
		return ErrNilAccountsDB
	}
	if check.IfNil(args.EsdtDataStorageHandler) {
		return ErrNilESDTDataStorageHandler
	}

	return nil
}
