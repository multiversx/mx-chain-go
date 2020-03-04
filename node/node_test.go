package node_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		},
	}

	return messenger
}

func getMarshalizer() marshal.Marshalizer {
	return &mock.MarshalizerMock{}
}

func getHasher() hashing.Hasher {
	return &mock.HasherMock{}
}

func TestNewNode(t *testing.T) {
	n, err := node.NewNode()

	assert.Nil(t, err)
	assert.False(t, check.IfNil(n))
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
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	err := n.Start()
	defer func() { _ = n.Stop() }()
	assert.Nil(t, err)
	assert.True(t, n.IsRunning())
}

func TestStart_CannotApplyOptions(t *testing.T) {

	messenger := getMessenger()
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	err := n.Start()
	require.Nil(t, err)
	defer func() { _ = n.Stop() }()

	err = n.ApplyOptions(node.WithDataPool(&mock.PoolsHolderStub{}))
	require.Error(t, err)
}

func TestStart_CorrectParamsApplyingOptions(t *testing.T) {

	n, _ := node.NewNode()
	messenger := getMessenger()
	err := n.ApplyOptions(
		node.WithMessenger(messenger),
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
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
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
	)
	err := n.Start()
	defer func() { _ = n.Stop() }()
	logError(err)

	assert.True(t, n.IsRunning())
}

func TestStop_NotStartedYet(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
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
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
	)

	_ = n.Start()

	err := n.Stop()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errorString)
}

func TestStop(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
	)
	err := n.Start()
	logError(err)

	err = n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestGetBalance_NoAddrConverterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_CreateAddressFailsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := &mock.AddressConverterStub{
		CreateAddressFromHexHandler: func(hexAddress string) (state.AddressContainer, error) {
			// Return that will result in a correct run of GenerateTransaction -> will fail test
			/*return mock.AddressContainerStub{
			}, nil*/

			return nil, errors.New("error")
		},
	}
	singleSigner := &mock.SinglesignMock{}
	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
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
	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
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
	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
	)
	balance, err := n.GetBalance(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestGetBalance(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(100))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
	)
	balance, err := n.GetBalance(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(100), balance)
}

//------- GenerateTransaction

func TestGenerateTransaction_NoAddrConverterShouldError(t *testing.T) {
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", &mock.PrivateKeyStub{})
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoPrivateKeyShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", nil)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_CreateAddressFailsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey)
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
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(&mock.SinglesignMock{}),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey)
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
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey)
	assert.Nil(t, err)
}

func TestGenerateTransaction_GetExistingAccountShouldWork(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey)
	assert.Nil(t, err)
}

func TestGenerateTransaction_MarshalErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}
	marshalizer := &mock.MarshalizerMock{
		MarshalHandler: func(obj interface{}) ([]byte, error) {
			return nil, errors.New("error")
		},
	}
	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer, testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_SignTxErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignFailMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_ShouldSetCorrectSignature(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	signature := []byte("signed")
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey)
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
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
}

func TestGenerateTransaction_CorrectParamsShouldNotError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey)
	assert.Nil(t, err)
}

func TestCreateTransaction_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := ""
	sender := ""
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "-"

	tx, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature)

	assert.Nil(t, tx)
	assert.Equal(t, node.ErrNilAddressConverter, err)
}

func TestCreateTransaction_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(&mock.AddressConverterStub{
			CreateAddressFromHexHandler: func(hexAddress string) (container state.AddressContainer, e error) {
				return state.NewAddress([]byte(hexAddress)), nil
			},
		}),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := ""
	sender := ""
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "-"

	tx, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature)

	assert.Nil(t, tx)
	assert.Equal(t, node.ErrNilAccountsAdapter, err)
}

func TestCreateTransaction_InvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(&mock.AddressConverterStub{
			CreateAddressFromHexHandler: func(hexAddress string) (container state.AddressContainer, e error) {
				return state.NewAddress([]byte(hexAddress)), nil
			},
		}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "-"

	tx, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
}

