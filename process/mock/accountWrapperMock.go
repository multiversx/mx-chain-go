package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// AccountWrapMock -
type AccountWrapMock struct {
	MockValue         int
	dataTrie          common.Trie
	nonce             uint64
	code              []byte
	codeMetadata      []byte
	codeHash          []byte
	rootHash          []byte
	address           []byte
	trackableDataTrie state.DataTrieTracker

	SetNonceWithJournalCalled    func(nonce uint64) error           `json:"-"`
	SetCodeHashWithJournalCalled func(codeHash []byte) error        `json:"-"`
	SetCodeWithJournalCalled     func(codeHash []byte) error        `json:"-"`
	AccountDataHandlerCalled     func() vmcommon.AccountDataHandler `json:"-"`
}

// NewAccountWrapMock -
func NewAccountWrapMock(adr []byte) *AccountWrapMock {
	return &AccountWrapMock{
		address:           adr,
		trackableDataTrie: state.NewTrackableDataTrie([]byte("identifier"), nil),
	}
}

// HasNewCode -
func (awm *AccountWrapMock) HasNewCode() bool {
	return false
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
	return nil
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

// GetCodeHash -
func (awm *AccountWrapMock) GetCodeHash() []byte {
	return awm.codeHash
}

// RetrieveValueFromDataTrieTracker -
func (awm *AccountWrapMock) RetrieveValueFromDataTrieTracker(key []byte) ([]byte, error) {
	return awm.trackableDataTrie.RetrieveValue(key)
}

// SetCodeHash -
func (awm *AccountWrapMock) SetCodeHash(codeHash []byte) {
	awm.codeHash = codeHash
}

// SetCode -
func (awm *AccountWrapMock) SetCode(code []byte) {
	awm.code = code
}

// SetCodeMetadata -
func (awm *AccountWrapMock) SetCodeMetadata(codeMetadata []byte) {
	awm.codeMetadata = codeMetadata
}

// GetCodeMetadata -
func (awm *AccountWrapMock) GetCodeMetadata() []byte {
	return awm.codeMetadata
}

// GetRootHash -
func (awm *AccountWrapMock) GetRootHash() []byte {
	return awm.rootHash
}

// SetRootHash -
func (awm *AccountWrapMock) SetRootHash(rootHash []byte) {
	awm.rootHash = rootHash
}

// AddressBytes -
func (awm *AccountWrapMock) AddressBytes() []byte {
	return awm.address
}

// DataTrie -
func (awm *AccountWrapMock) DataTrie() common.Trie {
	return awm.dataTrie
}

// SetDataTrie -
func (awm *AccountWrapMock) SetDataTrie(trie common.Trie) {
	awm.dataTrie = trie
	awm.trackableDataTrie.SetDataTrie(trie)
}

// DataTrieTracker -
func (awm *AccountWrapMock) DataTrieTracker() state.DataTrieTracker {
	return awm.trackableDataTrie
}

// AccountDataHandler -
func (awm *AccountWrapMock) AccountDataHandler() vmcommon.AccountDataHandler {
	if awm.AccountDataHandlerCalled != nil {
		return awm.AccountDataHandlerCalled()
	}
	return awm.trackableDataTrie
}

// IncreaseNonce -
func (awm *AccountWrapMock) IncreaseNonce(val uint64) {
	awm.nonce = awm.nonce + val
}

// GetNonce -
func (awm *AccountWrapMock) GetNonce() uint64 {
	return awm.nonce
}

// IsInterfaceNil -
func (awm *AccountWrapMock) IsInterfaceNil() bool {
	return awm == nil
}
