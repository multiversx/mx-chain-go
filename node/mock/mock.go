package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type Marshalizer struct {
	MarshalHandler func(obj interface{}) ([]byte, error)
	UnmarshalHandler func(obj interface{}, buff []byte) error
}

func (j Marshalizer) Marshal(obj interface{}) ([]byte, error) {
	if j.MarshalHandler != nil {
		return j.MarshalHandler(obj)
	}
	return nil, nil
}
func (j Marshalizer) Unmarshal(obj interface{}, buff []byte) error {
	if j.UnmarshalHandler != nil {
		return j.UnmarshalHandler(obj, buff)
	}
	return nil
}

type Hasher struct {
}

func (hash Hasher) Compute(s string) []byte {
	return nil
}
func (hash Hasher) EmptyHash() []byte {
	return nil
}
func (Hasher) Size() int {
	return 0
}

type AddressContainer struct {
	BytesHandler func() []byte
}

func (ac AddressContainer) Bytes() []byte {
	return ac.BytesHandler()
}

type PrivateKey struct {
	ToByteArrayHandler func() ([]byte, error)
	SignHandler func(message []byte) ([]byte, error)
	GeneratePublicHandler func() crypto.PublicKey
}

func (sk PrivateKey) ToByteArray() ([]byte, error) {
	return sk.ToByteArrayHandler()
}
func (sk PrivateKey) Sign(message []byte) ([]byte, error) {
	return sk.SignHandler(message)
}
func (sk PrivateKey) GeneratePublic() crypto.PublicKey {
	return sk.GeneratePublicHandler()
}

type AddressConverter struct {
	CreateAddressFromPublicKeyBytesHandler func(pubKey []byte) (state.AddressContainer, error)
	ConvertToHexHandler                    func(addressContainer state.AddressContainer) (string, error)
	CreateAddressFromHexHandler            func(hexAddress string) (state.AddressContainer, error)
	PrepareAddressBytesHandler             func(addressBytes []byte) ([]byte, error)
}

func (ac AddressConverter) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	return ac.CreateAddressFromPublicKeyBytesHandler(pubKey)
}
func (ac AddressConverter) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	return ac.ConvertToHexHandler(addressContainer)
}
func (ac AddressConverter) CreateAddressFromHex(hexAddress string) (state.AddressContainer, error) {
	return ac.CreateAddressFromHexHandler(hexAddress)
}
func (ac AddressConverter) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	return ac.PrepareAddressBytesHandler(addressBytes)
}

type AccountsAdapter struct {
	AddJournalEntryHandler func(je state.JournalEntry)
	CommitHandler func() ([]byte, error)
	GetJournalizedAccountHandler func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error)
	GetExistingAccountHandler func(addressContainer state.AddressContainer) (state.AccountWrapper, error)
	HasAccountHandler func(addressContainer state.AddressContainer) (bool, error)
	JournalLenHandler func() int
	PutCodeHandler func(journalizedAccountWrapper state.JournalizedAccountWrapper, code []byte) error
	RemoveAccountHandler func(addressContainer state.AddressContainer) error
	RemoveCodeHandler func(codeHash []byte) error
	LoadDataTrieHandler func(accountWrapper state.AccountWrapper) error
	RevertToSnapshotHandler func(snapshot int) error
	SaveJournalizedAccountHandler func(journalizedAccountWrapper state.JournalizedAccountWrapper) error
	SaveDataHandler func(journalizedAccountWrapper state.JournalizedAccountWrapper) error
	RootHashHandler func() []byte
	RecreateTrieHandler func(rootHash []byte) error
}
func (ad AccountsAdapter) AddJournalEntry(je state.JournalEntry) {
	ad.AddJournalEntryHandler(je)
}

func (ad AccountsAdapter) Commit() ([]byte, error) {
	return ad.CommitHandler()
}

func (ad AccountsAdapter) GetJournalizedAccount(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
	return ad.GetJournalizedAccountHandler(addressContainer)
}

func (ad AccountsAdapter) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountWrapper, error) {
	return ad.GetExistingAccountHandler(addressContainer)
}

func (ad AccountsAdapter) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	return ad.HasAccountHandler(addressContainer)
}

func (ad AccountsAdapter) JournalLen() int {
	return ad.JournalLenHandler()
}
func (ad AccountsAdapter) PutCode(journalizedAccountWrapper state.JournalizedAccountWrapper, code []byte) error {
	return ad.PutCodeHandler(journalizedAccountWrapper, code)
}
func (ad AccountsAdapter) RemoveAccount(addressContainer state.AddressContainer) error {
	return ad.RemoveAccountHandler(addressContainer)
}
func (ad AccountsAdapter) RemoveCode(codeHash []byte) error {
	return ad.RemoveCodeHandler(codeHash)
}
func (ad AccountsAdapter) LoadDataTrie(accountWrapper state.AccountWrapper) error {
	return ad.LoadDataTrieHandler(accountWrapper)
}
func (ad AccountsAdapter) RevertToSnapshot(snapshot int) error {
	return ad.RevertToSnapshotHandler(snapshot)
}
func (ad AccountsAdapter) SaveJournalizedAccount(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return ad.SaveJournalizedAccountHandler(journalizedAccountWrapper)
}
func (ad AccountsAdapter) SaveData(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return ad.SaveDataHandler(journalizedAccountWrapper)
}
func (ad AccountsAdapter) RootHash() []byte {
	return ad.RootHashHandler()
}
func (ad AccountsAdapter) RecreateTrie(rootHash []byte) error {
	return ad.RecreateTrie(rootHash)
}
type AccountWrapper struct {
	BaseAccountHandler func() *state.Account
	AddressContainerHandler func() state.AddressContainer
	CodeHandler func() []byte
	SetCodeHandler func(code []byte)
	DataTrieHandler func() trie.PatriciaMerkelTree
	SetDataTrieHandler func(trie trie.PatriciaMerkelTree)
	AppendRegistrationDataHandler func(data *state.RegistrationData) error
	CleanRegistrationDataHandler func() error
	TrimLastRegistrationDataHandler func() error
}

func (aw AccountWrapper) BaseAccount() *state.Account {
	return aw.BaseAccountHandler()
}
func (aw AccountWrapper) AddressContainer() state.AddressContainer {
	return aw.AddressContainerHandler()
}
func (aw AccountWrapper) Code() []byte {
	return aw.CodeHandler()
}
func (aw AccountWrapper) SetCode(code []byte) {
	aw.SetCodeHandler(code)
}
func (aw AccountWrapper) DataTrie() trie.PatriciaMerkelTree {
	return aw.DataTrieHandler()
}
func (aw AccountWrapper) SetDataTrie(trie trie.PatriciaMerkelTree) {
	aw.SetDataTrieHandler(trie)
}
func (aw AccountWrapper) AppendRegistrationData(data *state.RegistrationData) error {
	return aw.AppendRegistrationDataHandler(data)
}
func (aw AccountWrapper) CleanRegistrationData() error {
	return aw.CleanRegistrationDataHandler()
}
func (aw AccountWrapper) TrimLastRegistrationData() error {
	return aw.TrimLastRegistrationDataHandler()
}