func TestCreateTransaction_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
		node.WithAddressConverter(&mock.AddressConverterStub{
			CreateAddressFromHexHandler: func(hexAddress string) (container state.AddressContainer, e error) {
				return state.NewAddress([]byte(hexAddress)), nil
			},
		}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "617eff4f"

	tx, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature)

	assert.NotNil(t, tx)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
	assert.Equal(t, value, tx.Value)
	assert.True(t, bytes.Equal([]byte(receiver), tx.RcvAddr))
}

func TestSendBulkTransactions_NoTxShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}
	marshalizer := &mock.MarshalizerFake{}
	hasher := &mock.HasherFake{}
	adrConverter := mock.NewAddressConverterFake(32, "0x")
	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer, testSizeCheckDelta),
		node.WithAddressConverter(adrConverter),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithMessenger(mes),
		node.WithHasher(hasher),
	)
	txs := make([]*transaction.Transaction, 0)

	numOfTxsProcessed, err := n.SendBulkTransactions(txs)
	assert.Equal(t, uint64(0), numOfTxsProcessed)
	assert.Equal(t, node.ErrNoTxToProcess, err)
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
		node.WithMarshalizer(marshalizer, testSizeCheckDelta),
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
		value.String(),
		0,
		0,
		[]byte(txData),
		signature)

	marshalizedTx, _ := marshalizer.Marshal(&transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   senderBuff.Bytes(),
		RcvAddr:   receiverBuff.Bytes(),
		Data:      []byte(txData),
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
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
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
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
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
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
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

	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return nil
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
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
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return txShardedData
	}
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithHasher(getHasher()),
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
	}, "v0.1",
		"undefined",
	)

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
	}, "v0.1",
		"undefined",
	)

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
	}, "v0.1",
		"undefined",
	)

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
	}, "v0.1",
		"undefined",
	)

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
	}, "v0.1",
		"undefined",
	)

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
	}, "v0.1",
		"undefined",
	)

	assert.Equal(t, node.ErrWrongValues, err)
}

func TestNode_ConsensusTopicNilShardCoordinator(t *testing.T) {
	t.Parallel()

	messageProc := &mock.HeaderResolverStub{}
	n, _ := node.NewNode()

	err := n.CreateConsensusTopic(messageProc)
	require.Equal(t, node.ErrNilShardCoordinator, err)
}

func TestNode_ConsensusTopicValidatorAlreadySet(t *testing.T) {
	t.Parallel()

	messageProc := &mock.HeaderResolverStub{}
	n, _ := node.NewNode(
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return true
			},
		}),
	)

	err := n.CreateConsensusTopic(messageProc)
	require.Equal(t, node.ErrValidatorAlreadySet, err)
}

func TestNode_ConsensusTopicCreateTopicError(t *testing.T) {
	t.Parallel()

	localError := errors.New("error")
	messageProc := &mock.HeaderResolverStub{}
	n, _ := node.NewNode(
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			HasTopicCalled: func(name string) bool {
				return false
			},
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return localError
			},
		}),
	)

	err := n.CreateConsensusTopic(messageProc)
	require.Equal(t, localError, err)
}

func TestNode_ConsensusTopicNilMessageProcessor(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(node.WithShardCoordinator(&mock.ShardCoordinatorMock{}))

	err := n.CreateConsensusTopic(nil)
	require.Equal(t, node.ErrNilMessenger, err)
}

func TestNode_ValidatorStatisticsApi(t *testing.T) {
	t.Parallel()

	initialPubKeys := make(map[uint32][]string)
	keys := [][]string{{"key0"}, {"key1"}, {"key2"}}
	initialPubKeys[0] = keys[0]
	initialPubKeys[1] = keys[1]
	initialPubKeys[2] = keys[2]

	n, _ := node.NewNode(
		node.WithInitialNodesPubKeys(initialPubKeys),
		node.WithValidatorStatistics(&mock.ValidatorStatisticsProcessorStub{
			GetPeerAccountCalled: func(address []byte) (handler state.PeerAccountHandler, err error) {
				switch {
				case bytes.Equal(address, []byte(keys[0][0])):
					return nil, errors.New("error")
				case bytes.Equal(address, []byte(keys[1][0])):
					return &mock.PeerAccountHandlerMock{}, nil
				default:
					return state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
				}
			},
		}),
	)

	expectedData := &state.ValidatorApiResponse{}
	validatorsData, err := n.ValidatorStatisticsApi()
	require.Equal(t, expectedData, validatorsData[hex.EncodeToString([]byte(keys[2][0]))])
	require.Nil(t, err)
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
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "marshalizer")
}

func TestNode_StartHeartbeatNilKeygenShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
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
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "key generator")
}

func TestNode_StartHeartbeatHasTopicValidatorShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithSingleSigner(&mock.SinglesignMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithMessenger(&mock.MessengerStub{
			HasTopicValidatorCalled: func(name string) bool {
				return true
			},
		}),
		node.WithInitialNodesPubKeys(map[uint32][]string{0: {"pk1"}}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)

	assert.Equal(t, node.ErrValidatorAlreadySet, err)
}

func TestNode_StartHeartbeatCreateTopicFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	n, _ := node.NewNode(
		node.WithMarshalizer(getMarshalizer(), testSizeCheckDelta),
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
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)

	assert.Equal(t, errExpected, err)
}

func TestNode_StartHeartbeatRegisterMessageProcessorFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{}, testSizeCheckDelta),
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
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
		node.WithNetworkShardingCollector(
			&mock.NetworkShardingCollectorStub{
				UpdatePeerIdPublicKeyCalled: func(pid p2p.PeerID, pk []byte) {},
			}),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)

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
		}, testSizeCheckDelta),
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
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
		node.WithNetworkShardingCollector(
			&mock.NetworkShardingCollectorStub{
				UpdatePeerIdPublicKeyCalled: func(pid p2p.PeerID, pk []byte) {},
			}),
	)
	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)

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
		}, testSizeCheckDelta),
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
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
		node.WithNetworkShardingCollector(
			&mock.NetworkShardingCollectorStub{
				UpdatePeerIdPublicKeyCalled: func(pid p2p.PeerID, pk []byte) {},
			}),
	)

	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)
	assert.Nil(t, err)

	elements := n.HeartbeatMonitor().GetHeartbeats()
	assert.Equal(t, 3, len(elements))
}

func TestNode_StartHeartbeatShouldSetNodesFromInitialPubKeysAsValidators(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (bytes []byte, e error) {
				return make([]byte, 0), nil
			},
		}, testSizeCheckDelta),
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
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
		node.WithNetworkShardingCollector(
			&mock.NetworkShardingCollectorStub{
				UpdatePeerIdPublicKeyCalled: func(pid p2p.PeerID, pk []byte) {},
			}),
	)

	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)
	assert.Nil(t, err)

	elements := n.HeartbeatMonitor().GetHeartbeats()
	for _, status := range elements {
		assert.True(t, status.IsValidator)
	}
}

func TestNode_StartHeartbeatShouldWorkAndCanCallProcessMessage(t *testing.T) {
	t.Parallel()

	var registeredHandler p2p.MessageProcessor

	n, _ := node.NewNode(
		node.WithMarshalizer(&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (bytes []byte, e error) {
				return make([]byte, 0), nil
			},
		}, testSizeCheckDelta),
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
		node.WithDataStore(&mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return mock.NewStorerMock()
			},
		}),
		node.WithNetworkShardingCollector(
			&mock.NetworkShardingCollectorStub{
				UpdatePeerIdPublicKeyCalled: func(pid p2p.PeerID, pk []byte) {},
			}),
	)

	err := n.StartHeartbeat(config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 1,
		MaxTimeToWaitBetweenBroadcastsInSec: 2,
		DurationInSecToConsiderUnresponsive: 3,
		Enabled:                             true,
	}, "v0.1",
		"undefined",
	)
	assert.Nil(t, err)
	assert.NotNil(t, registeredHandler)

	err = registeredHandler.ProcessReceivedMessage(nil, nil)
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

