package alteredaccounts

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/outport/process/alteredaccounts/shared"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
func (aap *alteredAccountsProvider) ExtractAlteredAccountsFromPool(txPool *outportcore.Pool, options shared.AlteredAccountsOptions) (map[string]*outportcore.AlteredAccount, error) {
	if err := options.Verify(); err != nil {
		return nil, err
	}

	aap.mutExtractAccounts.Lock()
	defer aap.mutExtractAccounts.Unlock()

	if txPool == nil {
		log.Warn("alteredAccountsProvider: ExtractAlteredAccountsFromPool", "txPool is nil", "will return")
		return map[string]*outportcore.AlteredAccount{}, nil
	}

	markedAccounts := make(map[string]*markedAlteredAccount)
	aap.extractAddressesWithBalanceChange(txPool, markedAccounts)
	err := aap.tokensProc.extractESDTAccounts(txPool, markedAccounts)
	if err != nil {
		return nil, err
	}

	return aap.fetchDataForMarkedAccounts(markedAccounts, options)
}

func (aap *alteredAccountsProvider) fetchDataForMarkedAccounts(markedAccounts map[string]*markedAlteredAccount, options shared.AlteredAccountsOptions) (map[string]*outportcore.AlteredAccount, error) {
	alteredAccounts := make(map[string]*outportcore.AlteredAccount)
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
	alteredAccounts map[string]*outportcore.AlteredAccount,
	options shared.AlteredAccountsOptions,
) error {
	addressBytes := []byte(addressStr)
	encodedAddress := aap.addressConverter.Encode(addressBytes)

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

func (aap *alteredAccountsProvider) addAdditionalDataInAlteredAccount(alteredAccount *outportcore.AlteredAccount, userAccount state.UserAccountHandler, markedAccount *markedAlteredAccount) {
	alteredAccount.AdditionalData = &outportcore.AdditionalAccountData{
		IsSender:       markedAccount.isSender,
		BalanceChanged: markedAccount.balanceChanged,
		UserName:       string(userAccount.GetUserName()),
	}

	ownerAddressBytes := userAccount.GetOwnerAddress()
	if core.IsSmartContractAddress(userAccount.AddressBytes()) && len(ownerAddressBytes) == aap.addressConverter.Len() {
		alteredAccount.AdditionalData.CurrentOwner = aap.addressConverter.Encode(ownerAddressBytes)
	}
	developerRewards := userAccount.GetDeveloperReward()
	if developerRewards != nil {
		alteredAccount.AdditionalData.DeveloperRewards = developerRewards.String()
	}
}

func (aap *alteredAccountsProvider) getAlteredAccountFromUserAccounts(userEncodedAddress string, userAccount state.UserAccountHandler) *outportcore.AlteredAccount {
	alteredAccount := &outportcore.AlteredAccount{
		Address: userEncodedAddress,
		Balance: userAccount.GetBalance().String(),
		Nonce:   userAccount.GetNonce(),
	}

	return alteredAccount
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
	alteredAccounts map[string]*outportcore.AlteredAccount,
	options shared.AlteredAccountsOptions,
) error {
	nonce := markedAccountToken.nonce
	tokenID := markedAccountToken.identifier

	storageKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier)
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

	accountTokenData := &outportcore.AccountTokenData{
		Identifier: tokenID,
		Balance:    esdtToken.Value.String(),
		Nonce:      nonce,
		Properties: string(esdtToken.Properties),
		MetaData:   aap.convertMetaData(esdtToken.TokenMetaData),
	}
	if options.WithAdditionalOutportData {
		accountTokenData.AdditionalData = &outportcore.AdditionalAccountTokenData{
			IsNFTCreate: markedAccountToken.isNFTCreate,
		}
	}

	alteredAccount := alteredAccounts[encodedAddress]
	alteredAccount.Tokens = append(alteredAccounts[encodedAddress].Tokens, accountTokenData)

	return nil
}

func (aap *alteredAccountsProvider) convertMetaData(metaData *esdt.MetaData) *outportcore.TokenMetaData {
	if metaData == nil {
		return nil
	}

	return &outportcore.TokenMetaData{
		Nonce:      metaData.Nonce,
		Name:       string(metaData.Name),
		Creator:    aap.addressConverter.Encode(metaData.Creator),
		Royalties:  metaData.Royalties,
		Hash:       metaData.Hash,
		URIs:       metaData.URIs,
		Attributes: metaData.Attributes,
	}
}

func (aap *alteredAccountsProvider) extractAddressesWithBalanceChange(
	txPool *outportcore.Pool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	selfShardID := aap.shardCoordinator.SelfId()

	aap.extractAddressesFromTxsHandlers(selfShardID, txPool.Txs, markedAlteredAccounts, process.MoveBalance)
	aap.extractAddressesFromTxsHandlers(selfShardID, txPool.Scrs, markedAlteredAccounts, process.SCInvoking)
	aap.extractAddressesFromTxsHandlers(selfShardID, txPool.Rewards, markedAlteredAccounts, process.RewardTx)
	aap.extractAddressesFromTxsHandlers(selfShardID, txPool.Invalid, markedAlteredAccounts, process.InvalidTransaction)
}

func (aap *alteredAccountsProvider) extractAddressesFromTxsHandlers(
	selfShardID uint32,
	txsHandlers map[string]data.TransactionHandlerWithGasUsedAndFee,
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

func checkArgAlteredAccountsProvider(arg ArgsAlteredAccountsProvider) error {
	if check.IfNil(arg.ShardCoordinator) {
		return errNilShardCoordinator
	}
	if check.IfNil(arg.AddressConverter) {
		return ErrNilPubKeyConverter
	}
	if check.IfNil(arg.AccountsDB) {
		return ErrNilAccountsDB
	}
	if check.IfNil(arg.EsdtDataStorageHandler) {
		return ErrNilESDTDataStorageHandler
	}

	return nil
}
