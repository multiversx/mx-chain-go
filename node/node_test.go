package node_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getAccAdapter(balance *big.Int) *mock.AccountsStub {
	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &state.Account{Nonce: 1, Balance: balance}, nil
	}
	return accDB
}

func getPrivateKey() *mock.PrivateKeyStub {
	return &mock.PrivateKeyStub{}
}

func containString(search string, list []string) bool {
	for _, str := range list {
		if str == search {
			return true
		}
	}

	return false
}

func getMessenger() *mock.MessengerStub {
	messenger := &mock.MessengerStub{
		CloseCalled: func() error {
			return nil
		},
		BootstrapCalled: func() error {
			return nil
		},
		BroadcastCalled: func(topic string, buff []byte) {
			return
		},
	}

	return messenger
}

func TestNewNode(t *testing.T) {
	n, err := node.NewNode()
	assert.NotNil(t, n)
	assert.Nil(t, err)
}

func TestNewNode_NotRunning(t *testing.T) {
	n, _ := node.NewNode()
	assert.False(t, n.IsRunning())
}

func TestNewNode_NilOptionShouldError(t *testing.T) {
	_, err := node.NewNode(node.WithAccountsAdapter(nil))
	assert.NotNil(t, err)
}

func TestNewNode_ApplyNilOptionShouldError(t *testing.T) {

	n, _ := node.NewNode()
	err := n.ApplyOptions(node.WithAccountsAdapter(nil))
	assert.NotNil(t, err)
}

func TestStart_NoMessenger(t *testing.T) {
	n, _ := node.NewNode()
	err := n.Start()
	defer func() { _ = n.Stop() }()
	assert.NotNil(t, err)
}

func TestStart_CorrectParams(t *testing.T) {

	messenger := getMessenger()
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	err := n.Start()
	defer func() { _ = n.Stop() }()
	assert.Nil(t, err)
	assert.True(t, n.IsRunning())
}

func TestStart_CorrectParamsApplyingOptions(t *testing.T) {

	n, _ := node.NewNode()
	messenger := getMessenger()
	err := n.ApplyOptions(
		node.WithMessenger(messenger),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)

	logError(err)

	err = n.Start()
	defer func() { _ = n.Stop() }()
	assert.Nil(t, err)
	assert.True(t, n.IsRunning())
}

func TestApplyOptions_NodeStarted(t *testing.T) {

	messenger := getMessenger()
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
	)
	err := n.Start()
	defer func() { _ = n.Stop() }()
	logError(err)

	assert.True(t, n.IsRunning())
}

func TestStop_NotStartedYet(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
	)

	err := n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestStop_MessengerCloseErrors(t *testing.T) {
	errorString := "messenger close error"
	messenger := getMessenger()
	messenger.CloseCalled = func() error {
		return errors.New(errorString)
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
	)

	_ = n.Start()

	err := n.Stop()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errorString)
}

func TestStop(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
	)
	err := n.Start()
	logError(err)

	err = n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestGetBalance_NoAddrConverterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxSignPrivKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithTxSignPrivKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_CreateAddressFailsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.AddressConverterStub{
		CreateAddressFromHexHandler: func(hexAddress string) (state.AddressContainer, error) {
			// Return that will result in a correct run of GenerateTransaction -> will fail test
			/*return mock.AddressContainerStub{
			}, nil*/

			return nil, errors.New("error")
		},
	}
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestGetBalance_GetAccountFailsShouldError(t *testing.T) {

	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(addrContainer state.AddressContainer) (state.AccountHandler, error) {
			return nil, errors.New("error")
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
	)
	_, err := n.GetBalance(createDummyHexAddress(64))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "could not fetch sender address from provided param")
}

func createDummyHexAddress(chars int) string {
	if chars < 1 {
		return ""
	}

	buff := make([]byte, chars/2)
	_, _ = rand.Reader.Read(buff)

	return hex.EncodeToString(buff)
}

func TestGetBalance_GetAccountReturnsNil(t *testing.T) {

	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(addrContainer state.AddressContainer) (state.AccountHandler, error) {
			return nil, nil
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
	)
	balance, err := n.GetBalance(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestGetBalance(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(100))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
	)
	balance, err := n.GetBalance(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(100), balance)
}

//------- GenerateTransaction