func TestStartConsensus_NilSyncTimer(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}

	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
	)

	err := n.StartConsensus()
	assert.Equal(t, chronology.ErrNilSyncTimer, err)
}

func TestStartConsensus_ShardBootstrapperNilAccounts(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	n, _ := node.NewNode(
		node.WithDataPool(&mock.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &mock.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte)) {

					},
				}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {

					},
				}
			},
		}),
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithHasher(&mock.HasherMock{}),
		node.WithMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{
			CheckForkCalled: func() *process.ForkInfo {
				return &process.ForkInfo{}
			},
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 0
			},
			GetCalled: func(round int64) (bootstrapData bootstrapStorage.BootstrapData, err error) {
				return bootstrapStorage.BootstrapData{}, errors.New("localErr")
			},
		}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, state.ErrNilAccountsAdapter, err)
}

func TestStartConsensus_ShardBootstrapperErrorResolver(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	localErr := errors.New("error")
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return nil, localErr
		},
	}

	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithBootStorer(&mock.BoostrapStorerMock{}),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, localErr, err)
}

func TestStartConsensus_ShardBootstrapperNilPoolHolder(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
	}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithBootStorer(&mock.BoostrapStorerMock{}),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestStartConsensus_MetaBootstrapperNilPoolHolder(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	shardingCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardingCoordinator.CurrentShard = core.MetachainShardId
	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}
	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncStub{}),
		node.WithShardCoordinator(shardingCoordinator),
		node.WithDataStore(store),
		node.WithResolversFinder(&mock.ResolversFinderStub{
			IntraShardResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
				return &mock.MiniBlocksResolverStub{}, nil
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{}),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithPendingMiniBlocksHandler(&mock.PendingMiniBlocksHandlerStub{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestStartConsensus_MetaBootstrapperWrongNumberShards(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	shardingCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardingCoordinator.CurrentShard = 2
	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}
	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncStub{}),
		node.WithShardCoordinator(shardingCoordinator),
		node.WithDataStore(store),
	)

	err := n.StartConsensus()
	assert.Equal(t, sharding.ErrShardIdOutOfRange, err)
}

func TestStartConsensus_ShardBootstrapperPubKeyToByteArrayError(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	localErr := errors.New("err")
	n, _ := node.NewNode(
		node.WithDataPool(&mock.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &mock.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte)) {

					},
				}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {

					},
				}
			},
		}),
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithHasher(&mock.HasherMock{}),
		node.WithMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlackListHandler(&mock.RequestedItemsHandlerStub{}),
		node.WithMessenger(&mock.MessengerStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return false
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 0
			},
			GetCalled: func(round int64) (bootstrapData bootstrapStorage.BootstrapData, err error) {
				return bootstrapStorage.BootstrapData{}, errors.New("localErr")
			},
		}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithRequestedItemsHandler(&mock.RequestedItemsHandlerStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithPubKey(&mock.PublicKeyMock{
			ToByteArrayHandler: func() (i []byte, err error) {
				return []byte("nil"), localErr
			},
		}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, localErr, err)
}

func TestStartConsensus_ShardBootstrapperInvalidConsensusType(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithDataPool(&mock.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &mock.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte)) {

					},
				}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {

					},
				}
			},
		}),
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithHasher(&mock.HasherMock{}),
		node.WithMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlackListHandler(&mock.RequestedItemsHandlerStub{}),
		node.WithMessenger(&mock.MessengerStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return false
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 0
			},
			GetCalled: func(round int64) (bootstrapData bootstrapStorage.BootstrapData, err error) {
				return bootstrapStorage.BootstrapData{}, errors.New("localErr")
			},
		}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithRequestedItemsHandler(&mock.RequestedItemsHandlerStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithPubKey(&mock.PublicKeyMock{
			ToByteArrayHandler: func() (i []byte, err error) {
				return []byte("keyBytes"), nil
			},
		}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, sposFactory.ErrInvalidConsensusType, err)
}

