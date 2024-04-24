package vmcommonMocks

import (
	"math/big"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// UserAccountStub -
type UserAccountStub struct {
	GetCodeMetadataCalled       func() []byte
	SetCodeMetadataCalled       func(codeMetadata []byte)
	GetCodeHashCalled           func() []byte
	GetRootHashCalled           func() []byte
	AccountDataHandlerCalled    func() vmcommon.AccountDataHandler
	AddToBalanceCalled          func(value *big.Int) error
	SubFromBalanceCalled        func(value *big.Int) error
	GetBalanceCalled            func() *big.Int
	ClaimDeveloperRewardsCalled func([]byte) (*big.Int, error)
	GetDeveloperRewardCalled    func() *big.Int
	ChangeOwnerAddressCalled    func([]byte, []byte) error
	SetOwnerAddressCalled       func([]byte)
	GetOwnerAddressCalled       func() []byte
	SetUserNameCalled           func(userName []byte)
	GetUserNameCalled           func() []byte
	AddressBytesCalled          func() []byte
	IncreaseNonceCalled         func(nonce uint64)
	GetNonceCalled              func() uint64
}

// GetCodeMetadata -
func (uas *UserAccountStub) GetCodeMetadata() []byte {
	if uas.GetCodeMetadataCalled != nil {
		uas.GetCodeMetadataCalled()
	}
	return nil
}

// SetCodeMetadata -
func (uas *UserAccountStub) SetCodeMetadata(codeMetaData []byte) {
	if uas.SetCodeMetadataCalled != nil {
		uas.SetCodeMetadataCalled(codeMetaData)
	}
}

// GetCodeHash -
func (uas *UserAccountStub) GetCodeHash() []byte {
	if uas.GetCodeHashCalled != nil {
		return uas.GetCodeHashCalled()
	}
	return nil
}

// GetRootHash -
func (uas *UserAccountStub) GetRootHash() []byte {
	if uas.GetRootHashCalled != nil {
		return uas.GetRootHashCalled()
	}
	return nil
}

// AccountDataHandler -
func (uas *UserAccountStub) AccountDataHandler() vmcommon.AccountDataHandler {
	if uas.AccountDataHandlerCalled != nil {
		return uas.AccountDataHandlerCalled()
	}
	return nil
}

// AddToBalance -
func (uas *UserAccountStub) AddToBalance(value *big.Int) error {
	if uas.AddToBalanceCalled != nil {
		return uas.AddToBalanceCalled(value)
	}
	return nil
}

// SubFromBalance -
func (uas *UserAccountStub) SubFromBalance(value *big.Int) error {
	if uas.AddToBalanceCalled != nil {
		return uas.SubFromBalanceCalled(value)
	}
	return nil
}

// GetBalance -
func (uas *UserAccountStub) GetBalance() *big.Int {
	if uas.GetBalanceCalled != nil {
		return uas.GetBalanceCalled()
	}
	return nil
}

// ClaimDeveloperRewards -
func (uas *UserAccountStub) ClaimDeveloperRewards(sndAddress []byte) (*big.Int, error) {
	if uas.ClaimDeveloperRewardsCalled != nil {
		return uas.ClaimDeveloperRewardsCalled(sndAddress)
	}
	return nil, nil
}

// GetDeveloperReward -
func (uas *UserAccountStub) GetDeveloperReward() *big.Int {
	if uas.GetDeveloperRewardCalled != nil {
		return uas.GetDeveloperRewardCalled()
	}
	return nil
}

// ChangeOwnerAddress -
func (uas *UserAccountStub) ChangeOwnerAddress(sndAddress []byte, newAddress []byte) error {
	if uas.ChangeOwnerAddressCalled != nil {
		return uas.ChangeOwnerAddressCalled(sndAddress, newAddress)
	}
	return nil
}

// SetOwnerAddress -
func (uas *UserAccountStub) SetOwnerAddress(address []byte) {
	if uas.SetOwnerAddressCalled != nil {
		uas.SetOwnerAddressCalled(address)
	}
}

// GetOwnerAddress -
func (uas *UserAccountStub) GetOwnerAddress() []byte {
	if uas.GetOwnerAddressCalled != nil {
		return uas.GetOwnerAddressCalled()
	}
	return nil
}

// SetUserName -
func (uas *UserAccountStub) SetUserName(userName []byte) {
	if uas.SetUserNameCalled != nil {
		uas.SetUserNameCalled(userName)
	}
}

// GetUserName -
func (uas *UserAccountStub) GetUserName() []byte {
	if uas.GetUserNameCalled != nil {
		return uas.GetUserNameCalled()
	}
	return nil
}

// AddressBytes -
func (uas *UserAccountStub) AddressBytes() []byte {
	if uas.AddressBytesCalled != nil {
		return uas.AddressBytesCalled()
	}
	return nil
}

// IncreaseNonce -
func (uas *UserAccountStub) IncreaseNonce(nonce uint64) {
	if uas.IncreaseNonceCalled != nil {
		uas.IncreaseNonceCalled(nonce)
	}
}

// GetNonce -
func (uas *UserAccountStub) GetNonce() uint64 {
	if uas.GetNonceCalled != nil {
		return uas.GetNonceCalled()
	}
	return 0
}

// IsInterfaceNil -
func (uas *UserAccountStub) IsInterfaceNil() bool {
	return uas == nil
}