func TestGenerateTransaction_NoAddrConverterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxSignPrivKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithTxSignPrivKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoPrivateKeyShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_CreateAddressFailsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountFailsShouldError(t *testing.T) {

	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(addrContainer state.AddressContainer) (state.AccountHandler, error) {
			return nil, nil
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(&mock.SinglesignMock{}),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountReturnsNilShouldWork(t *testing.T) {

	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(addrContainer state.AddressContainer) (state.AccountHandler, error) {
			return &state.Account{}, nil
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
}

func TestGenerateTransaction_GetExistingAccountShouldWork(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
}

func TestGenerateTransaction_MarshalErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}
	marshalizer := mock.MarshalizerMock{
		MarshalHandler: func(obj interface{}) ([]byte, error) {
			return nil, errors.New("error")
		},
	}
	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_SignTxErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignFailMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_ShouldSetCorrectSignature(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	signature := []byte("signed")
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(singleSigner),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
	assert.Equal(t, signature, tx.Signature)
}

func TestGenerateTransaction_ShouldSetCorrectNonce(t *testing.T) {

	nonce := uint64(7)
	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(addrContainer state.AddressContainer) (state.AccountHandler, error) {
			return &state.Account{
				Nonce:   nonce,
				Balance: big.NewInt(0),
			}, nil
		},
	}

	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(singleSigner),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
}

func TestGenerateTransaction_CorrectParamsShouldNotError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSignPrivKey(privateKey),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
}

func TestSendTransaction_ShouldWork(t *testing.T) {
	txSent := false
	mes := &mock.MessengerStub{
		BroadcastOnChannelCalled: func(pipe string, topic string, buff []byte) {
			txSent = true
		},
	}

	marshalizer := &mock.MarshalizerFake{}
	hasher := &mock.HasherFake{}
	adrConverter := mock.NewAddressConverterFake(32, "0x")

	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithAddressConverter(adrConverter),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithMessenger(mes),
		node.WithHasher(hasher),
	)

	nonce := uint64(50)
	value := big.NewInt(567)
	sender := createDummyHexAddress(64)
	receiver := createDummyHexAddress(64)
	txData := "data"
	signature := []byte("signature")

	senderBuff, _ := adrConverter.CreateAddressFromHex(sender)
	receiverBuff, _ := adrConverter.CreateAddressFromHex(receiver)

	txHexHashResulted, err := n.SendTransaction(
		nonce,
		sender,
		receiver,
		value,
		0,
		0,
		txData,
		signature)

	marshalizedTx, _ := marshalizer.Marshal(&transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   senderBuff.Bytes(),
		RcvAddr:   receiverBuff.Bytes(),
		Data:      txData,
		Signature: signature,
	})
	txHexHashExpected := hex.EncodeToString(hasher.Compute(string(marshalizedTx)))

	assert.Nil(t, err)
	assert.Equal(t, txHexHashExpected, txHexHashResulted)
	assert.True(t, txSent)
}

func TestCreateShardedStores_NilShardCoordinatorShouldError(t *testing.T) {
	messenger := getMessenger()
	dataPool := &mock.PoolsHolderStub{}

	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	err := n.Start()
	logError(err)
	defer func() { _ = n.Stop() }()
	err = n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil shard coordinator")
}

func TestCreateShardedStores_NilDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	err := n.Start()
	logError(err)
	defer func() { _ = n.Stop() }()
	err = n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil data pool")
}

func TestCreateShardedStores_NilTransactionDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	dataPool := &mock.PoolsHolderStub{}
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}
	dataPool.HeadersCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	err := n.Start()
	logError(err)
	defer func() { _ = n.Stop() }()
	err = n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil transaction sharded data store")
}

func TestCreateShardedStores_NilHeaderDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	dataPool := &mock.PoolsHolderStub{}
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() storage.Cacher {
		return nil
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	err := n.Start()
	logError(err)
	defer func() { _ = n.Stop() }()
	err = n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil header sharded data store")
}

