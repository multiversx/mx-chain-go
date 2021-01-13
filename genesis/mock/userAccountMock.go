package mock

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
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
func (uam UserAccountMock) GetCodeHash() []byte {
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
func (uam *UserAccountMock) SetDataTrie(_ data.Trie) {
}

// DataTrie -
func (uam *UserAccountMock) DataTrie() data.Trie {
	return nil
}

// DataTrieTracker -
func (uam *UserAccountMock) DataTrieTracker() state.DataTrieTracker {
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
