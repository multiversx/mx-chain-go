package alteredaccounts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/sharding"
)

const (
	idxTokenIDInTopics         = 0
	idxTokenNonceInTopics      = 1
	idxReceiverAddressInTopics = 3
)

type tokensProcessor struct {
	shardCoordinator sharding.Coordinator
	tokensIdentifier map[string]struct{}
}

func newTokensProcessor(shardCoordinator sharding.Coordinator) *tokensProcessor {
	return &tokensProcessor{
		tokensIdentifier: map[string]struct{}{
			core.BuiltInFunctionESDTTransfer:         {},
			core.BuiltInFunctionESDTBurn:             {},
			core.BuiltInFunctionESDTLocalMint:        {},
			core.BuiltInFunctionESDTLocalBurn:        {},
			core.BuiltInFunctionESDTWipe:             {},
			core.BuiltInFunctionMultiESDTNFTTransfer: {},
			core.BuiltInFunctionESDTNFTTransfer:      {},
			core.BuiltInFunctionESDTNFTBurn:          {},
			core.BuiltInFunctionESDTNFTAddQuantity:   {},
			core.BuiltInFunctionESDTNFTCreate:        {},
			core.BuiltInFunctionESDTFreeze:           {},
			core.BuiltInFunctionESDTUnFreeze:         {},
		},
		shardCoordinator: shardCoordinator,
	}
}

func (tp *tokensProcessor) extractESDTAccounts(
	txPool *outportcore.TransactionPool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	var err error
	for _, txLog := range txPool.Logs {
		for _, event := range txLog.Log.Events {
			err = tp.processEvent(event, markedAlteredAccounts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (tp *tokensProcessor) processEvent(
	event data.EventHandler,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	_, isESDT := tp.tokensIdentifier[string(event.GetIdentifier())]
	if !isESDT {
		return nil
	}

	topics := event.GetTopics()
	if len(topics) < idxTokenNonceInTopics+1 {
		return nil
	}

	nonce := topics[idxTokenNonceInTopics]
	nonceBigInt := big.NewInt(0).SetBytes(nonce)
	err := tp.extractEsdtData(event, nonceBigInt, markedAlteredAccounts)
	if err != nil {
		log.Debug("cannot extract esdt data", "error", err)
		return nil
	}

	return nil
}

func (tp *tokensProcessor) extractEsdtData(
	event data.EventHandler,
	nonce *big.Int,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	address := event.GetAddress()
	topics := event.GetTopics()
	if len(topics) == 0 {
		return nil
	}

	identifier := string(event.GetIdentifier())
	isNFTCreate := identifier == core.BuiltInFunctionESDTNFTCreate
	tokenID := topics[idxTokenIDInTopics]
	err := tp.processEsdtDataForAddress(address, nonce, string(tokenID), markedAlteredAccounts, isNFTCreate)
	if err != nil {
		return err
	}

	// in case of esdt transfer, nft transfer, wipe or multi esdt transfers, the 3rd index of the topics contains the destination address
	eventShouldContainReceiverAddress := identifier == core.BuiltInFunctionESDTTransfer ||
		identifier == core.BuiltInFunctionESDTNFTTransfer ||
		identifier == core.BuiltInFunctionESDTWipe ||
		identifier == core.BuiltInFunctionMultiESDTNFTTransfer ||
		identifier == core.BuiltInFunctionESDTFreeze ||
		identifier == core.BuiltInFunctionESDTUnFreeze

	if eventShouldContainReceiverAddress && len(topics) > idxReceiverAddressInTopics {
		destinationAddress := topics[idxReceiverAddressInTopics]
		err = tp.processEsdtDataForAddress(destinationAddress, nonce, string(tokenID), markedAlteredAccounts, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tp *tokensProcessor) processEsdtDataForAddress(
	address []byte,
	nonce *big.Int,
	tokenID string,
	markedAlteredAccounts map[string]*markedAlteredAccount,
	isNFTCreate bool,
) error {
	if !tp.isSameShard(address) {
		return nil
	}

	addressStr := string(address)
	markedAccount, exists := markedAlteredAccounts[addressStr]
	if !exists {
		markedAccount = &markedAlteredAccount{}
		markedAlteredAccounts[addressStr] = markedAccount
	}

	if markedAccount.tokens == nil {
		markedAccount.tokens = make(map[string]*markedAlteredAccountToken)
	}

	tokenKey := tokenID + string(nonce.Bytes())
	_, alreadyExists := markedAccount.tokens[tokenKey]
	if alreadyExists {
		markedAccount.tokens[tokenKey].isNFTCreate = markedAccount.tokens[tokenKey].isNFTCreate || isNFTCreate
		return nil
	}

	markedAccount.tokens[tokenKey] = &markedAlteredAccountToken{
		identifier:  tokenID,
		nonce:       nonce.Uint64(),
		isNFTCreate: isNFTCreate,
	}

	return nil
}

func (tp *tokensProcessor) isSameShard(address []byte) bool {
	return tp.shardCoordinator.SelfId() == tp.shardCoordinator.ComputeId(address)
}
