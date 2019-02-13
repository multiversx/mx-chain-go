package node_test

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	transaction2 "github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
)

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
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
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
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
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
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

	err = n.ApplyOptions(
		node.WithContext(context.Background()),
	)

	assert.NotNil(t, err)
	assert.True(t, n.IsRunning())
}

func TestStop_NotStartedYet(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
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
		node.WithContext(context.Background()),
	)

	n.Start()

	err := n.Stop()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errorString)
}

func TestStop(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
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
		node.WithContext(context.Background()),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
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
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestGetBalance_GetAccountFailsShouldError(t *testing.T) {

	accAdapter := mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, errors.New("error")
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GetBalance(createDummyHexAddress(64))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "could not fetch sender address from provided param")
}

func createDummyHexAddress(chars int) string {
	if chars < 1 {
		return ""
	}

	var characters = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

	rdm := rand.New(rand.NewSource(time.Now().Unix()))

	buff := make([]byte, chars)
	for i := 0; i < chars; i++ {
		buff[i] = characters[rdm.Int()%16]
	}

	return string(buff)
}

func TestGetBalance_GetAccountReturnsNil(t *testing.T) {

	accAdapter := mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, nil
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
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
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
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
		node.WithContext(context.Background()),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoPrivateKeyShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
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
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountFailsShouldError(t *testing.T) {

	accAdapter := mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, errors.New("error")
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountReturnsNilShouldWork(t *testing.T) {

	accAdapter := mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, nil
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
}

func TestGenerateTransaction_GetExistingAccountShouldWork(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
}

func TestGenerateTransaction_MarshalErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	marshalizer := mock.MarshalizerMock{
		MarshalHandler: func(obj interface{}) ([]byte, error) {
			return nil, errors.New("error")
		},
	}
	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_SignTxErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := mock.PrivateKeyStub{
		SignHandler: func(message []byte, signer crypto.SingleSigner) ([]byte, error) {
			return nil, errors.New("error")
		},
	}
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_ShouldSetCorrectSignature(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	signature := []byte{69}
	privateKey := mock.PrivateKeyStub{
		SignHandler: func(message []byte, signer crypto.SingleSigner) ([]byte, error) {
			return signature, nil
		},
	}

	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
	assert.Equal(t, signature, tx.Signature)
}

func TestGenerateTransaction_ShouldSetCorrectNonce(t *testing.T) {

	nonce := uint64(7)
	accAdapter := mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return mock.AccountWrapperStub{
				BaseAccountHandler: func() *state.Account {
					return &state.Account{
						Nonce:   nonce,
						Balance: big.NewInt(0),
					}
				},
			}, nil
		},
	}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
}

func TestGenerateTransaction_CorrectParamsShouldNotError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code")
	assert.Nil(t, err)
}

//------- GenerateAndSendBulkTransactions

func TestGenerateAndSendBulkTransactions_ZeroTxShouldErr(t *testing.T) {
	n, _ := node.NewNode()

	err := n.GenerateAndSendBulkTransactions("", big.NewInt(0), 0)
	assert.Equal(t, "can not generate and broadcast 0 transactions", err.Error())
}

func TestGenerateAndSendBulkTransactions_NilAccountAdapterShouldErr(t *testing.T) {
	marshalizer := &mock.MarshalizerFake{}

	mes := &mock.MessengerStub{}
	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	suite := &mock.SuiteMock{}
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	sk, pk := signing.NewKeyGenerator(suite).GeneratePair()

	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(0), 1)
	assert.Equal(t, node.ErrNilAccountsAdapter, err)
}

func TestGenerateAndSendBulkTransactions_NilAddressConverterShouldErr(t *testing.T) {
	marshalizer := &mock.MarshalizerFake{}
	accAdapter := getAccAdapter(big.NewInt(0))
	suite := &mock.SuiteMock{}
	sk, pk := signing.NewKeyGenerator(suite).GeneratePair()
	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(0), 1)
	assert.Equal(t, node.ErrNilAddressConverter, err)
}

