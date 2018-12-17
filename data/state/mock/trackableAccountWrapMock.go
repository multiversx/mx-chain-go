package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type TrackableAccountWrapMock struct {
	*state.Account
}

func NewTrackableAccountWrapMock() *TrackableAccountWrapMock {
	return &TrackableAccountWrapMock{Account: state.NewAccount()}
}

func (tawm *TrackableAccountWrapMock) AppendRegistrationData(data *state.RegistrationData) error {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) CleanRegistrationData() error {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) TrimLastRegistrationData() error {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) BaseAccount() *state.Account {
	return tawm.Account
}

func (tawm *TrackableAccountWrapMock) AddressContainer() state.AddressContainer {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) Code() []byte {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) SetCode(code []byte) {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) DataTrie() trie.PatriciaMerkelTree {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) SetDataTrie(trie trie.PatriciaMerkelTree) {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) ClearDataCaches() {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) DirtyData() map[string][]byte {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) OriginalValue(key []byte) []byte {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) RetrieveValue(key []byte) ([]byte, error) {
	panic("implement me")
}

func (tawm *TrackableAccountWrapMock) SaveKeyValue(key []byte, value []byte) {
	panic("implement me")
}
