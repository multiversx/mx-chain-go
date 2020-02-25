package accounts

//func TestAccount_MarshalUnmarshal_ShouldWork(t *testing.T) {
//	t.Parallel()
//
//	addr := &mock.AddressMock{}
//	addrTr := &mock.AccountTrackerStub{}
//	acnt, _ := state.NewAccount(addr, addrTr)
//	acnt.Nonce = 0
//	acnt.Balance = big.NewInt(56)
//	acnt.CodeHash = nil
//	acnt.RootHash = nil
//
//	marshalizer := mock.MarshalizerMock{}
//
//	buff, _ := marshalizer.Marshal(&acnt)
//
//	acntRecovered, _ := state.NewAccount(addr, addrTr)
//	_ = marshalizer.Unmarshal(acntRecovered, buff)
//
//	assert.Equal(t, acnt, acntRecovered)
//}