func TestCreateShardedStores_ReturnsSuccessfully(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	nrOfShards := uint32(2)
	shardCoordinator.SetNoShards(nrOfShards)
	dataPool := &mock.PoolsHolderStub{}

	var txShardedStores []string
	txShardedData := &mock.ShardedDataStub{}
	txShardedData.CreateShardStoreCalled = func(cacherId string) {
		txShardedStores = append(txShardedStores, cacherId)
	}
	headerShardedData := &mock.CacherStub{}
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txShardedData
	}
	dataPool.HeadersCalled = func() storage.Cacher {
		return headerShardedData
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	err := n.Start()
	logError(err)
	defer func() { _ = n.Stop() }()
	err = n.CreateShardedStores()
	assert.Nil(t, err)

	assert.True(t, containString(process.ShardCacherIdentifier(0, 0), txShardedStores))
	assert.True(t, containString(process.ShardCacherIdentifier(0, 1), txShardedStores))
	assert.True(t, containString(process.ShardCacherIdentifier(1, 0), txShardedStores))
}

//------- StartHeartbeat

func TestNode_StartHeartbeatDisabledShouldNotCreateObjects(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             false,
	})

	assert.Nil(t, err)
	assert.Nil(t, n.HeartbeatMonitor())
	assert.Nil(t, n.HeartbeatSender())
	assert.Nil(t, n.GetHeartbeats())
}

func TestNode_StartHeartbeatInvalidMinTimeShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: -1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})

	assert.Equal(t, node.ErrNegativeMinTimeToWaitBetweenBroadcastsInSec, err)
}

func TestNode_StartHeartbeatInvalidMaxTimeShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: -1,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})

	assert.Equal(t, node.ErrNegativeMaxTimeToWaitBetweenBroadcastsInSec, err)
}

func TestNode_StartHeartbeatInvalidDurationShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 1,
		DurationInSecToConsiderUnresponsive: -1,
		Enabled:                             true,
	})

	assert.Equal(t, node.ErrNegativeDurationInSecToConsiderUnresponsive, err)
}

func TestNode_StartHeartbeatInvalidMaxTimeMinTimeShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 1,
		DurationInSecToConsiderUnresponsive: 2,
		Enabled:                             true,
	})

	assert.Equal(t, node.ErrWrongValues, err)
}

func TestNode_StartHeartbeatInvalidMaxTimeDurationShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 2,
		Enabled:                             true,
	})

	assert.Equal(t, node.ErrWrongValues, err)
}

func TestNode_StartHeartbeatNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicCalled: func(name string) bool {
				return false
			},
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1"}}),
		node.WithPrivKey(&mock.PrivateKeyStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "marshalizer")
}

func TestNode_StartHeartbeatNilKeygenShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{}),
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicCalled: func(name string) bool {
				return false
			},
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1"}}),
		node.WithPrivKey(&mock.PrivateKeyStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "key generator")
}

func TestNode_StartHeartbeatHasTopicValidatorShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{}),
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return true
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1"}}),
		node.WithTxSignPrivKey(&mock.PrivateKeyStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})

	assert.Equal(t, node.ErrValidatorAlreadySet, err)
}

func TestNode_StartHeartbeatCreateTopicFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{}),
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			HasTopicCalled: func(name string) bool {
				return false
			},
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return errExpected
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1"}}),
		node.WithTxSignPrivKey(&mock.PrivateKeyStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})

	assert.Equal(t, errExpected, err)
}

func TestNode_StartHeartbeatRegisterMessageProcessorFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{}),
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			HasTopicCalled: func(name string) bool {
				return false
			},
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return errExpected
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1"}}),
		node.WithPrivKey(&mock.PrivateKeyStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})

	assert.Equal(t, errExpected, err)
}

func TestNode_StartHeartbeatShouldWorkAndCallSendHeartbeat(t *testing.T) {
	t.Parallel()

	wasBroadcast := atomic.Value{}
	wasBroadcast.Store(false)
	buffData := []byte("buff data")
	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (bytes []byte, e error) {
				return buffData, nil
			},
		}),
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			HasTopicCalled: func(name string) bool {
				return false
			},
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
			BroadcastCalled: func(topic string, buff []byte) {
				if bytes.Equal(buffData, buff) {
					wasBroadcast.Store(true)
				}
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1"}}),
		node.WithPrivKey(&mock.PrivateKeyStub{
			GeneratePublicHandler: func() crypto.PublicKey {
				return &mock.PublicKeyMock{
					ToByteArrayHandler: func() (i []byte, e error) {
						return []byte("pk1"), nil
					},
				}
			},
		}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})

	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	assert.Equal(t, true, wasBroadcast.Load())
}

