package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type AccountWrapperStub struct {
	BaseAccountHandler              func() *state.Account
	AddressContainerHandler         func() state.AddressContainer
	CodeHandler                     func() []byte
	SetCodeHandler                  func(code []byte)
	DataTrieHandler                 func() trie.PatriciaMerkelTree
	SetDataTrieHandler              func(trie trie.PatriciaMerkelTree)
	AppendRegistrationDataHandler   func(data *state.RegistrationData) error
	CleanRegistrationDataHandler    func() error
	TrimLastRegistrationDataHandler func() error
}

func (aw AccountWrapperStub) BaseAccount() *state.Account {
	return aw.BaseAccountHandler()
}
func (aw AccountWrapperStub) AddressContainer() state.AddressContainer {
	return aw.AddressContainerHandler()
}
func (aw AccountWrapperStub) Code() []byte {
	return aw.CodeHandler()
}
func (aw AccountWrapperStub) SetCode(code []byte) {
	aw.SetCodeHandler(code)
}
func (aw AccountWrapperStub) DataTrie() trie.PatriciaMerkelTree {
	return aw.DataTrieHandler()
}
func (aw AccountWrapperStub) SetDataTrie(trie trie.PatriciaMerkelTree) {
	aw.SetDataTrieHandler(trie)
}
func (aw AccountWrapperStub) AppendRegistrationData(data *state.RegistrationData) error {
	return aw.AppendRegistrationDataHandler(data)
}
func (aw AccountWrapperStub) CleanRegistrationData() error {
	return aw.CleanRegistrationDataHandler()
}
func (aw AccountWrapperStub) TrimLastRegistrationData() error {
	return aw.TrimLastRegistrationDataHandler()
}
