package alteredaccounts

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

var (
	log        = logger.GetOrCreate("process/block/alteredaccounts")
	zeroBigInt = big.NewInt(0)
)

const (
	idxTokenIDInTopics    = 0
	idxTokenNonceInTopics = 1
)

type markedAlteredAccountToken struct {
	identifier string
	nonce      uint64
}

type markedAlteredAccount struct {
	tokens map[string]*markedAlteredAccountToken
}

// ArgsAlteredAccountsProvider holds the arguments needed for creating a new instance of alteredAccountsProvider
type ArgsAlteredAccountsProvider struct {
	ShardCoordinator sharding.Coordinator
	AddressConverter core.PubkeyConverter
	AccountsDB       state.AccountsAdapter
	Marshalizer      marshal.Marshalizer
}

type alteredAccountsProvider struct {
	shardCoordinator            sharding.Coordinator
	addressConverter            core.PubkeyConverter
	accountsDB                  state.AccountsAdapter
	marshalizer                 marshal.Marshalizer
	fungibleTokensIdentifier    map[string]struct{}
	nonFungibleTokensIdentifier map[string]struct{}
	mutExtractAccounts          sync.Mutex
}

// NewAlteredAccountsProvider returns a new instance of alteredAccountsProvider
func NewAlteredAccountsProvider(args ArgsAlteredAccountsProvider) (*alteredAccountsProvider, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, errNilShardCoordinator
	}
	if check.IfNil(args.AddressConverter) {
		return nil, errNilPubKeyConverter
	}
	if check.IfNil(args.AccountsDB) {
		return nil, errNilAccountsDB
	}
	if check.IfNil(args.Marshalizer) {
		return nil, errNilMarshalizer
	}

	return &alteredAccountsProvider{
		shardCoordinator: args.ShardCoordinator,
		addressConverter: args.AddressConverter,
		accountsDB:       args.AccountsDB,
		marshalizer:      args.Marshalizer,
		fungibleTokensIdentifier: map[string]struct{}{
			core.BuiltInFunctionESDTTransfer:         {},
			core.BuiltInFunctionESDTBurn:             {},
			core.BuiltInFunctionESDTLocalMint:        {},
			core.BuiltInFunctionESDTLocalBurn:        {},
			core.BuiltInFunctionESDTWipe:             {},
			core.BuiltInFunctionMultiESDTNFTTransfer: {},
		},
		nonFungibleTokensIdentifier: map[string]struct{}{
			core.BuiltInFunctionESDTNFTTransfer:      {},
			core.BuiltInFunctionESDTNFTBurn:          {},
			core.BuiltInFunctionESDTNFTAddQuantity:   {},
			core.BuiltInFunctionESDTNFTCreate:        {},
			core.BuiltInFunctionMultiESDTNFTTransfer: {},
		},
	}, nil
}

// ExtractAlteredAccountsFromPool will extract and return altered accounts from the pool
func (aap *alteredAccountsProvider) ExtractAlteredAccountsFromPool(txPool *indexer.Pool) (map[string]*indexer.AlteredAccount, error) {
	aap.mutExtractAccounts.Lock()
	defer aap.mutExtractAccounts.Unlock()

	markedAccounts := make(map[string]*markedAlteredAccount)
	aap.extractAddressesWithBalanceChange(txPool, markedAccounts)
	err := aap.extractESDTAccounts(txPool, markedAccounts)
	if err != nil {
		return nil, err
	}

	return aap.fetchDataForMarkedAccounts(markedAccounts)
}

func (aap *alteredAccountsProvider) fetchDataForMarkedAccounts(markedAccounts map[string]*markedAlteredAccount) (map[string]*indexer.AlteredAccount, error) {
	alteredAccounts := make(map[string]*indexer.AlteredAccount)
	var err error
	for address, markedAccount := range markedAccounts {
		err = aap.processMarkedAccountData(address, markedAccount.tokens, alteredAccounts)
		if err != nil {
			return nil, err
		}
	}

	return alteredAccounts, nil
}

func (aap *alteredAccountsProvider) processMarkedAccountData(
	addressStr string,
	markedAccountTokens map[string]*markedAlteredAccountToken,
	alteredAccounts map[string]*indexer.AlteredAccount,
) error {
	addressBytes := []byte(addressStr)
	encodedAddress := aap.addressConverter.Encode(addressBytes)

	account, err := aap.accountsDB.LoadAccount(addressBytes)
	if err != nil {
		log.Warn("cannot load account when computing altered accounts",
			"address", addressBytes,
			"error", err)
		return err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		log.Warn("cannot cast AccountHandler to UserAccountHandler", "address", addressBytes)
		return err
	}

	alteredAccounts[encodedAddress] = &indexer.AlteredAccount{
		Address: encodedAddress,
		Balance: userAccount.GetBalance().String(),
		Nonce:   userAccount.GetNonce(),
	}

	for tokenKey, tokenData := range markedAccountTokens {
		err = aap.fetchTokensDataForMarkedAccount([]byte(tokenKey), encodedAddress, userAccount, tokenData, alteredAccounts)
		if err != nil {
			log.Warn("cannot fetch token data for marked account", "error", err)
			return err
		}
	}

	return nil
}

