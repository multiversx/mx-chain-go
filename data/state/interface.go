package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// HashLength defines how many bytes are used in a hash
const HashLength = 32

// TimeStamp is a moment defined by epoch and round
type TimeStamp struct {
	Epoch uint64
	Round uint64
}

// TimePeriod holds start and end time
type TimePeriod struct {
	StartTime TimeStamp
	EndTime   TimeStamp
}

// SignRate is used to keep the number of success and failed signings
type SignRate struct {
	NrSuccess uint32
	NrFailure uint32
}

// ValidatorApiResponse represents the data which is fetched from each validator for returning it in API call
type ValidatorApiResponse struct {
	NrLeaderSuccess    uint32 `json:"nrLeaderSuccess"`
	NrLeaderFailure    uint32 `json:"nrLeaderFailure"`
	NrValidatorSuccess uint32 `json:"nrValidatorSuccess"`
	NrValidatorFailure uint32 `json:"nrValidatorFailure"`
}

// AddressConverter is used to convert to/from AddressContainer
type AddressConverter interface {
	AddressLen() int
	CreateAddressFromPublicKeyBytes(pubKey []byte) (AddressContainer, error)
	ConvertToHex(addressContainer AddressContainer) (string, error)
	CreateAddressFromHex(hexAddress string) (AddressContainer, error)
	PrepareAddressBytes(addressBytes []byte) ([]byte, error)
	IsInterfaceNil() bool
}

// AddressContainer models what an Address struct should do
type AddressContainer interface {
	Bytes() []byte
	IsInterfaceNil() bool
}

// AccountFactory creates an account of different types
type AccountFactory interface {
	CreateAccount(address AddressContainer) (AccountHandler, error)
	IsInterfaceNil() bool
}

// AccountTracker saves an account state and journalizes new entries
type AccountTracker interface {
	SaveAccount(accountHandler AccountHandler) error
	Journalize(entry JournalEntry)
	IsInterfaceNil() bool
}

// Updater set a new value for a key, implemented by trie
type Updater interface {
	Update(key, value []byte) error
	IsInterfaceNil() bool
}

// AccountHandler models a state account, which can journalize and revert
// It knows about code and data, as data structures not hashes
type AccountHandler interface {
	AddressContainer() AddressContainer
	SetNonce(nonce uint64)
	GetNonce() uint64
	SetCode(code []byte)
	GetCode() []byte
	SetCodeHash([]byte)
	GetCodeHash() []byte
	SetRootHash([]byte)
	GetRootHash() []byte
	SetDataTrie(trie data.Trie)
	DataTrie() data.Trie
	DataTrieTracker() DataTrieTracker
	IsInterfaceNil() bool
}

// PeerAccountHandler models a peer state account, which can journalize a normal account's data
//  with some extra features like signing statistics or rating information
type PeerAccountHandler interface {
	GetBLSPublicKey() []byte
	SetBLSPublicKey([]byte)
	GetSchnorrPublicKey() []byte
	SetSchnorrPublicKey([]byte)
	GetRewardAddress() []byte
	SetRewardAddress([]byte) error
	GetStake() *big.Int
	SetStake(*big.Int)
	GetAccumulatedFees() *big.Int
	SetAccumulatedFees(*big.Int)
	GetJailTime() TimePeriod
	SetJailTime(TimePeriod)
	GetCurrentShardId() uint32
	SetCurrentShardId(uint32)
	GetNextShardId() uint32
	SetNextShardId(uint32)
	GetNodeInWaitingList() bool
	SetNodeInWaitingList(bool)
	GetUnStakedNonce() uint64
	SetUnStakedNonce(uint64)
	IncreaseLeaderSuccessRate(uint32)
	DecreaseLeaderSuccessRate(uint32)
	IncreaseValidatorSuccessRate(uint32)
	DecreaseValidatorSuccessRate(uint32)
	GetNumSelectedInSuccessBlocks() uint32
	SetNumSelectedInSuccessBlocks(uint32)
	GetLeaderSuccessRate() SignRate
	GetValidatorSuccessRate() SignRate
	GetRating() uint32
	SetRating(uint32)
	GetTempRating() uint32
	SetTempRating(uint32)
	ResetAtNewEpoch() error
	AccountHandler
}

// UserAccountHandler models a user account, which can journalize account's data with some extra features
// like balance, developer rewards, owner
type UserAccountHandler interface {
	SetBalance(*big.Int)
	GetBalance() *big.Int
	SetDeveloperReward(*big.Int)
	GetDeveloperReward() *big.Int
	SetOwnerAddress([]byte)
	GetOwnerAddress() []byte
	AccountHandler
}

// DataTrieTracker models what how to manipulate data held by a SC account
type DataTrieTracker interface {
	ClearDataCaches()
	DirtyData() map[string][]byte
	OriginalValue(key []byte) []byte
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte)
	SetDataTrie(tr data.Trie)
	DataTrie() data.Trie
	IsInterfaceNil() bool
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	GetExistingAccount(addressContainer AddressContainer) (AccountHandler, error)
	LoadAccount(address AddressContainer) (AccountHandler, error)
	SaveAccount(account AccountHandler) error
	RemoveAccount(addressContainer AddressContainer) error
	Commit() ([]byte, error)
	JournalLen() int
	RevertToSnapshot(snapshot int) error

	RootHash() ([]byte, error)
	RecreateTrie(rootHash []byte) error
	PruneTrie(rootHash []byte) error
	CancelPrune(rootHash []byte)
	SnapshotState(rootHash []byte)
	SetStateCheckpoint(rootHash []byte)
	IsPruningEnabled() bool
	GetAllLeaves(rootHash []byte) (map[string][]byte, error)
	IsInterfaceNil() bool
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert() (AccountHandler, error)
	IsInterfaceNil() bool
}

// TriesHolder is used to store multiple tries
type TriesHolder interface {
	Put([]byte, data.Trie)
	Get([]byte) data.Trie
	GetAll() []data.Trie
	Reset()
	IsInterfaceNil() bool
}

type ValidatorInfo interface {
	GetPublicKey() []byte
	GetShardId() uint32
	GetList() string
	GetIndex() uint32
	GetTempRating() uint32
	GetRating() uint32
	String() string
}
