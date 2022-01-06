package alteredaccounts

import (
	"math/big"

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

type markedAlteredAccountToken struct {
	identifier string
	nonce      uint64
}

type markedAlteredAccount struct {
	tokens  []*markedAlteredAccountToken
}

// ArgsAlteredAccountsProvider
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
}

// NewAlteredAccountsProvider returns a new instance of alteredAccountsProvider
func NewAlteredAccountsProvider(args ArgsAlteredAccountsProvider) (*alteredAccountsProvider, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(args.AddressConverter) {
		return nil, ErrNilPubKeyConverter
	}
	if check.IfNil(args.AccountsDB) {
		return nil, ErrNilAccountsDB
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
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
	markedAccounts := make(map[string]*markedAlteredAccount)
	aap.extractMoveBalanceAddresses(txPool, markedAccounts)
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
	encodedAddress string,
	markedAccountTokens []*markedAlteredAccountToken,
	alteredAccounts map[string]*indexer.AlteredAccount,
) error {
	addressBytes, err := aap.addressConverter.Decode(encodedAddress)
	if err != nil {
		log.Warn("cannot decode marked altered account address", "error", err)
		return err
	}

	account, err := aap.accountsDB.LoadAccount(addressBytes)
	if err != nil {
		log.Warn("cannot load account when computing altered accounts",
			"address", encodedAddress,
			"error", err)
		return err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		log.Warn("cannot cast AccountHandler to UserAccountHandler", "address", encodedAddress)
		return err
	}

	alteredAccounts[encodedAddress] = &indexer.AlteredAccount{
		Address: encodedAddress,
		Balance: userAccount.GetBalance().String(),
		Nonce:   userAccount.GetNonce(),
	}

	if len(markedAccountTokens) > 0 {
		for _, tokenData := range markedAccountTokens {
			err = aap.fetchTokensDataForMarkedAccount(encodedAddress, tokenData, alteredAccounts)
			if err != nil {
				log.Warn("cannot fetch token data for marked account", "error", err)
				return err
			}
		}
	}

	return nil
}

func (aap *alteredAccountsProvider) fetchTokensDataForMarkedAccount(
	encodedAddress string,
	markedAccountToken *markedAlteredAccountToken,
	alteredAccounts map[string]*indexer.AlteredAccount,
) error {
	nonce := markedAccountToken.nonce
	nonceBigInt := big.NewInt(0).SetUint64(nonce)

	tokenID := markedAccountToken.identifier

	storageKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier)
	storageKey = append(storageKey, []byte(tokenID)...)
	if nonce > 0 {
		storageKey = append(storageKey, nonceBigInt.Bytes()...)
	}

	address, err := aap.addressConverter.Decode(encodedAddress)
	if err != nil {
		log.Warn("cannot decode encoded address", "error", err)
		return err
	}

	account, err := aap.accountsDB.LoadAccount(address)
	if err != nil {
		return err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil
	}

	esdtTokenBytes, err := userAccount.RetrieveValueFromDataTrieTracker(storageKey)
	if err != nil {
		return err
	}

	var esdtToken esdt.ESDigitalToken
	err = aap.marshalizer.Unmarshal(&esdtToken, esdtTokenBytes)
	if err != nil {
		return err
	}

	alteredAccount := alteredAccounts[encodedAddress]

	if alteredAccount.Tokens == nil {
		alteredAccount.Tokens = make([]*indexer.AccountTokenData, 0)
	}

	alteredAccount.Tokens = append(alteredAccount.Tokens, &indexer.AccountTokenData{
		Identifier: tokenID,
		Balance:    esdtToken.Value.String(),
		Nonce:      nonce,
		MetaData:   esdtToken.TokenMetaData,
	})

	alteredAccounts[encodedAddress] = alteredAccount

	return nil
}

func (aap *alteredAccountsProvider) extractMoveBalanceAddresses(
	txPool *indexer.Pool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	selfShardID := aap.shardCoordinator.SelfId()
	for _, regularTx := range txPool.Txs {
		senderShardID := aap.shardCoordinator.ComputeId(regularTx.GetSndAddr())
		receiverShardID := aap.shardCoordinator.ComputeId(regularTx.GetRcvAddr())
		if senderShardID != selfShardID && receiverShardID != selfShardID {
			continue
		}

		if senderShardID == selfShardID {
			aap.addMoveBalanceAddressInMap(regularTx.GetSndAddr(), markedAlteredAccounts)
		}
		if receiverShardID == selfShardID {
			aap.addMoveBalanceAddressInMap(regularTx.GetRcvAddr(), markedAlteredAccounts)
		}
	}
}

func (aap *alteredAccountsProvider) addMoveBalanceAddressInMap(
	address []byte,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	encodedAddress := aap.addressConverter.Encode(address)
	_, addressAlreadySelected := markedAlteredAccounts[encodedAddress]
	if addressAlreadySelected {
		return
	}

	markedAlteredAccounts[encodedAddress] = &markedAlteredAccount{}
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
	}

	_, isNftOperation := aap.nonFungibleTokensIdentifier[string(event.GetIdentifier())]
	if isNftOperation {
		topics := event.GetTopics()
		if len(topics) == 0 {
			return nil
		}

		nonce := topics[1]
		nonceBigInt := big.NewInt(0).SetBytes(nonce)
		err := aap.extractEsdtData(event, nonceBigInt, markedAlteredAccounts)
		if err != nil {
			log.Debug("cannot extract nft data", "error", err)
			return nil
		}
	}

	return nil
}

func (aap *alteredAccountsProvider) extractEsdtData(
	event data.EventHandler,
	nonce *big.Int,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	address := event.GetAddress()
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil
	}

	tokenID := topics[0]

	encodedAddress := aap.addressConverter.Encode(address)
	_, exists := markedAlteredAccounts[encodedAddress]
	if !exists {
		markedAlteredAccounts[encodedAddress] = &markedAlteredAccount{}
	}

	markedAccount := markedAlteredAccounts[encodedAddress]
	if len(markedAccount.tokens) == 0 {
		markedAccount.tokens = make([]*markedAlteredAccountToken, 0)
	}

	if aap.isTokenAlreadyMarked(markedAccount, string(tokenID), nonce.Uint64()) {
		return nil
	}

	markedAccount.tokens = append(markedAccount.tokens, &markedAlteredAccountToken{
		identifier: string(tokenID),
		nonce:      nonce.Uint64(),
	})

	return nil
}

func (aap *alteredAccountsProvider) isTokenAlreadyMarked(markedAccount *markedAlteredAccount, identifier string, nonce uint64) bool {
	for _, tokenData := range markedAccount.tokens {
		if tokenData.identifier == identifier && tokenData.nonce == nonce {
			return true
		}
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (aap *alteredAccountsProvider) IsInterfaceNil() bool {
	return aap == nil
}
