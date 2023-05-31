//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. accountWrapperMock.proto
package state

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ state.UserAccountHandler = (*AccountWrapMock)(nil)

// AccountWrapMock -
type AccountWrapMock struct {
	AccountWrapMockData
	nonce             uint64
	code              []byte
	CodeHash          []byte
	CodeMetadata      []byte
	RootHash          []byte
	address           []byte
	trackableDataTrie state.DataTrieTracker
	Balance           *big.Int
	guarded           bool

	SetNonceWithJournalCalled    func(nonce uint64) error           `json:"-"`
	SetCodeHashWithJournalCalled func(codeHash []byte) error        `json:"-"`
	SetCodeWithJournalCalled     func([]byte) error                 `json:"-"`
	AccountDataHandlerCalled     func() vmcommon.AccountDataHandler `json:"-"`
}

// NewAccountWrapMock -
func NewAccountWrapMock(adr []byte) *AccountWrapMock {
	return &AccountWrapMock{
		address:           adr,
		trackableDataTrie: state.NewTrackableDataTrie([]byte("identifier"), nil),
	}
}

// SetTrackableDataTrie -
func (awm *AccountWrapMock) SetTrackableDataTrie(tdt state.DataTrieTracker) {
	awm.trackableDataTrie = tdt
}

// SetUserName -
func (awm *AccountWrapMock) SetUserName(_ []byte) {
}

// GetUserName -
func (awm *AccountWrapMock) GetUserName() []byte {
	return nil
}

// AddToBalance -
func (awm *AccountWrapMock) AddToBalance(_ *big.Int) error {
	return nil
}

// SubFromBalance -
func (awm *AccountWrapMock) SubFromBalance(_ *big.Int) error {
	return nil
}

// GetBalance -
func (awm *AccountWrapMock) GetBalance() *big.Int {
	return awm.Balance
}

// ClaimDeveloperRewards -
func (awm *AccountWrapMock) ClaimDeveloperRewards([]byte) (*big.Int, error) {
	return nil, nil
}

// AddToDeveloperReward -
func (awm *AccountWrapMock) AddToDeveloperReward(*big.Int) {

}

// GetDeveloperReward -
func (awm *AccountWrapMock) GetDeveloperReward() *big.Int {
	return nil
}

// ChangeOwnerAddress -
func (awm *AccountWrapMock) ChangeOwnerAddress([]byte, []byte) error {
	return nil
}

// SetOwnerAddress -
func (awm *AccountWrapMock) SetOwnerAddress([]byte) {

}

// GetOwnerAddress -
func (awm *AccountWrapMock) GetOwnerAddress() []byte {
	return nil
}

// IsInterfaceNil -
func (awm *AccountWrapMock) IsInterfaceNil() bool {
	return awm == nil
}

// GetCodeHash -
func (awm *AccountWrapMock) GetCodeHash() []byte {
	return awm.CodeHash
}

// SetCodeHash -
func (awm *AccountWrapMock) SetCodeHash(codeHash []byte) {
	awm.CodeHash = codeHash
}

// SetCode -
func (awm *AccountWrapMock) SetCode(code []byte) {
	awm.code = code
}

// RetrieveValue -
func (awm *AccountWrapMock) RetrieveValue(key []byte) ([]byte, uint32, error) {
	return awm.trackableDataTrie.RetrieveValue(key)
}

// SaveKeyValue -
func (awm *AccountWrapMock) SaveKeyValue(key []byte, value []byte) error {
	return awm.trackableDataTrie.SaveKeyValue(key, value)
}

// HasNewCode -
func (awm *AccountWrapMock) HasNewCode() bool {
	return len(awm.code) > 0
}

// SetCodeMetadata -
func (awm *AccountWrapMock) SetCodeMetadata(codeMetadata []byte) {
	awm.CodeMetadata = codeMetadata
}

// GetCodeMetadata -
func (awm *AccountWrapMock) GetCodeMetadata() []byte {
	return awm.CodeMetadata
}

// GetRootHash -
func (awm *AccountWrapMock) GetRootHash() []byte {
	return awm.RootHash
}

// SetRootHash -
func (awm *AccountWrapMock) SetRootHash(rootHash []byte) {
	awm.RootHash = rootHash
}

// AddressBytes -
func (awm *AccountWrapMock) AddressBytes() []byte {
	return awm.address
}

// DataTrie -
func (awm *AccountWrapMock) DataTrie() common.DataTrieHandler {
	return awm.trackableDataTrie.DataTrie()
}

// SaveDirtyData -
func (awm *AccountWrapMock) SaveDirtyData(trie common.Trie) (map[string][]byte, error) {
	return awm.trackableDataTrie.SaveDirtyData(trie)
}

// SetDataTrie -
func (awm *AccountWrapMock) SetDataTrie(trie common.Trie) {
	awm.trackableDataTrie.SetDataTrie(trie)
}

//IncreaseNonce adds the given value to the current nonce
func (awm *AccountWrapMock) IncreaseNonce(val uint64) {
	awm.nonce = awm.nonce + val
}

// AccountDataHandler -
func (awm *AccountWrapMock) AccountDataHandler() vmcommon.AccountDataHandler {
	if awm.AccountDataHandlerCalled != nil {
		return awm.AccountDataHandlerCalled()
	}
	return awm.trackableDataTrie
}

// GetNonce gets the nonce of the account
func (awm *AccountWrapMock) GetNonce() uint64 {
	return awm.nonce
}

// IsGuarded -
func (awm *AccountWrapMock) IsGuarded() bool {
	return awm.guarded
}
