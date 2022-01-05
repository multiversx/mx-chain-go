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
	alteredAccountsInSelfShard := make(map[string]*indexer.AlteredAccount, 0)

	// TODO: analyze if it's better to mark all the altered accounts and read the balance/esdt data only at the end
	// this way, it wouldn't matter if the transactions or logs are processed in order

	aap.extractMoveBalanceAddresses(txPool, alteredAccountsInSelfShard)
	aap.extractESDTAccounts(txPool, alteredAccountsInSelfShard)

	return alteredAccountsInSelfShard, nil
}

func (aap *alteredAccountsProvider) extractMoveBalanceAddresses(txPool *indexer.Pool, alteredAccountsInSelfShard map[string]*indexer.AlteredAccount) {
	selfShardID := aap.shardCoordinator.SelfId()
	for _, regularTx := range txPool.Txs {
		senderShardID := aap.shardCoordinator.ComputeId(regularTx.GetSndAddr())
		receiverShardID := aap.shardCoordinator.ComputeId(regularTx.GetRcvAddr())
		if senderShardID != selfShardID && receiverShardID != selfShardID {
			continue
		}

		if senderShardID == selfShardID {
			aap.addMoveBalanceAddressInMap(regularTx.GetSndAddr(), alteredAccountsInSelfShard)
		}
		if receiverShardID == selfShardID {
			aap.addMoveBalanceAddressInMap(regularTx.GetRcvAddr(), alteredAccountsInSelfShard)
		}
	}
}

func (aap *alteredAccountsProvider) addMoveBalanceAddressInMap(address []byte, alteredAccountsInSelfShard map[string]*indexer.AlteredAccount) {
	encodedAddress := aap.addressConverter.Encode(address)
	_, addressAlreadySelected := alteredAccountsInSelfShard[encodedAddress]
	if addressAlreadySelected {
		return
	}

	account, err := aap.accountsDB.LoadAccount(address)
	if err != nil {
		log.Warn("cannot load account when computing altered accounts",
			"address", encodedAddress,
			"error", err)
		return
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		log.Warn("cannot cast AccountHandler to UserAccountHandler", "address", encodedAddress)
		return
	}

	alteredAccountsInSelfShard[encodedAddress] = &indexer.AlteredAccount{
		Address: encodedAddress,
		Balance: userAccount.GetBalance().String(),
		Nonce:   userAccount.GetNonce(),
	}
}

func (aap *alteredAccountsProvider) extractESDTAccounts(txPool *indexer.Pool, alteredAccountsInSelfShard map[string]*indexer.AlteredAccount) {
	for _, txLog := range txPool.Logs {
		for _, event := range txLog.LogHandler.GetLogEvents() {
			aap.processEvent(event, alteredAccountsInSelfShard)
		}
	}
}

func (aap *alteredAccountsProvider) processEvent(event data.EventHandler, alteredAccountsInSelfShard map[string]*indexer.AlteredAccount) {
	_, isEsdtOperation := aap.fungibleTokensIdentifier[string(event.GetIdentifier())]
	if isEsdtOperation {
		err := aap.extractEsdtData(event, zeroBigInt, alteredAccountsInSelfShard)
		if err != nil {
			log.Debug("cannot extract esdt data", "error", err)
			return
		}
	}

	_, isNftOperation := aap.nonFungibleTokensIdentifier[string(event.GetIdentifier())]
	if isNftOperation {
		topics := event.GetTopics()
		if len(topics) == 0 {
			return
		}

		nonce := topics[1]
		nonceBigInt := big.NewInt(0).SetBytes(nonce)
		err := aap.extractEsdtData(event, nonceBigInt, alteredAccountsInSelfShard)
		if err != nil {
			log.Debug("cannot extract nft data", "error", err)
			return
		}
	}
}

func (aap *alteredAccountsProvider) extractEsdtData(event data.EventHandler, nonce *big.Int, alteredAccountsInSelfShard map[string]*indexer.AlteredAccount) error {
	address := event.GetAddress()
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil
	}

	tokenID := topics[0]

	storageKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier)
	storageKey = append(storageKey, tokenID...)
	if nonce.Uint64() > 0 {
		storageKey = append(storageKey, nonce.Bytes()...)
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

	encodedAddress := aap.addressConverter.Encode(address)
	aap.addTokenDataToAccount(encodedAddress, userAccount, esdtToken, string(tokenID), nonce.Uint64(), alteredAccountsInSelfShard)

	return nil
}

func (aap *alteredAccountsProvider) addTokenDataToAccount(
	address string,
	accountHandler state.UserAccountHandler,
	esdtToken esdt.ESDigitalToken,
	tokenID string,
	nonce uint64,
	alteredAccountsInSelfShard map[string]*indexer.AlteredAccount) {
	alteredAccount, exists := alteredAccountsInSelfShard[address]
	if !exists {
		alteredAccount = &indexer.AlteredAccount{
			Address: address,
			Balance: accountHandler.GetBalance().String(),
			Nonce:   accountHandler.GetNonce(),
		}
	}

	tokenData := indexer.AccountTokenData{}
	tokenData.Balance = esdtToken.Value.String()
	tokenData.Identifier = tokenID
	tokenData.Nonce = nonce
	tokenData.MetaData = esdtToken.TokenMetaData

	if len(alteredAccount.Tokens) == 0 {
		alteredAccount.Tokens = make([]*indexer.AccountTokenData, 0)
	}
	alteredAccount.Tokens = append(alteredAccount.Tokens, &tokenData)

	alteredAccountsInSelfShard[address] = alteredAccount
}

// IsInterfaceNil returns true if there is no value under the interface
func (aap *alteredAccountsProvider) IsInterfaceNil() bool {
	return aap == nil
}