func (aap *alteredAccountsProvider) fetchTokensDataForMarkedAccount(
	tokenKey []byte,
	encodedAddress string,
	userAccount state.UserAccountHandler,
	markedAccountToken *markedAlteredAccountToken,
	alteredAccounts map[string]*indexer.AlteredAccount,
) error {
	nonce := markedAccountToken.nonce
	tokenID := markedAccountToken.identifier

	storageKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier)
	storageKey = append(storageKey, tokenKey...)

	esdtTokenBytes, err := userAccount.RetrieveValueFromDataTrieTracker(tokenKey)
	if err != nil {
		return err
	}

	var esdtToken esdt.ESDigitalToken
	err = aap.marshalizer.Unmarshal(&esdtToken, esdtTokenBytes)
	if err != nil {
		return err
	}

	alteredAccount := alteredAccounts[encodedAddress]

	alteredAccount.Tokens = append(alteredAccount.Tokens, &indexer.AccountTokenData{
		Identifier: tokenID,
		Balance:    esdtToken.Value.String(),
		Nonce:      nonce,
		MetaData:   esdtToken.TokenMetaData,
	})

	alteredAccounts[encodedAddress] = alteredAccount

	return nil
}

func (aap *alteredAccountsProvider) extractAddressesWithBalanceChange(
	txPool *indexer.Pool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	selfShardID := aap.shardCoordinator.SelfId()

	aap.extractAddressesFromTxsHandlers(selfShardID, txPool.Txs, markedAlteredAccounts)
	aap.extractAddressesFromTxsHandlers(selfShardID, txPool.Scrs, markedAlteredAccounts)
	aap.extractAddressesFromTxsHandlers(selfShardID, txPool.Rewards, markedAlteredAccounts)
	aap.extractAddressesFromTxsHandlers(selfShardID, txPool.Invalid, markedAlteredAccounts)
}

func (aap *alteredAccountsProvider) extractAddressesFromTxsHandlers(
	selfShardID uint32,
	txsHandlers map[string]data.TransactionHandler,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	for _, txHandler := range txsHandlers {
		senderShardID := aap.shardCoordinator.ComputeId(txHandler.GetSndAddr())
		receiverShardID := aap.shardCoordinator.ComputeId(txHandler.GetRcvAddr())

		if senderShardID == selfShardID {
			aap.addAddressWithBalanceChangeInMap(txHandler.GetSndAddr(), markedAlteredAccounts)
		}
		if receiverShardID == selfShardID {
			aap.addAddressWithBalanceChangeInMap(txHandler.GetRcvAddr(), markedAlteredAccounts)
		}
	}
}

func (aap *alteredAccountsProvider) addAddressWithBalanceChangeInMap(
	address []byte,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	_, addressAlreadySelected := markedAlteredAccounts[string(address)]
	if addressAlreadySelected {
		return
	}

	markedAlteredAccounts[string(address)] = &markedAlteredAccount{}
}

func (aap *alteredAccountsProvider) extractESDTAccounts(
	txPool *indexer.Pool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	var err error
	for _, txLog := range txPool.Logs {
		for _, event := range txLog.LogHandler.GetLogEvents() {
			err = aap.processEvent(event, markedAlteredAccounts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (aap *alteredAccountsProvider) processEvent(
	event data.EventHandler,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	_, isEsdtOperation := aap.fungibleTokensIdentifier[string(event.GetIdentifier())]
	if isEsdtOperation {
		err := aap.extractEsdtData(event, zeroBigInt, markedAlteredAccounts)
		if err != nil {
			log.Debug("cannot extract esdt data", "error", err)
			return err
		}

		return nil
	}

	_, isNftOperation := aap.nonFungibleTokensIdentifier[string(event.GetIdentifier())]
	if isNftOperation {
		topics := event.GetTopics()
		if len(topics) == 0 {
			return nil
		}

		nonce := topics[idxTokenNonceInTopics]
		nonceBigInt := big.NewInt(0).SetBytes(nonce)
		err := aap.extractEsdtData(event, nonceBigInt, markedAlteredAccounts)
		if err != nil {
			log.Debug("cannot extract nft data", "error", err)
			return nil
		}

		return nil
	}

	return nil
}

func (aap *alteredAccountsProvider) extractEsdtData(
	event data.EventHandler,
	nonce *big.Int,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	address := event.GetAddress()
	addressStr := string(address)
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil
	}

	// TODO: treat destination as well (for NFT and Multi transfers - topics[3] is the receiver address)
	tokenID := topics[idxTokenIDInTopics]

	_, exists := markedAlteredAccounts[addressStr]
	if !exists {
		markedAlteredAccounts[addressStr] = &markedAlteredAccount{}
	}

	markedAccount := markedAlteredAccounts[addressStr]
	if markedAccount.tokens == nil {
		markedAccount.tokens = make(map[string]*markedAlteredAccountToken)
	}

	tokenKey := string(tokenID) + string(nonce.Bytes())
	_, alreadyExists := markedAccount.tokens[tokenKey]
	if alreadyExists {
		return nil
	}

	markedAccount.tokens[tokenKey] = &markedAlteredAccountToken{
		identifier: string(tokenID),
		nonce:      nonce.Uint64(),
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (aap *alteredAccountsProvider) IsInterfaceNil() bool {
	return aap == nil
}