func TestNode_StartHeartbeatShouldWorkAndHaveAllPublicKeys(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (bytes []byte, e error) {
				return make([]byte, 0), nil
			},
		}),
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			HasTopicCalled: func(name string) bool {
				return false
			},
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
			BroadcastCalled: func(topic string, buff []byte) {
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1", "pk2"}, 1: {"pk3"}}),
		node.WithPrivKey(&mock.PrivateKeyStub{
			GeneratePublicHandler: func() crypto.PublicKey {
				return &mock.PublicKeyMock{
					ToByteArrayHandler: func() (i []byte, e error) {
						return []byte("pk1"), nil
					},
				}
			},
		}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
	)

	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})
	assert.Nil(t, err)

	elements := n.HeartbeatMonitor().GetHeartbeats()
	assert.Equal(t, 3, len(elements))
}

func TestNode_StartHeartbeatShouldWorkAndCanCallProcessMessage(t *testing.T) {
	t.Parallel()

	var registeredHandler p2p.MessageProcessor

	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (bytes []byte, e error) {
				return make([]byte, 0), nil
			},
		}),
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			HasTopicCalled: func(name string) bool {
				return false
			},
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				registeredHandler = handler
				return nil
			},
			BroadcastCalled: func(topic string, buff []byte) {
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1"}}),
		node.WithPrivKey(&mock.PrivateKeyStub{
			GeneratePublicHandler: func() crypto.PublicKey {
				return &mock.PublicKeyMock{
					ToByteArrayHandler: func() (i []byte, e error) {
						return []byte("pk1"), nil
					},
				}
			},
		}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
	)

	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, registeredHandler)

	err = registeredHandler.ProcessReceivedMessage(nil)
	assert.NotNil(t, err)
	assert.Contains(t, "nil message", err.Error())
}

func TestNode_StartConsensusGenesisBlockNotInitializedShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithBlockChain(&mock.ChainHandlerStub{
			GetGenesisHeaderHashCalled: func() []byte {
				return nil
			},
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		}),
	)

	err := n.StartConsensus()

	assert.Equal(t, node.ErrGenesisBlockNotInitialized, err)

}

func TestNode_CreateMetaGenesisBlockShouldCreateSaveAndStoreMetaBlock(t *testing.T) {
	t.Parallel()

	var savedHeader *block.MetaBlock
	var storedHeaderValues []byte
	var storedHeaderKey []byte
	var storedHeaderHash []byte

	testMarshalizer := &mock.MarshalizerFake{}
	testHasher := &mock.HasherFake{}

	chainHandler := &mock.ChainHandlerStub{
		SetGenesisHeaderCalled: func(gb data.HeaderHandler) error {
			savedHeader = gb.(*block.MetaBlock)
			return nil
		},
		SetGenesisHeaderHashCalled: func(hash []byte) {
			storedHeaderHash = hash
		},
	}
	storageService := &mock.ChainStorerMock{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			storedHeaderKey = key
			storedHeaderValues = value
			return nil
		},
	}

	n, _ := node.NewNode(
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithBlockChain(chainHandler),
		node.WithDataStore(storageService),
	)

	err := n.CreateMetaGenesisBlock()

	assert.Nil(t, err)
	assert.NotNil(t, savedHeader)
	assert.NotNil(t, storedHeaderValues)
	assert.NotNil(t, storedHeaderKey)
	assert.NotNil(t, storedHeaderHash)

	assert.Equal(t, uint64(0), savedHeader.Nonce)
	assert.Equal(t, uint32(0), savedHeader.Round)

	//test saved data matches
	marshalizedSavedHeader, _ := testMarshalizer.Marshal(savedHeader)
	assert.Equal(t, marshalizedSavedHeader, storedHeaderValues)
	savedHeaderHash := testHasher.Compute(string(marshalizedSavedHeader))
	assert.Equal(t, savedHeaderHash, storedHeaderHash)
	assert.Equal(t, savedHeaderHash, storedHeaderKey)
}

//------- GetAccount

func TestNode_GetAccountWithNilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithAddressConverter(mock.NewAddressConverterFake(32, "")),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, recovAccnt)
	assert.Equal(t, node.ErrNilAccountsAdapter, err)
}

func TestNode_GetAccountWithNilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, recovAccnt)
	assert.Equal(t, node.ErrNilAddressConverter, err)
}