func TestGenerateAndSendBulkTransactions_NilPrivateKeyShouldErr(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	suite := &mock.SuiteMock{}
	_, pk := signing.NewKeyGenerator(suite).GeneratePair()
	n, _ := node.NewNode(
		node.WithAccountsAdapter(accAdapter),
		node.WithAddressConverter(addrConverter),
		node.WithPublicKey(pk),
		node.WithMarshalizer(&mock.MarshalizerFake{}),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(0), 1)
	assert.True(t, strings.Contains(err.Error(), "trying to set nil private key"))
}

func TestGenerateAndSendBulkTransactions_NilPublicKeyShouldErr(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	suite := &mock.SuiteMock{}
	sk, _ := signing.NewKeyGenerator(suite).GeneratePair()
	n, _ := node.NewNode(
		node.WithAccountsAdapter(accAdapter),
		node.WithAddressConverter(addrConverter),
		node.WithPrivateKey(sk),
	)

	err := n.GenerateAndSendBulkTransactions("", big.NewInt(0), 1)
	assert.Equal(t, "trying to set nil public key", err.Error())
}

func TestGenerateAndSendBulkTransactions_InvalidReceiverAddressShouldErr(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	suite := &mock.SuiteMock{}
	sk, pk := signing.NewKeyGenerator(suite).GeneratePair()
	n, _ := node.NewNode(
		node.WithAccountsAdapter(accAdapter),
		node.WithAddressConverter(addrConverter),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
	)

	err := n.GenerateAndSendBulkTransactions("", big.NewInt(0), 1)
	assert.Contains(t, err.Error(), "could not create receiver address from provided param")
}

func TestGenerateAndSendBulkTransactions_CreateAddressFromPublicKeyBytesErrorsShouldErr(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := &mock.AddressConverterStub{}
	addrConverter.CreateAddressFromPublicKeyBytesHandler = func(pubKey []byte) (container state.AddressContainer, e error) {
		return nil, errors.New("error")
	}
	suite := &mock.SuiteMock{}
	sk, pk := signing.NewKeyGenerator(suite).GeneratePair()
	n, _ := node.NewNode(
		node.WithAccountsAdapter(accAdapter),
		node.WithAddressConverter(addrConverter),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
	)

	err := n.GenerateAndSendBulkTransactions("", big.NewInt(0), 1)
	assert.Equal(t, "error", err.Error())
}

func TestGenerateAndSendBulkTransactions_MarshalizerErrorsShouldErr(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	marshalizer := &mock.MarshalizerFake{}
	marshalizer.Fail = true
	suite := &mock.SuiteMock{}
	sk, pk := signing.NewKeyGenerator(suite).GeneratePair()
	n, _ := node.NewNode(
		node.WithAccountsAdapter(accAdapter),
		node.WithAddressConverter(addrConverter),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
		node.WithMarshalizer(marshalizer),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(1), 1)
	assert.True(t, strings.Contains(err.Error(), "could not marshal transaction"))
}

func TestGenerateAndSendBulkTransactions_ShouldWork(t *testing.T) {
	marshalizer := &mock.MarshalizerFake{}

	noOfTx := 1000
	mutRecoveredTransactions := &sync.RWMutex{}
	recoveredTransactions := make(map[uint64]*transaction.Transaction)

	topic := p2p.NewTopic(string(factory.TransactionTopic), transaction2.NewInterceptedTransaction(), marshalizer)
	topic.SendData = func(data []byte) error {
		//handler to capture sent data
		tx := transaction.Transaction{}

		err := marshalizer.Unmarshal(&tx, data)
		if err != nil {
			return err
		}

		mutRecoveredTransactions.Lock()
		recoveredTransactions[tx.Nonce] = &tx
		mutRecoveredTransactions.Unlock()

		return nil
	}

	mes := &mock.MessengerStub{}
	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(factory.TransactionTopic) {
			return topic
		}

		return nil
	}

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	suite := &mock.SuiteMock{}
	sk, pk := signing.NewKeyGenerator(suite).GeneratePair()
	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
	)

	n.SetMessenger(mes)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(1), uint64(noOfTx))
	assert.Nil(t, err)
	mutRecoveredTransactions.RLock()
	assert.Equal(t, noOfTx, len(recoveredTransactions))
	mutRecoveredTransactions.RUnlock()
}

