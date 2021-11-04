package testscommon

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SimpleNFTStorageHandlerStub -
type SimpleNFTStorageHandlerStub struct {
	GetESDTNFTTokenOnDestinationCalled   func(accnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error)
	SaveNFTMetaDataToSystemAccountCalled func(tx data.TransactionHandler) error
}

// GetESDTNFTTokenOnDestination -
func (s *SimpleNFTStorageHandlerStub) GetESDTNFTTokenOnDestination(accnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
	if s.GetESDTNFTTokenOnDestinationCalled != nil {
		return s.GetESDTNFTTokenOnDestinationCalled(accnt, esdtTokenKey, nonce)
	}
	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, true, nil
}

// SaveNFTMetaDataToSystemAccount -
func (s *SimpleNFTStorageHandlerStub) SaveNFTMetaDataToSystemAccount(tx data.TransactionHandler) error {
	if s.SaveNFTMetaDataToSystemAccountCalled != nil {
		return s.SaveNFTMetaDataToSystemAccountCalled(tx)
	}
	return nil
}

// IsInterfaceNil -
func (s *SimpleNFTStorageHandlerStub) IsInterfaceNil() bool {
	return s == nil
}
