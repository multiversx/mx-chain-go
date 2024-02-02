package stakingcommon

import (
	"github.com/multiversx/mx-chain-core-go/data/validator"
	"github.com/multiversx/mx-chain-go/common"
)

// ValidatorsProviderStub -
type ValidatorsProviderStub struct {
	GetLatestValidatorsCalled func() map[string]*validator.ValidatorStatistics
	GetAuctionListCalled      func() ([]*common.AuctionListValidatorAPIResponse, error)
}

// GetLatestValidators -
func (vp *ValidatorsProviderStub) GetLatestValidators() map[string]*validator.ValidatorStatistics {
	if vp.GetLatestValidatorsCalled != nil {
		return vp.GetLatestValidatorsCalled()
	}

	return nil
}

// GetAuctionList -
func (vp *ValidatorsProviderStub) GetAuctionList() ([]*common.AuctionListValidatorAPIResponse, error) {
	if vp.GetAuctionListCalled != nil {
		return vp.GetAuctionListCalled()
	}

	return nil, nil
}

// Close -
func (vp *ValidatorsProviderStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (vp *ValidatorsProviderStub) IsInterfaceNil() bool {
	return vp == nil
}
