package accounts

//func TestPeerAccount_MarshalUnmarshal_ShouldWork(t *testing.T) {
//	t.Parallel()
//
//	addr := &mock.AddressMock{}
//	addrTr := &mock.AccountTrackerStub{}
//	acnt, _ := state.NewPeerAccount(addr, addrTr)
//
//	marshalizer := mock.MarshalizerMock{}
//	buff, _ := marshalizer.Marshal(&acnt)
//
//	acntRecovered, _ := state.NewPeerAccount(addr, addrTr)
//	_ = marshalizer.Unmarshal(acntRecovered, buff)
//
//	assert.Equal(t, acnt, acntRecovered)
//}
