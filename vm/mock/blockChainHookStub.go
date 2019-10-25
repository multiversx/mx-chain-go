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

// LastNonce returns the nonce from from the last committed block
func (b *BlockChainHookStub) LastNonce() uint64 {
	panic("not implemented")
}

// LastRound returns the round from the last committed block
func (b *BlockChainHookStub) LastRound() uint64 {
	panic("not implemented")
}

// LastTimeStamp returns the timeStamp from the last committed block
func (b *BlockChainHookStub) LastTimeStamp() uint64 {
	panic("not implemented")
}

// LastRandomSeed returns the random seed from the last committed block
func (b *BlockChainHookStub) LastRandomSeed() []byte {
	panic("not implemented")
}

// LastEpoch returns the epoch from the last committed block
func (b *BlockChainHookStub) LastEpoch() uint32 {
	panic("not implemented")
}

// GetStateRootHash returns the state root hash from the last committed block
func (b *BlockChainHookStub) GetStateRootHash() []byte {
	panic("not implemented")
}

// CurrentNonce returns the nonce from the current block
func (b *BlockChainHookStub) CurrentNonce() uint64 {
	panic("not implemented")
}

// CurrentRound returns the round from the current block
func (b *BlockChainHookStub) CurrentRound() uint64 {
	panic("not implemented")
}

// CurrentTimeStamp return the timestamp from the current block
func (b *BlockChainHookStub) CurrentTimeStamp() uint64 {
	panic("not implemented")
}

// CurrentRandomSeed returns the random seed from the current header
func (b *BlockChainHookStub) CurrentRandomSeed() []byte {
	panic("not implemented")
}

// CurrentEpoch returns the current epoch
func (b *BlockChainHookStub) CurrentEpoch() uint32 {
	panic("not implemented")
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
