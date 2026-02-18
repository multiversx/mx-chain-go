package state

import (
	"context"
	"math/big"

	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// lightAccountInfo is a minimal, read-only representation of an on-chain account,
// holding only nonce, balance, and codeMetadata (for guardian checks).
//
// Why this exists:
// During the Supernova proposal phase, the transaction selection flow needs to read
// account nonce, balance, and guarded status for thousands of unique senders. The standard
// path (AccountsDB.getAccount → accountFactory.CreateAccount → Unmarshal) creates a full
// UserAccount object for each address, which includes:
//   - A TrackableDataTrie (tracks storage key changes for the VM)
//   - A DataTrieLeafParser (parses storage trie leaves)
//   - A UserAccount struct with all 9 protobuf fields
//
// None of these are needed — the proposal phase only calls GetNonce(), GetBalance(), and
// IsGuarded(). Creating 10k full UserAccount objects adds ~5-6ms of pure object creation overhead.
//
// lightAccountInfo bypasses all of that by unmarshalling only the UserAccountData protobuf
// (which directly contains Nonce, Balance, and CodeMetadata) without creating the heavy
// supporting objects. This implements UserAccountHandler so it can be stored in the
// AccountsEphemeralProvider cache alongside full UserAccount objects — the cache consumers
// never know the difference.
//
// IMPORTANT: This struct is intentionally read-only. All mutating methods (SetCode, SetRootHash,
// AddToBalance, SaveKeyValue, etc.) panic if called. This is safe because the proposal phase
// never modifies accounts — it only reads nonce, balance, and guarded status for breadcrumb
// validation and virtual session creation. If a code path ever needs to mutate an account or
// access full state (e.g., GetActiveGuardian which needs the data trie), it must go through
// the standard LoadAccount → SaveAccount path. The AccountsEphemeralProvider.GetUserAccount()
// method transparently upgrades lightAccountInfo to full accounts when needed.
type lightAccountInfo struct {
	address      []byte
	nonce        uint64
	balance      *big.Int
	codeMetadata []byte
	rootHash     []byte // data trie root hash — used to skip main trie re-read during upgrade
}

// compile-time interface check
var _ UserAccountHandler = (*lightAccountInfo)(nil)

// NewLightAccountInfo creates a lightweight account info from pre-extracted fields.
// Exported so that state/factory can construct lightAccountInfo from UserAccountData
// without circular imports (state/factory imports state, but state cannot import state/accounts).
func NewLightAccountInfo(address []byte, nonce uint64, balance *big.Int, codeMetadata []byte, rootHash []byte) *lightAccountInfo {
	return &lightAccountInfo{
		address:      address,
		nonce:        nonce,
		balance:      balance,
		codeMetadata: codeMetadata,
		rootHash:     rootHash,
	}
}

// --- Read-only methods (the only ones used in the proposal phase) ---

// GetNonce returns the account nonce.
func (l *lightAccountInfo) GetNonce() uint64 {
	return l.nonce
}

// GetBalance returns a copy of the account balance.
func (l *lightAccountInfo) GetBalance() *big.Int {
	if l.balance == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(l.balance)
}

// AddressBytes returns the account address.
func (l *lightAccountInfo) AddressBytes() []byte {
	return l.address
}

// IsInterfaceNil returns true if there is no value under the interface.
func (l *lightAccountInfo) IsInterfaceNil() bool {
	return l == nil
}

// --- Stub methods required by UserAccountHandler interface ---
// These are never called in the proposal phase. They exist only to satisfy the interface
// so lightAccountInfo can be stored in the same cache as full UserAccount objects.
// All mutating operations panic to catch accidental misuse early.
// Read-only getters return zero values.

func (l *lightAccountInfo) IncreaseNonce(_ uint64) {
	panic("lightAccountInfo: IncreaseNonce called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) SetCode(_ []byte) {
	panic("lightAccountInfo: SetCode called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) SetCodeMetadata(_ []byte) {
	panic("lightAccountInfo: SetCodeMetadata called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) GetCodeMetadata() []byte { return l.codeMetadata }

func (l *lightAccountInfo) SetCodeHash(_ []byte) {
	panic("lightAccountInfo: SetCodeHash called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) GetCodeHash() []byte { return nil }

func (l *lightAccountInfo) SetRootHash(_ []byte) {
	panic("lightAccountInfo: SetRootHash called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) GetRootHash() []byte { return l.rootHash }

func (l *lightAccountInfo) SetDataTrie(_ common.Trie) {
	panic("lightAccountInfo: SetDataTrie called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) DataTrie() common.DataTrieHandler { return nil }

func (l *lightAccountInfo) RetrieveValue(_ []byte) ([]byte, uint32, error) {
	panic("lightAccountInfo: RetrieveValue called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) SaveKeyValue(_ []byte, _ []byte) error {
	panic("lightAccountInfo: SaveKeyValue called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) AddToBalance(_ *big.Int) error {
	panic("lightAccountInfo: AddToBalance called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) SubFromBalance(_ *big.Int) error {
	panic("lightAccountInfo: SubFromBalance called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) ClaimDeveloperRewards(_ []byte) (*big.Int, error) {
	panic("lightAccountInfo: ClaimDeveloperRewards called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) AddToDeveloperReward(_ *big.Int) {
	panic("lightAccountInfo: AddToDeveloperReward called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) GetDeveloperReward() *big.Int { return big.NewInt(0) }

func (l *lightAccountInfo) ChangeOwnerAddress(_ []byte, _ []byte) error {
	panic("lightAccountInfo: ChangeOwnerAddress called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) SetOwnerAddress(_ []byte) {
	panic("lightAccountInfo: SetOwnerAddress called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) GetOwnerAddress() []byte { return nil }

func (l *lightAccountInfo) SetUserName(_ []byte) {
	panic("lightAccountInfo: SetUserName called on read-only lightweight account (proposal phase only)")
}

func (l *lightAccountInfo) GetUserName() []byte { return nil }

// IsGuarded returns true if the account has the guarded flag set in its code metadata.
func (l *lightAccountInfo) IsGuarded() bool {
	codeMetaData := vmcommon.CodeMetadataFromBytes(l.codeMetadata)
	return codeMetaData.Guarded
}

func (l *lightAccountInfo) GetAllLeaves(_ *common.TrieIteratorChannels, _ context.Context) error {
	panic("lightAccountInfo: GetAllLeaves called on read-only lightweight account (proposal phase only)")
}
