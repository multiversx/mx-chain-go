package mock

import vmcommon "github.com/ElrondNetwork/elrond-vm-common"

// AuctionSCMock -
type AuctionSCMock struct {
	ExecuteCalled func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode
}

// Execute -
func (asc *AuctionSCMock) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if asc.ExecuteCalled != nil {
		return asc.ExecuteCalled(args)
	}

	return vmcommon.UserError
}

// IsInterfaceNil -
func (asc *AuctionSCMock) IsInterfaceNil() bool {
	return asc == nil
}
