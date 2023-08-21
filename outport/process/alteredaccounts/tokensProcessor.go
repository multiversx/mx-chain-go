package alteredaccounts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/sharding"
)

const (
	idxTokenIDInTopics         = 0
	idxTokenNonceInTopics      = 1
	idxReceiverAddressInTopics = 3
	minTopicsMultiTransferV2   = 4
)

type tokensProcessor struct {
	enableEpochsHandler common.EnableEpochsHandler
	shardCoordinator    sharding.Coordinator
	tokensIdentifier    map[string]struct{}
}

func newTokensProcessor(shardCoordinator sharding.Coordinator, enableEpochsHandler common.EnableEpochsHandler) *tokensProcessor {
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
		shardCoordinator:    shardCoordinator,
		enableEpochsHandler: enableEpochsHandler,
	}
}

func (tp *tokensProcessor) extractESDTAccounts(
	txPool *outportcore.TransactionPool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	for _, txLog := range txPool.Logs {
		for _, event := range txLog.Log.Events {
			tp.processEvent(event, markedAlteredAccounts)
		}
	}

	return nil
}

func (tp *tokensProcessor) processEvent(
	event data.EventHandler,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) {
	eventIdentifier := string(event.GetIdentifier())
	_, isESDT := tp.tokensIdentifier[eventIdentifier]
	if !isESDT {
		return
	}

	topics := event.GetTopics()
	if len(topics) < idxTokenNonceInTopics+1 {
		return
	}

	isMultiTransferEventV2 := eventIdentifier == core.BuiltInFunctionMultiESDTNFTTransfer && tp.enableEpochsHandler.IsScToScEventLogEnabled()
	if isMultiTransferEventV2 {
		tp.processMultiTransferEventV2(event, markedAlteredAccounts)
		return
	}

	nonce := topics[idxTokenNonceInTopics]
	nonceBigInt := big.NewInt(0).SetBytes(nonce)
	err := tp.extractEsdtData(event, nonceBigInt, markedAlteredAccounts)
	if err != nil {
		log.Debug("cannot extract esdt data", "error", err)
		return
	}
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
	tp.processEsdtDataForAddress(address, nonce, string(tokenID), markedAlteredAccounts, isNFTCreate)

	// in case of esdt transfer, nft transfer, wipe or multi esdt transfers, the 3rd index of the topics contains the destination address
	eventShouldContainReceiverAddress := identifier == core.BuiltInFunctionESDTTransfer ||
		identifier == core.BuiltInFunctionESDTNFTTransfer ||
		identifier == core.BuiltInFunctionESDTWipe ||
		identifier == core.BuiltInFunctionMultiESDTNFTTransfer ||
		identifier == core.BuiltInFunctionESDTFreeze ||
		identifier == core.BuiltInFunctionESDTUnFreeze

	if eventShouldContainReceiverAddress && len(topics) > idxReceiverAddressInTopics {
		destinationAddress := topics[idxReceiverAddressInTopics]
		tp.processEsdtDataForAddress(destinationAddress, nonce, string(tokenID), markedAlteredAccounts, false)
	}

	return nil
}

func (tp *tokensProcessor) processMultiTransferEventV2(event data.EventHandler, markedAlteredAccounts map[string]*markedAlteredAccount) {
	topics := event.GetTopics()
	// MultiESDTNFTTransfer V2 event
	// N = len(topics)
	// i := 0; i < N-1; i+=3
	// {
	// 		topics[i] --- token identifier
	// 		topics[i+1] --- token nonce
	// 		topics[i+2] --- transferred value
	// }
	// topics[N-1]   --- destination address
	if len(topics) < minTopicsMultiTransferV2 {
		return
	}

	address := event.GetAddress()
	numOfTopics := len(topics)
	destinationAddress := topics[numOfTopics-1]
	for i := 0; i < numOfTopics-1; i += 3 {
		tokenID := topics[i]
		nonceBigInt := big.NewInt(0).SetBytes(topics[i+1])
		// process event for the sender address
		tp.processEsdtDataForAddress(address, nonceBigInt, string(tokenID), markedAlteredAccounts, false)

		// process event for the destination address
		tp.processEsdtDataForAddress(destinationAddress, nonceBigInt, string(tokenID), markedAlteredAccounts, false)
	}
}

func (tp *tokensProcessor) processEsdtDataForAddress(
	address []byte,
	nonce *big.Int,
	tokenID string,
	markedAlteredAccounts map[string]*markedAlteredAccount,
	isNFTCreate bool,
) {
	if !tp.isSameShard(address) {
		return
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
		return
	}

	markedAccount.tokens[tokenKey] = &markedAlteredAccountToken{
		identifier:  tokenID,
		nonce:       nonce.Uint64(),
		isNFTCreate: isNFTCreate,
	}
}

func (tp *tokensProcessor) isSameShard(address []byte) bool {
	return tp.shardCoordinator.SelfId() == tp.shardCoordinator.ComputeId(address)
}
