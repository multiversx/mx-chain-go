package state

//func getNewAccounts(t *testing.T) *Accounts{
//	tr, err := trie.NewTrie(make([]byte, 0), trie.NewDBWriteCache(mock.NewMemDatabase()))
//	assert.Nil(t, err)
//
//	ac, err := NewAccounts(tr)
//	assert.Nil(t, err)
//
//	return ac
//}
//
//func convertToBase64(buff []byte) string{
//	return base64.StdEncoding.EncodeToString(buff)
//}
//
//func TestNilTrie(t *testing.T){
//	_, err := NewAccounts(nil)
//
//	assert.Equal(t, err, ErrorNilTrie)
//}
//
//func TestGetPut(t *testing.T){
//	ac := getNewAccounts(t)
//
//	key := []byte{65, 66, 67}
//	val := []byte{68, 69, 70}
//
//	err := ac.Put(key, val)
//	assert.Nil(t, err)
//
//	valRetriv, err := ac.Get(key)
//	assert.Nil(t, err)
//	assert.Equal(t, val, valRetriv)
//
//	fmt.Printf("Hash: %s\n", convertToBase64(ac.Root()))
//}
//
//func TestDelete(t *testing.T){
//	ac := getNewAccounts(t)
//
//	key := []byte{65, 66, 67}
//	val := []byte{68, 69, 70}
//
//	err := ac.Put(key, val)
//	assert.Nil(t, err)
//
//	valRetriv, err := ac.Get(key)
//	assert.Nil(t, err)
//	assert.Equal(t, val, valRetriv)
//
//	fmt.Printf("Hash: %s\n", convertToBase64(ac.Root()))
//
//	err = ac.Delete(key)
//	assert.Nil(t, err)
//
//	valRetriv, err = ac.Get(key)
//	assert.Nil(t, err)
//	assert.Nil(t, valRetriv)
//
//	fmt.Printf("Hash: %s\n", convertToBase64(ac.Root()))
//	fmt.Printf("Empty root hash: %s\n", convertToBase64(trie.GetEmptyRoot().Bytes()))
//}
//
//func TestCommit(t *testing.T){
//	ac := getNewAccounts(t)
//
//	key := []byte{65, 66, 67}
//	val := []byte{68, 69, 70}
//
//	err := ac.Put(key, val)
//	assert.Nil(t, err)
//
//	valRetriv, err := ac.Get(key)
//	assert.Nil(t, err)
//	assert.Equal(t, val, valRetriv)
//
//	fmt.Printf("Hash: %s\n", convertToBase64(ac.Root()))
//
//	err = ac.Commit()
//	assert.Nil(t, err)
//
//	trie2, err := trie.NewTrie(ac.Root(), ac.tr.DBW())
//	assert.Nil(t, err)
//
//	fmt.Printf("Hash2: %s\n", convertToBase64(trie2.Root()))
//	assert.Equal(t, ac.Root(), trie2.Root())
//}
//
//func TestUndo(t *testing.T){
//	ac := getNewAccounts(t)
//
//	key := []byte{65, 66, 67}
//	val := []byte{68, 69, 70}
//
//	err := ac.Put(key, val)
//	assert.Nil(t, err)
//
//	valRetriv, err := ac.Get(key)
//	assert.Nil(t, err)
//	assert.Equal(t, val, valRetriv)
//
//	fmt.Printf("Hash: %s\n", convertToBase64(ac.Root()))
//
//	err = ac.Undo()
//	assert.Nil(t, err)
//
//	fmt.Printf("Hash after undo: %s\n", convertToBase64(ac.Root()))
//
//	trie2, err := trie.NewTrie(ac.Root(), ac.tr.DBW())
//	assert.Nil(t, err)
//
//	fmt.Printf("Hash2: %s\n", convertToBase64(trie2.Root()))
//	assert.Equal(t, ac.Root(), trie2.Root())
//}
//
//func TestPutGetSimpleAccount(t *testing.T){
//
//	acnts := getNewAccounts(t)
//
//	accountA := Account{Nonce:20, Balance: big.NewInt(100000)}
//	adrA := HexToAddress("0xd1220A0cf47c7B9BE7a2e6Ba89F429762E7B9Adb")
//
//	acnts.PutAccount(*adrA, accountA)
//
//	accountArecov, err := acnts.GetAccount(*adrA)
//	assert.Nil(t, err)
//
//	assert.Equal(t, accountA, *accountArecov)
//}
//
//func TestPutGetSCAccount(t *testing.T){
//	acnts := getNewAccounts(t)
//
//	accountA := Account{Nonce:20, Balance: big.NewInt(100000), code:[]byte{65, 66, 67}, data: []byte{68, 69, 70}}
//	adrA := HexToAddress("0xd1220A0cf47c7B9BE7a2e6Ba89F429762E7B9Adb")
//
//	acnts.PutAccount(*adrA, accountA)
//
//	accountA.CodeHash = DefHasher.Compute(string(accountA.Code()))
//	accountA.DataHash = DefHasher.Compute(string(accountA.Data()))
//
//	accountArecov, err := acnts.GetAccount(*adrA)
//	assert.Nil(t, err)
//
//	assert.Equal(t, accountA, *accountArecov)
//}
