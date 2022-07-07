package testscommon

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// EsdtStorageHandlerStub -
type EsdtStorageHandlerStub struct {
	SaveESDTNFTTokenCalled                               func(senderAddress []byte, acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64, esdtData *esdt.ESDigitalToken, isCreation bool, isReturnWithError bool) ([]byte, error)
	GetESDTNFTTokenOnSenderCalled                        func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, error)
	GetESDTNFTTokenOnDestinationCalled                   func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error)
	WasAlreadySentToDestinationShardAndUpdateStateCalled func(tickerID []byte, nonce uint64, dstAddress []byte) (bool, error)
	SaveNFTMetaDataToSystemAccountCalled                 func(tx data.TransactionHandler) error
	AddToLiquiditySystemAccCalled                        func(esdtTokenKey []byte, nonce uint64, transferValue *big.Int) error
}

// SaveESDTNFTToken -
func (e *EsdtStorageHandlerStub) SaveESDTNFTToken(senderAddress []byte, acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64, esdtData *esdt.ESDigitalToken, isCreation bool, isReturnWithError bool) ([]byte, error) {
	if e.SaveESDTNFTTokenCalled != nil {
		return e.SaveESDTNFTTokenCalled(senderAddress, acnt, esdtTokenKey, nonce, esdtData, isCreation, isReturnWithError)
	}

	return nil, nil
}

// GetESDTNFTTokenOnSender -
func (e *EsdtStorageHandlerStub) GetESDTNFTTokenOnSender(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, error) {
	if e.GetESDTNFTTokenOnSenderCalled != nil {
		return e.GetESDTNFTTokenOnSenderCalled(acnt, esdtTokenKey, nonce)
	}

	return nil, nil
}

// GetESDTNFTTokenOnDestination -
func (e *EsdtStorageHandlerStub) GetESDTNFTTokenOnDestination(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
	if e.GetESDTNFTTokenOnDestinationCalled != nil {
		return e.GetESDTNFTTokenOnDestinationCalled(acnt, esdtTokenKey, nonce)
	}

	return nil, false, nil
}

// WasAlreadySentToDestinationShardAndUpdateState -
func (e *EsdtStorageHandlerStub) WasAlreadySentToDestinationShardAndUpdateState(tickerID []byte, nonce uint64, dstAddress []byte) (bool, error) {
	if e.WasAlreadySentToDestinationShardAndUpdateStateCalled != nil {
		return e.WasAlreadySentToDestinationShardAndUpdateStateCalled(tickerID, nonce, dstAddress)
	}

	return false, nil
}

// SaveNFTMetaDataToSystemAccount -
func (e *EsdtStorageHandlerStub) SaveNFTMetaDataToSystemAccount(tx data.TransactionHandler) error {
	if e.SaveNFTMetaDataToSystemAccountCalled != nil {
		return e.SaveNFTMetaDataToSystemAccountCalled(tx)
	}

	return nil
}

// AddToLiquiditySystemAcc -
func (e *EsdtStorageHandlerStub) AddToLiquiditySystemAcc(esdtTokenKey []byte, nonce uint64, transferValue *big.Int) error {
	if e.AddToLiquiditySystemAccCalled != nil {
		return e.AddToLiquiditySystemAccCalled(esdtTokenKey, nonce, transferValue)
	}

	return nil
}

// IsInterfaceNil -
func (e *EsdtStorageHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
