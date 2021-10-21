package disabled

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SimpleNFTStorageHandler implements the SimpleNFTStorageHandler interface but does nothing as it is disabled
type SimpleNFTStorageHandler struct {
}

// GetESDTNFTTokenOnDestination is disabled
func (s *SimpleNFTStorageHandler) GetESDTNFTTokenOnDestination(_ vmcommon.UserAccountHandler, _ []byte, _ uint64) (*esdt.ESDigitalToken, bool, error) {
	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, true, nil
}

// SaveNFTMetaDataToSystemAccount is disabled
func (s *SimpleNFTStorageHandler) SaveNFTMetaDataToSystemAccount(_ data.TransactionHandler) error {
	return nil
}

// IsInterfaceNil return true if underlying object is nil
func (s *SimpleNFTStorageHandler) IsInterfaceNil() bool {
	return s == nil
}
