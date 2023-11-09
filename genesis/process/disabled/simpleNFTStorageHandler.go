package disabled

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// SimpleNFTStorage implements the SimpleNFTStorage interface but does nothing as it is disabled
type SimpleNFTStorage struct {
}

// GetESDTNFTTokenOnDestination is disabled
func (s *SimpleNFTStorage) GetESDTNFTTokenOnDestination(_ vmcommon.UserAccountHandler, _ []byte, _ uint64) (*esdt.ESDigitalToken, bool, error) {
	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, true, nil
}

// SaveNFTMetaDataToSystemAccount is disabled
func (s *SimpleNFTStorage) SaveNFTMetaDataToSystemAccount(_ data.TransactionHandler) error {
	return nil
}

// IsInterfaceNil return true if underlying object is nil
func (s *SimpleNFTStorage) IsInterfaceNil() bool {
	return s == nil
}