func TestNode_GetAccountAddressConverterFailsShouldErr(t *testing.T) {
	t.Parallel()

	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}

	errExpected := errors.New("expected error")
	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressConverter(mock.AddressConverterStub{
			CreateAddressFromHexHandler: func(hexAddress string) (container state.AddressContainer, e error) {
				return nil, errExpected
			},
		}),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, recovAccnt)
	assert.Equal(t, errExpected, err)
}

func TestNode_GetAccountAccountDoesNotExistsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressConverter(mock.NewAddressConverterFake(32, "")),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, err)
	assert.Equal(t, uint64(0), recovAccnt.Nonce)
	assert.Equal(t, big.NewInt(0), recovAccnt.Balance)
	assert.Nil(t, recovAccnt.CodeHash)
	assert.Nil(t, recovAccnt.RootHash)
}

func TestNode_GetAccountAccountsAdapterFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressConverter(mock.NewAddressConverterFake(32, "")),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, recovAccnt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errExpected.Error())
}

func TestNode_GetAccountAccountExistsShouldReturn(t *testing.T) {
	t.Parallel()

	accnt := &state.Account{
		Balance:  big.NewInt(1),
		Nonce:    2,
		RootHash: []byte("root hash"),
		CodeHash: []byte("code hash"),
	}

	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressConverter(mock.NewAddressConverterFake(32, "")),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, err)
	assert.Equal(t, accnt, recovAccnt)
}

func TestNode_AppStatusHandlerWithPrometheusIncrementAndDecrementMetric(t *testing.T) {
	t.Parallel()

	testValue := float64(0)
	metricKey := core.MetricIsSyncing

	ashMock := mock.AppStatusHandlerStub{
		IncrementHandler: func(key string) {
			testValue++
		},
		DecrementHandler: func(key string) {
			testValue--
		},
	}

	n, _ := node.NewNode(
		node.WithAppStatusHandler(&ashMock))

	ash := n.GetAppStatusHandler()

	ash.Increment(metricKey)
	ash.Increment(metricKey)
	assert.Equal(t, float64(2), testValue)

	ash.Decrement(metricKey)
	assert.Equal(t, float64(1), testValue)
}

func TestNode_AppStatusHandlerWithPrometheusSetInt64AndUInt64ValueMetric(t *testing.T) {
	t.Parallel()

	metricKey := core.MetricNonce
	testValue := float64(0)

	aphStub := mock.AppStatusHandlerStub{
		SetInt64ValueHandler: func(key string, value int64) {
			testValue = float64(value)
		},
		SetUInt64ValueHandler: func(key string, value uint64) {
			testValue = float64(value)
		},
	}

	n, _ := node.NewNode(
		node.WithAppStatusHandler(&aphStub))
	aph := n.GetAppStatusHandler()

	aph.SetInt64Value(metricKey, int64(20))
	assert.Equal(t, float64(20), testValue)

	aph.SetUInt64Value(metricKey, uint64(30))
	assert.Equal(t, float64(30), testValue)
}

func TestNode_AppStatusHandlerWithFacadeIncrementAndDecrementMetric(t *testing.T) {
	t.Parallel()

	metricKey := core.MetricCurrentRound
	//testValue := float64(0)
	incrementCalled := make(chan bool)
	decrementCalled := make(chan bool)

	// create a prometheus status handler which will be passed to the facade
	prometheusStub := mock.AppStatusHandlerStub{
		IncrementHandler: func(key string) {
			incrementCalled <- true
		},
		DecrementHandler: func(key string) {
			decrementCalled <- true
		},
	}
	appStatusFacade, err := statusHandler.NewAppStatusFacadeWithHandlers(&prometheusStub)
	assert.Nil(t, err)

	asfStub := mock.AppStatusHandlerStub{
		IncrementHandler: func(key string) {
			appStatusFacade.Increment(key)
		},
		DecrementHandler: func(key string) {
			appStatusFacade.Decrement(key)
		},
	}
	n, _ := node.NewNode(
		node.WithAppStatusHandler(&asfStub))
	asf := n.GetAppStatusHandler()

	asf.Increment(metricKey)

	asf.Decrement(metricKey)

	select {
	case <-incrementCalled:
	case <-decrementCalled:
	case <-time.After(5 * time.Millisecond):
		assert.Fail(t, "Timeout - functions not called")
	}
}
