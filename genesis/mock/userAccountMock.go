package mock

import (
	"context"
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// ErrNegativeValue -
var ErrNegativeValue = errors.New("negative value provided")

// UserAccountMock -
type UserAccountMock struct {
	BaseAccountMock
	code         []byte
	codeMetadata []byte
	codeHash     []byte
	rootHash     []byte
	BalanceField *big.Int
}

// HasNewCode -
func (uam *UserAccountMock) HasNewCode() bool {
	return false
}

// SetCode -
func (uam *UserAccountMock) SetCode(code []byte) {
	uam.code = code
}

// GetCode -
func (uam *UserAccountMock) GetCode() []byte {
	return uam.code
}

// SetCodeMetadata -
func (uam *UserAccountMock) SetCodeMetadata(codeMetadata []byte) {
	uam.codeMetadata = codeMetadata
}

// GetCodeMetadata -
func (uam *UserAccountMock) GetCodeMetadata() []byte {
	return uam.codeMetadata
}

// SetCodeHash -
func (uam *UserAccountMock) SetCodeHash(bytes []byte) {
	uam.codeHash = bytes
}

// GetCodeHash -
func (uam *UserAccountMock) GetCodeHash() []byte {
	return uam.codeHash
}

// SetRootHash -
func (uam *UserAccountMock) SetRootHash(bytes []byte) {
	uam.rootHash = bytes
}

// GetRootHash -
func (uam *UserAccountMock) GetRootHash() []byte {
	return uam.rootHash
}

// SetDataTrie -
func (uam *UserAccountMock) SetDataTrie(_ common.Trie) {
}

// DataTrie -
func (uam *UserAccountMock) DataTrie() common.DataTrieHandler {
	return nil
}

// RetrieveValue -
func (uam *UserAccountMock) RetrieveValue(_ []byte) ([]byte, uint32, error) {
	return nil, 0, nil
}

// SaveKeyValue -
func (uam *UserAccountMock) SaveKeyValue(_ []byte, _ []byte) error {
	return nil
}

// AddToBalance -
func (uam *UserAccountMock) AddToBalance(value *big.Int) error {
	if value.Cmp(big.NewInt(0)) < 0 {
		return ErrNegativeValue
	}

	uam.BalanceField.Add(uam.BalanceField, value)

	return nil
}

// SubFromBalance -
func (uam *UserAccountMock) SubFromBalance(value *big.Int) error {
	uam.BalanceField.Sub(uam.BalanceField, value)

	return nil
}

// GetBalance -
func (uam *UserAccountMock) GetBalance() *big.Int {
	return uam.BalanceField
}

// ClaimDeveloperRewards -
func (uam *UserAccountMock) ClaimDeveloperRewards(_ []byte) (*big.Int, error) {
	return nil, nil
}

// AddToDeveloperReward -
func (uam *UserAccountMock) AddToDeveloperReward(_ *big.Int) {
}

// GetDeveloperReward -
func (uam *UserAccountMock) GetDeveloperReward() *big.Int {
	return nil
}

// ChangeOwnerAddress -
func (uam *UserAccountMock) ChangeOwnerAddress(_ []byte, _ []byte) error {
	return nil
}

// SetOwnerAddress -
func (uam *UserAccountMock) SetOwnerAddress(_ []byte) {
}

// GetOwnerAddress -
func (uam *UserAccountMock) GetOwnerAddress() []byte {
	return nil
}

// SetUserName -
func (uam *UserAccountMock) SetUserName(_ []byte) {
}

// GetUserName -
func (uam *UserAccountMock) GetUserName() []byte {
	return nil
}

// SaveDirtyData -
func (uam *UserAccountMock) SaveDirtyData(_ common.Trie) ([]state.DataTrieChange, []core.TrieData, error) {
	return nil, nil, nil
}

// IsGuarded -
func (uam *UserAccountMock) IsGuarded() bool {
	return false
}

// GetAllLeaves -
func (uam *UserAccountMock) GetAllLeaves(_ *common.TrieIteratorChannels, _ context.Context) error {
	return nil
}
