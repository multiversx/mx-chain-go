//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. accountWrapperMock.proto
package state

import (
	"context"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
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
	Owner             []byte
	address           []byte
	Balance           *big.Int
	guarded           bool
	trackableDataTrie state.DataTrieTracker
	version           uint8

	SetNonceWithJournalCalled    func(nonce uint64) error           `json:"-"`
	SetCodeHashWithJournalCalled func(codeHash []byte) error        `json:"-"`
	SetCodeWithJournalCalled     func([]byte) error                 `json:"-"`
	AccountDataHandlerCalled     func() vmcommon.AccountDataHandler `json:"-"`
}

var errInsufficientBalance = fmt.Errorf("insufficient balance")

// NewAccountWrapMock -
func NewAccountWrapMock(adr []byte) *AccountWrapMock {
	tdt, _ := trackableDataTrie.NewTrackableDataTrie(
		[]byte("identifier"),
		&hashingMocks.HasherMock{},
		&marshallerMock.MarshalizerMock{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	return &AccountWrapMock{
		address:           adr,
		trackableDataTrie: tdt,
		Balance:           big.NewInt(0),
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
func (awm *AccountWrapMock) AddToBalance(val *big.Int) error {
	newBalance := big.NewInt(0).Add(awm.Balance, val)
	if newBalance.Cmp(big.NewInt(0)) < 0 {
		return errInsufficientBalance
	}
	awm.Balance = newBalance
	return nil
}

// SubFromBalance -
func (awm *AccountWrapMock) SubFromBalance(val *big.Int) error {
	newBalance := big.NewInt(0).Sub(awm.Balance, val)
	if newBalance.Cmp(big.NewInt(0)) < 0 {
		return errInsufficientBalance
	}
	awm.Balance = newBalance
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
func (awm *AccountWrapMock) SetOwnerAddress(owner []byte) {
	awm.Owner = owner
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

// GetCode -
func (awm *AccountWrapMock) GetCode() []byte {
	return awm.code
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
func (awm *AccountWrapMock) SaveDirtyData(trie common.Trie) ([]core.TrieData, error) {
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

// SetVersion -
func (awm *AccountWrapMock) SetVersion(version uint8) {
	awm.version = version
}

// GetVersion -
func (awm *AccountWrapMock) GetVersion() uint8 {
	return awm.version
}

// GetAllLeaves -
func (awm *AccountWrapMock) GetAllLeaves(_ *common.TrieIteratorChannels, _ context.Context) error {
	return nil
}