func getAccAdapter(balance *big.Int) mock.AccountsAdapterStub {
	return mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return mock.AccountWrapperStub{
				BaseAccountHandler: func() *state.Account {
					return &state.Account{
						Nonce:   1,
						Balance: balance,
					}
				},
			}, nil
		},
	}
}

func getPrivateKey() mock.PrivateKeyStub {
	return mock.PrivateKeyStub{
		SignHandler: func(message []byte, signer crypto.SingleSigner) ([]byte, error) {
			return []byte{2}, nil
		},
	}
}

func TestSendTransaction_TopicDoesNotExistsShouldErr(t *testing.T) {
	n, _ := node.NewNode(
		node.WithAddressConverter(mock.NewAddressConverterFake(32, "0x")),
	)

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	nonce := uint64(50)
	value := big.NewInt(567)
	sender := createDummyHexAddress(64)
	receiver := createDummyHexAddress(64)
	txData := "data"
	signature := []byte("signature")

	tx, err := n.SendTransaction(
		nonce,
		sender,
		receiver,
		value,
		txData,
		signature)

	assert.Equal(t, "could not get transaction topic", err.Error())
	assert.Nil(t, tx)
}

func TestSendTransaction_BroadcastErrShouldErr(t *testing.T) {
	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerFake{}),
		node.WithAddressConverter(mock.NewAddressConverterFake(32, "0x")),
	)

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	broadcastErr := errors.New("failure")

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, &mock.MarshalizerMock{})
	topicTx.SendData = func(data []byte) error {
		return broadcastErr
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(factory.TransactionTopic) {
			return topicTx
		}

		return nil
	}

	nonce := uint64(50)
	value := big.NewInt(567)
	sender := createDummyHexAddress(64)
	receiver := createDummyHexAddress(64)
	txData := "data"
	signature := []byte("signature")

	tx, err := n.SendTransaction(
		nonce,
		sender,
		receiver,
		value,
		txData,
		signature)

	assert.Equal(t, "could not broadcast transaction: "+broadcastErr.Error(), err.Error())
	assert.Nil(t, tx)
}

func TestSendTransaction_ShouldWork(t *testing.T) {
	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerFake{}),
		node.WithAddressConverter(mock.NewAddressConverterFake(32, "0x")),
	)

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	txSent := false

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTx.SendData = func(data []byte) error {
		txSent = true
		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(factory.TransactionTopic) {
			return topicTx
		}

		return nil
	}

	nonce := uint64(50)
	value := big.NewInt(567)
	sender := createDummyHexAddress(64)
	receiver := createDummyHexAddress(64)
	txData := "data"
	signature := []byte("signature")

	tx, err := n.SendTransaction(
		nonce,
		sender,
		receiver,
		value,
		txData,
		signature)

	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.True(t, txSent)
}

func TestCreateShardedStores_NilShardCoordinatorShouldError(t *testing.T) {
	messenger := getMessenger()
	dataPool := &mock.TransientDataPoolMock{}

	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
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
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
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
	dataPool := &mock.TransientDataPoolMock{}
	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
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
	dataPool := &mock.TransientDataPoolMock{}
	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
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
	dataPool := &mock.TransientDataPoolMock{}
	var txShardedDataResult uint32
	txShardedData := &mock.ShardedDataStub{}
	txShardedData.CreateShardStoreCalled = func(destShardID uint32) {
		txShardedDataResult = destShardID
	}
	var headerShardedDataResult uint32
	headerShardedData := &mock.ShardedDataStub{}
	headerShardedData.CreateShardStoreCalled = func(destShardID uint32) {
		headerShardedDataResult = destShardID
	}
	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return txShardedData
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return headerShardedData
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithContext(context.Background()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
	)
	err := n.Start()
	logError(err)
	defer func() { _ = n.Stop() }()
	err = n.CreateShardedStores()
	assert.Nil(t, err)
	assert.Equal(t, txShardedDataResult, nrOfShards-1)
	assert.Equal(t, headerShardedDataResult, nrOfShards-1)
}

func getMessenger() *mock.MessengerStub {
	messenger := mock.NewMessengerStub()
	messenger.BootstrapCalled = func(ctx context.Context) {}
	messenger.CloseCalled = func() error {
		return nil
	}
	return messenger
}