func TestStartConsensus_ShardBootstrapper(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithDataPool(&mock.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &mock.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte)) {

					},
				}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {

					},
				}
			},
		}),
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithHasher(&mock.HasherMock{}),
		node.WithMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{
			CheckForkCalled: func() *process.ForkInfo {
				return &process.ForkInfo{}
			},
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
		}),
		node.WithBlackListHandler(&mock.RequestedItemsHandlerStub{}),
		node.WithMessenger(&mock.MessengerStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return false
			},
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			HasTopicCalled: func(name string) bool {
				return true
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 0
			},
			GetCalled: func(round int64) (bootstrapData bootstrapStorage.BootstrapData, err error) {
				return bootstrapStorage.BootstrapData{}, errors.New("localErr")
			},
		}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithRequestedItemsHandler(&mock.RequestedItemsHandlerStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithPubKey(&mock.PublicKeyMock{
			ToByteArrayHandler: func() (i []byte, err error) {
				return []byte("keyBytes"), nil
			},
		}),
		node.WithConsensusType("bls"),
		node.WithPrivKey(&mock.PrivateKeyStub{}),
		node.WithSingleSigner(&mock.SingleSignerMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithChainID([]byte("id")),
		node.WithHeaderSigVerifier(&mock.HeaderSigVerifierStub{}),
		node.WithMultiSigner(&mock.MultisignMock{}),
		node.WithValidatorStatistics(&mock.ValidatorStatisticsProcessorStub{}),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartSubscriber(&mock.EpochStartNotifierStub{}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithNetworkShardingCollector(&mock.NetworkShardingCollectorStub{}),
	)

	// TODO: when feature for starting from a higher epoch number is ready we should add a test for that as well
	err := n.StartConsensus()
	assert.Nil(t, err)
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
		node.WithAddressConverter(&mock.AddressConverterStub{
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

func TestNode_AppStatusHandlersShouldIncrement(t *testing.T) {
	t.Parallel()

	metricKey := core.MetricCurrentRound
	incrementCalled := make(chan bool, 1)

	appStatusHandlerStub := mock.AppStatusHandlerStub{
		IncrementHandler: func(key string) {
			incrementCalled <- true
		},
	}

	n, _ := node.NewNode(
		node.WithAppStatusHandler(&appStatusHandlerStub))
	asf := n.GetAppStatusHandler()

	asf.Increment(metricKey)

	select {
	case <-incrementCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_AppStatusHandlerShouldDecrement(t *testing.T) {
	t.Parallel()

	metricKey := core.MetricCurrentRound
	decrementCalled := make(chan bool, 1)

	appStatusHandlerStub := mock.AppStatusHandlerStub{
		DecrementHandler: func(key string) {
			decrementCalled <- true
		},
	}

	n, _ := node.NewNode(
		node.WithAppStatusHandler(&appStatusHandlerStub))
	asf := n.GetAppStatusHandler()

	asf.Decrement(metricKey)

	select {
	case <-decrementCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_AppStatusHandlerShouldSetInt64Value(t *testing.T) {
	t.Parallel()

	metricKey := core.MetricCurrentRound
	setInt64ValueCalled := make(chan bool, 1)

	appStatusHandlerStub := mock.AppStatusHandlerStub{
		SetInt64ValueHandler: func(key string, value int64) {
			setInt64ValueCalled <- true
		},
	}

	n, _ := node.NewNode(
		node.WithAppStatusHandler(&appStatusHandlerStub))
	asf := n.GetAppStatusHandler()

	asf.SetInt64Value(metricKey, int64(1))

	select {
	case <-setInt64ValueCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_AppStatusHandlerShouldSetUInt64Value(t *testing.T) {
	t.Parallel()

	metricKey := core.MetricCurrentRound
	setUInt64ValueCalled := make(chan bool, 1)

	appStatusHandlerStub := mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			setUInt64ValueCalled <- true
		},
	}

	n, _ := node.NewNode(
		node.WithAppStatusHandler(&appStatusHandlerStub))
	asf := n.GetAppStatusHandler()

	asf.SetUInt64Value(metricKey, uint64(1))

	select {
	case <-setUInt64ValueCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_SendBulkTransactionsMultiShardTxsShouldBeMappedCorrectly(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerFake{}

	mutRecoveredTransactions := &sync.RWMutex{}
	recoveredTransactions := make(map[uint32][]*transaction.Transaction)
	signer := &mock.SinglesignStub{
		VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		items := strings.Split(string(address.Bytes()), "Shard")
		sId, _ := strconv.ParseUint(items[1], 2, 32)
		return uint32(sId)
	}

	var txsToSend []*transaction.Transaction
	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     10,
		Value:     new(big.Int).SetInt64(15),
		RcvAddr:   []byte("receiverShard1"),
		SndAddr:   []byte("senderShard0"),
		GasPrice:  5,
		GasLimit:  11,
		Data:      []byte(""),
		Signature: []byte("sig0"),
	})

	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     11,
		Value:     new(big.Int).SetInt64(25),
		RcvAddr:   []byte("receiverShard1"),
		SndAddr:   []byte("senderShard0"),
		GasPrice:  6,
		GasLimit:  12,
		Data:      []byte(""),
		Signature: []byte("sig1"),
	})

	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     12,
		Value:     new(big.Int).SetInt64(35),
		RcvAddr:   []byte("receiverShard0"),
		SndAddr:   []byte("senderShard1"),
		GasPrice:  7,
		GasLimit:  13,
		Data:      []byte(""),
		Signature: []byte("sig2"),
	})

	wg := sync.WaitGroup{}
	wg.Add(len(txsToSend))

	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		chDone <- struct{}{}
	}()

	mes := &mock.MessengerStub{
		BroadcastOnChannelBlockingCalled: func(pipe string, topic string, buff []byte) error {
			txsBuff := make([][]byte, 0)

			err := marshalizer.Unmarshal(&txsBuff, buff)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			for _, txBuff := range txsBuff {
				tx := transaction.Transaction{}
				errMarshal := marshalizer.Unmarshal(&tx, txBuff)
				require.Nil(t, errMarshal)

				mutRecoveredTransactions.Lock()
				sId := shardCoordinator.ComputeId(state.NewAddress(tx.SndAddr))
				recoveredTransactions[sId] = append(recoveredTransactions[sId], &tx)
				mutRecoveredTransactions.Unlock()

				wg.Done()
			}
			return nil
		},
	}

	dataPool := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return nil
				},
			}
		},
	}
	accAdapter := getAccAdapter(big.NewInt(100))
	addrConverter := mock.NewAddressConverterFake(32, "0x")
	keyGen := &mock.KeyGenMock{
		PublicKeyFromByteArrayMock: func(b []byte) (crypto.PublicKey, error) {
			return nil, nil
		},
	}
	feeHandler := &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return 100
		},
		ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(100)
		},
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
			return nil
		},
	}
	n, _ := node.NewNode(
		node.WithMarshalizer(marshalizer, testSizeCheckDelta),
		node.WithHasher(&mock.HasherMock{}),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithKeyGenForAccounts(keyGen),
		node.WithTxSingleSigner(signer),
		node.WithShardCoordinator(shardCoordinator),
		node.WithMessenger(mes),
		node.WithDataPool(dataPool),
		node.WithTxFeeHandler(feeHandler),
	)

	numTxs, err := n.SendBulkTransactions(txsToSend)
	assert.Equal(t, len(txsToSend), int(numTxs))
	assert.Nil(t, err)

	select {
	case <-chDone:
	case <-time.After(timeoutWait):
		assert.Fail(t, "timout while waiting the broadcast of the generated transactions")
		return
	}

	mutRecoveredTransactions.RLock()
	// check if all txs were recovered and are assigned to correct shards
	recTxsSize := 0
	for sId, txsSlice := range recoveredTransactions {
		for _, tx := range txsSlice {
			if !strings.Contains(string(tx.SndAddr), fmt.Sprint(sId)) {
				assert.Fail(t, "txs were not distributed correctly to shards")
			}
			recTxsSize++
		}
	}

	assert.Equal(t, len(txsToSend), recTxsSize)
	mutRecoveredTransactions.RUnlock()
}
