package interceptedTxMocks

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

// InterceptedTxHandlerStub -
type InterceptedTxHandlerStub struct {
	SenderShardIdCalled                    func() uint32
	ReceiverShardIdCalled                  func() uint32
	NonceCalled                            func() uint64
	SenderAddressCalled                    func() []byte
	FeeCalled                              func() *big.Int
	TransactionCalled                      func() data.TransactionHandler
	GetInterceptorUsedGuardianPubKeyCalled func() []byte
	SetInterceptorUsedGuardianPubKeyCalled func([]byte) error
}

// SenderShardId -
func (iths *InterceptedTxHandlerStub) SenderShardId() uint32 {
	if iths.SenderShardIdCalled != nil {
		return iths.SenderShardIdCalled()
	}
	return 0
}

// ReceiverShardId -
func (iths *InterceptedTxHandlerStub) ReceiverShardId() uint32 {
	if iths.ReceiverShardIdCalled != nil {
		return iths.ReceiverShardIdCalled()
	}
	return 0
}

// Nonce -
func (iths *InterceptedTxHandlerStub) Nonce() uint64 {
	if iths.NonceCalled != nil {
		return iths.NonceCalled()
	}
	return 0
}

// SenderAddress -
func (iths *InterceptedTxHandlerStub) SenderAddress() []byte {
	if iths.SenderAddressCalled != nil {
		return iths.SenderAddressCalled()
	}
	return nil
}

// Fee -
func (iths *InterceptedTxHandlerStub) Fee() *big.Int {
	if iths.FeeCalled != nil {
		return iths.FeeCalled()
	}
	return nil
}

// Transaction -
func (iths *InterceptedTxHandlerStub) Transaction() data.TransactionHandler {
	if iths.TransactionCalled != nil {
		return iths.TransactionCalled()
	}
	return nil
}

// GetInterceptorUsedGuardianPubKey -
func (iths *InterceptedTxHandlerStub) GetInterceptorUsedGuardianPubKey() []byte {
	if iths.GetInterceptorUsedGuardianPubKeyCalled != nil {
		return iths.GetInterceptorUsedGuardianPubKeyCalled()
	}
	return nil
}

// SetInterceptorUsedGuardianPubKey -
func (iths *InterceptedTxHandlerStub) SetInterceptorUsedGuardianPubKey(guardianPubKey []byte) error {
	if iths.SetInterceptorUsedGuardianPubKeyCalled != nil {
		return iths.SetInterceptorUsedGuardianPubKeyCalled(guardianPubKey)
	}
	return nil
}
