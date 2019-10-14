package mock

import "math/big"

type BlockChainHookStub struct {
	AccountExtistsCalled func(address []byte) (bool, error)
	NewAddressCalled     func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	GetBalanceCalled     func(address []byte) (*big.Int, error)
	GetNonceCalled       func(address []byte) (uint64, error)
	GetStorageDataCalled func(accountsAddress []byte, index []byte) ([]byte, error)
	IsCodeEmptyCalled    func(address []byte) (bool, error)
	GetCodeCalled        func(address []byte) ([]byte, error)
	GetBlockHashCalled   func(offset *big.Int) ([]byte, error)
}

func (b *BlockChainHookStub) AccountExists(address []byte) (bool, error) {
	if b.AccountExtistsCalled != nil {
		return b.AccountExtistsCalled(address)
	}
	return false, nil
}

func (b *BlockChainHookStub) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	if b.NewAddressCalled != nil {
		return b.NewAddressCalled(creatorAddress, creatorNonce, vmType)
	}
	return []byte("newAddress"), nil
}

func (b *BlockChainHookStub) GetBalance(address []byte) (*big.Int, error) {
	if b.GetBalanceCalled != nil {
		return b.GetBalanceCalled(address)
	}
	return big.NewInt(0), nil
}

func (b *BlockChainHookStub) GetNonce(address []byte) (uint64, error) {
	if b.GetNonceCalled != nil {
		return b.GetNonceCalled(address)
	}
	return 0, nil
}

func (b *BlockChainHookStub) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	if b.GetStorageDataCalled != nil {
		return b.GetStorageDataCalled(accountAddress, index)
	}
	return nil, nil
}

func (b *BlockChainHookStub) IsCodeEmpty(address []byte) (bool, error) {
	if b.IsCodeEmptyCalled != nil {
		return b.IsCodeEmptyCalled(address)
	}
	return true, nil
}

func (b *BlockChainHookStub) GetCode(address []byte) ([]byte, error) {
	if b.GetCodeCalled != nil {
		return b.GetCodeCalled(address)
	}
	return nil, nil
}

func (b *BlockChainHookStub) GetBlockhash(offset *big.Int) ([]byte, error) {
	if b.GetBlockHashCalled != nil {
		return b.GetBlockHashCalled(offset)
	}
	return []byte("roothash"), nil
}
