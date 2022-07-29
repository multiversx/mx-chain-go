package alteredaccounts

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const (
	idxTokenIDInTopics         = 0
	idxTokenNonceInTopics      = 1
	idxReceiverAddressInTopics = 3
	issueFungibleESDTFunc      = "issue"
	issueSemiFungibleESDTFunc  = "issueSemiFungible"
	issueNonFungibleESDTFunc   = "issueNonFungible"
	registerMetaESDTFunc       = "registerMetaESDT"
	changeSFTToMetaESDTFunc    = "changeSFTToMetaESDT"
	transferOwnershipFunc      = "transferOwnership"
	registerAndSetRolesFunc    = "registerAndSetAllRoles"
)

type tokensProcessor struct {
	shardCoordinator            sharding.Coordinator
	fungibleTokensIdentifiers   map[string]struct{}
	nonFungibleTokensIdentifier map[string]struct{}
}

func newTokensProcessor(shardCoordinator sharding.Coordinator) *tokensProcessor {
	return &tokensProcessor{
		fungibleTokensIdentifiers: map[string]struct{}{
			core.BuiltInFunctionESDTTransfer:         {},
			core.BuiltInFunctionESDTBurn:             {},
			core.BuiltInFunctionESDTLocalMint:        {},
			core.BuiltInFunctionESDTLocalBurn:        {},
			core.BuiltInFunctionESDTWipe:             {},
			core.BuiltInFunctionMultiESDTNFTTransfer: {},
			transferOwnershipFunc:                    {},
			issueFungibleESDTFunc:                    {},
			registerAndSetRolesFunc:                  {},
		},
		nonFungibleTokensIdentifier: map[string]struct{}{
			core.BuiltInFunctionESDTNFTTransfer:      {},
			core.BuiltInFunctionESDTNFTBurn:          {},
			core.BuiltInFunctionESDTNFTAddQuantity:   {},
			core.BuiltInFunctionESDTNFTCreate:        {},
			core.BuiltInFunctionMultiESDTNFTTransfer: {},
			issueSemiFungibleESDTFunc:                {},
			issueNonFungibleESDTFunc:                 {},
			registerMetaESDTFunc:                     {},
			changeSFTToMetaESDTFunc:                  {},
			transferOwnershipFunc:                    {},
			registerAndSetRolesFunc:                  {},
		},
		shardCoordinator: shardCoordinator,
	}
}

func (tp *tokensProcessor) extractESDTAccounts(
	txPool *outportcore.Pool,
	markedAlteredAccounts map[string]*markedAlteredAccount,
) error {
	var err error
	for _, txLog := range txPool.Logs {
		for _, event := range txLog.LogHandler.GetLogEvents() {
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
	_, isEsdtOperation := tp.fungibleTokensIdentifiers[string(event.GetIdentifier())]
	if isEsdtOperation {
		err := tp.extractEsdtData(event, zeroBigInt, markedAlteredAccounts)
		if err != nil {
			log.Debug("cannot extract esdt data", "error", err)
			return err
		}

		return nil
	}

	_, isNftOperation := tp.nonFungibleTokensIdentifier[string(event.GetIdentifier())]
	if isNftOperation {
		topics := event.GetTopics()
		if len(topics) == 0 {
			return nil
		}

		nonce := topics[idxTokenNonceInTopics]
		nonceBigInt := big.NewInt(0).SetBytes(nonce)
		err := tp.extractEsdtData(event, nonceBigInt, markedAlteredAccounts)
		if err != nil {
			log.Debug("cannot extract nft data", "error", err)
			return nil
		}

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

	tokenID := topics[idxTokenIDInTopics]
	err := tp.processEsdtDataForAddress(address, nonce, string(tokenID), markedAlteredAccounts)
	if err != nil {
		return err
	}

	// in case of esdt transfer, nft transfer, wipe or multi esdt transfers, the 3rd index of the topics contains the destination address
	identifier := string(event.GetIdentifier())
	eventShouldContainReceiverAddress := identifier == core.BuiltInFunctionESDTTransfer ||
		identifier == core.BuiltInFunctionESDTNFTTransfer ||
		identifier == core.BuiltInFunctionESDTWipe ||
		identifier == core.BuiltInFunctionMultiESDTNFTTransfer

	if eventShouldContainReceiverAddress && len(topics) > idxReceiverAddressInTopics {
		destinationAddress := topics[idxReceiverAddressInTopics]
		err = tp.processEsdtDataForAddress(destinationAddress, nonce, string(tokenID), markedAlteredAccounts)
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
		return nil
	}

	markedAccount.tokens[tokenKey] = &markedAlteredAccountToken{
		identifier: tokenID,
		nonce:      nonce.Uint64(),
	}

	return nil
}

func (tp *tokensProcessor) isSameShard(address []byte) bool {
	return tp.shardCoordinator.SelfId() == tp.shardCoordinator.ComputeId(address)
}
