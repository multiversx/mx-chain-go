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

	"github.com/ElrondNetwork/elrond-go/core"
	atomicCore "github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func getAccAdapter(balance *big.Int) *mock.AccountsStub {
	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		acc, _ := state.NewUserAccount(address)
		_ = acc.AddToBalance(balance)
		acc.IncreaseNonce(1)

		return acc, nil
	}
	return accDB
}

func getPrivateKey() *mock.PrivateKeyStub {
	return &mock.PrivateKeyStub{}
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
	return &mock.MarshalizerFake{}
}

func getHasher() hashing.Hasher {
	return &mock.HasherMock{}
}

func TestNewNode(t *testing.T) {
	n, err := node.NewNode()

	assert.Nil(t, err)
	assert.False(t, check.IfNil(n))
}

func TestNewNode_NilOptionShouldError(t *testing.T) {
	_, err := node.NewNode(node.WithCoreComponents(nil))
	assert.NotNil(t, err)
}

func TestNewNode_ApplyNilOptionShouldError(t *testing.T) {
	n, _ := node.NewNode()
	err := n.ApplyOptions(node.WithCoreComponents(nil))
	assert.NotNil(t, err)
}

func TestGetBalance_NoAddrConverterShouldError(t *testing.T) {
	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = nil
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.AddrPubKeyConv = nil
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and PubkeyConverter first", err.Error())
}

func TestGetBalance_NoAccAdapterShouldError(t *testing.T) {
	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = nil
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and PubkeyConverter first", err.Error())
}

func TestGetBalance_GetAccountFailsShouldError(t *testing.T) {
	expectedErr := errors.New("error")

	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return nil, expectedErr
		},
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	_, err := n.GetBalance(createDummyHexAddress(64))
	assert.Equal(t, expectedErr, err)
}

func createDummyHexAddress(hexChars int) string {
	if hexChars < 1 {
		return ""
	}

	buff := make([]byte, hexChars/2)
	_, _ = rand.Reader.Read(buff)

	return hex.EncodeToString(buff)
}

func TestGetBalance_GetAccountReturnsNil(t *testing.T) {

	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return nil, nil
		},
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	balance, err := n.GetBalance(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestGetBalance(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(100))
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	balance, err := n.GetBalance(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(100), balance)
}

func TestGetUsername(t *testing.T) {
	expectedUsername := []byte("elrond")

	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		acc, _ := state.NewUserAccount(address)
		acc.UserName = expectedUsername
		acc.IncreaseNonce(1)

		return acc, nil
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()

	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accDB

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	username, err := n.GetUsername(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, string(expectedUsername), username)
}

//------- GenerateTransaction

func TestGenerateTransaction_NoAddrConverterShouldError(t *testing.T) {
	privateKey := getPrivateKey()
	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = nil
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoAccAdapterShouldError(t *testing.T) {
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	stateComponents.Accounts = nil
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", &mock.PrivateKeyStub{}, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoPrivateKeyShouldError(t *testing.T) {
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", nil, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_CreateAddressFailsShouldError(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountFailsShouldError(t *testing.T) {

	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return nil, nil
		},
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = &mock.SinglesignMock{}

	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountReturnsNilShouldWork(t *testing.T) {

	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return state.NewUserAccount(address)
		},
	}
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
}

func TestGenerateTransaction_GetExistingAccountShouldWork(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
}

func TestGenerateTransaction_MarshalErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}
	marshalizer := &mock.MarshalizerMock{
		MarshalHandler: func(obj interface{}) ([]byte, error) {
			return nil, errors.New("error")
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizer
	coreComponents.VmMarsh = marshalizer
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_SignTxErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignFailMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_ShouldSetCorrectSignature(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	signature := []byte("signed")
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
	assert.Equal(t, signature, tx.Signature)
}

func TestGenerateTransaction_ShouldSetCorrectNonce(t *testing.T) {

	nonce := uint64(7)
	accAdapter := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (state.AccountHandler, error) {
			acc, _ := state.NewUserAccount(address)
			_ = acc.AddToBalance(big.NewInt(0))
			acc.IncreaseNonce(nonce)

			return acc, nil
		},
	}

	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
}

func TestGenerateTransaction_CorrectParamsShouldNotError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)

	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
}

func TestCreateTransaction_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := ""
	sender := ""
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "-"

	coreComponents.AddrPubKeyConv = nil
	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature, "chainID", 1)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.Equal(t, node.ErrNilPubkeyConverter, err)
}

func TestCreateTransaction_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
	}

	stateComponents := getDefaultStateComponents()
	processComponents := getDefaultProcessComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := ""
	sender := ""
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "-"

	stateComponents.Accounts = nil

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature, "chainID", 1)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.Equal(t, node.ErrNilAccountsAdapter, err)
}

func TestCreateTransaction_InvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "-"

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature, "chainID", 1)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.NotNil(t, err)
}

func TestCreateTransaction_InvalidChainIDShouldErr(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "617eff4f"
	_, _, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature, "", 1)
	assert.Equal(t, node.ErrInvalidChainID, err)
}

func TestCreateTransaction_InvalidTxVersionShouldErr(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "617eff4f"
	_, _, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature, "", 0)
	assert.Equal(t, node.ErrInvalidTransactionVersion, err)
}

func TestCreateTransaction_SenderShardIdIsInDifferentShardShouldNotValidate(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
	crtShardID := uint32(1)
	chainID := []byte("chain ID")
	version := uint32(1)
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(
			mock.HasherMock{
				ComputeCalled: func(s string) []byte {
					return expectedHash
				},
			},
		),
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{
			ComputeIdCalled: func(i []byte) uint32 {
				return crtShardID + 1
			},
			SelfShardId: crtShardID,
		}),
		node.WithWhiteListHandler(&mock.WhiteListHandlerStub{}),
		node.WithWhiteListHandlerVerified(&mock.WhiteListHandlerStub{}),
		node.WithKeyGenForAccounts(&mock.KeyGenMock{}),
		node.WithTxSingleSigner(&mock.SingleSignerMock{}),
		node.WithTxFeeHandler(&mock.FeeHandlerStub{
			CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
				return nil
			},
		}),
		node.WithChainID(chainID),
		node.WithMinTransactionVersion(version),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "617eff4f"

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature, string(chainID), version)
	assert.NotNil(t, tx)
	assert.Equal(t, expectedHash, txHash)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
	assert.Equal(t, value, tx.Value)
	assert.True(t, bytes.Equal([]byte(receiver), tx.RcvAddr))

	err = n.ValidateTransaction(tx)
	assert.True(t, errors.Is(err, node.ErrDifferentSenderShardId))
}

func TestCreateTransaction_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}

	processComponents := getDefaultProcessComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "617eff4f"

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, sender, gasPrice, gasLimit, txData, signature, string(chainID), version)
	assert.NotNil(t, tx)
	assert.Equal(t, expectedHash, txHash)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
	assert.Equal(t, value, tx.Value)
	assert.True(t, bytes.Equal([]byte(receiver), tx.RcvAddr))

	err = n.ValidateTransaction(tx)
	assert.Nil(t, err)
}

func TestSendBulkTransactions_NoTxShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}
	marshalizer := &mock.MarshalizerFake{}
	hasher := &mock.HasherFake{}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizer
	coreComponents.VmMarsh = marshalizer
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	coreComponents.Hash = hasher
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = mock.NewOneShardCoordinatorMock()
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = mes

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
	)
	txs := make([]*transaction.Transaction, 0)

	numOfTxsProcessed, err := n.SendBulkTransactions(txs)
	assert.Equal(t, uint64(0), numOfTxsProcessed)
	assert.Equal(t, node.ErrNoTxToProcess, err)
}

func TestCreateShardedStores_NilShardCoordinatorShouldError(t *testing.T) {
	messenger := getMessenger()
	dataPool := testscommon.NewPoolsHolderStub()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	processComponents := getDefaultProcessComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
	)

	processComponents.ShardCoord = nil
	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil shard coordinator")
}

func TestCreateShardedStores_NilDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}
	dataComponents := getDefaultDataComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
	)

	dataComponents.DataPool = nil
	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil data pool")
}

func TestCreateShardedStores_NilTransactionDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	dataPool := testscommon.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
	)

	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil transaction sharded data store")
}

func TestCreateShardedStores_NilHeaderDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	dataPool := testscommon.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}

	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return nil
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
	)

	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil header sharded data store")
}

func TestCreateShardedStores_ReturnsSuccessfully(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	nrOfShards := uint32(2)
	shardCoordinator.SetNoShards(nrOfShards)

	dataPool := testscommon.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = &mock.AccountsStub{}
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
	)

	err := n.CreateShardedStores()
	assert.Nil(t, err)
}

func TestNode_ValidatorStatisticsApi(t *testing.T) {
	t.Parallel()

	initialPubKeys := make(map[uint32][]string)
	keys := [][]string{{"key0"}, {"key1"}, {"key2"}}
	initialPubKeys[0] = keys[0]
	initialPubKeys[1] = keys[1]
	initialPubKeys[2] = keys[2]

	validatorsInfo := make(map[uint32][]*state.ValidatorInfo)

	for shardId, pubkeysPerShard := range initialPubKeys {
		validatorsInfo[shardId] = make([]*state.ValidatorInfo, 0)
		for _, pubKey := range pubkeysPerShard {
			validatorsInfo[shardId] = append(validatorsInfo[shardId], &state.ValidatorInfo{
				PublicKey:                  []byte(pubKey),
				ShardId:                    shardId,
				List:                       "",
				Index:                      0,
				TempRating:                 0,
				Rating:                     0,
				RewardAddress:              nil,
				LeaderSuccess:              0,
				LeaderFailure:              0,
				ValidatorSuccess:           0,
				ValidatorFailure:           0,
				NumSelectedInSuccessBlocks: 0,
				AccumulatedFees:            nil,
				TotalLeaderSuccess:         0,
				TotalLeaderFailure:         0,
				TotalValidatorSuccess:      0,
				TotalValidatorFailure:      0,
			})
		}
	}

	vsp := &mock.ValidatorStatisticsProcessorStub{
		RootHashCalled: func() (i []byte, err error) {
			return []byte("hash"), nil
		},
		GetValidatorInfoForRootHashCalled: func(rootHash []byte) (m map[uint32][]*state.ValidatorInfo, err error) {
			return validatorsInfo, nil
		},
	}

	validatorProvider := &mock.ValidatorsProviderStub{GetLatestValidatorsCalled: func() map[string]*state.ValidatorApiResponse {
		apiResponses := make(map[string]*state.ValidatorApiResponse)

		for _, vis := range validatorsInfo {
			for _, vi := range vis {
				apiResponses[hex.EncodeToString(vi.GetPublicKey())] = &state.ValidatorApiResponse{}
			}
		}

		return apiResponses
	},
	}

	processComponents := getDefaultProcessComponents()
	processComponents.ValidatorProvider = validatorProvider
	processComponents.ValidatorStatistics = vsp

	n, _ := node.NewNode(
		node.WithInitialNodesPubKeys(initialPubKeys),
		node.WithProcessComponents(processComponents),
	)

	expectedData := &state.ValidatorApiResponse{}
	validatorsData, err := n.ValidatorStatisticsApi()
	require.Equal(t, expectedData, validatorsData[hex.EncodeToString([]byte(keys[2][0]))])
	require.Nil(t, err)
}

//------- GetAccount

func TestNode_GetAccountWithNilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	stateComponents.Accounts = nil
	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, recovAccnt)
	assert.Equal(t, node.ErrNilAccountsAdapter, err)
}

func TestNode_GetAccountWithNilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accDB
	coreComponents := getDefaultCoreComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	coreComponents.AddrPubKeyConv = nil
	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, recovAccnt)
	assert.Equal(t, node.ErrNilPubkeyConverter, err)
}

func TestNode_GetAccountPubkeyConverterFailsShouldErr(t *testing.T) {
	t.Parallel()

	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}

	errExpected := errors.New("expected error")
	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return nil, errExpected
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accDB

	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(coreComponents),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, recovAccnt)
	assert.Equal(t, errExpected, err)
}

func TestNode_GetAccountAccountDoesNotExistsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accDB
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, err)
	assert.Equal(t, uint64(0), recovAccnt.GetNonce())
	assert.Equal(t, big.NewInt(0), recovAccnt.GetBalance())
	assert.Nil(t, recovAccnt.GetCodeHash())
	assert.Nil(t, recovAccnt.GetRootHash())
}

func TestNode_GetAccountAccountsAdapterFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accDB
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, recovAccnt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errExpected.Error())
}

func TestNode_GetAccountAccountExistsShouldReturn(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewUserAccount([]byte("1234"))
	_ = accnt.AddToBalance(big.NewInt(1))
	accnt.IncreaseNonce(2)
	accnt.SetRootHash([]byte("root hash"))
	accnt.SetCodeHash([]byte("code hash"))

	accDB := &mock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accDB
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
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

	coreComponents := getDefaultCoreComponents()
	coreComponents.AppStatusHdl = &appStatusHandlerStub

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents))
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

	coreComponents := getDefaultCoreComponents()
	coreComponents.AppStatusHdl = &appStatusHandlerStub

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents))
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

	coreComponents := getDefaultCoreComponents()
	coreComponents.AppStatusHdl = &appStatusHandlerStub

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents))
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

	coreComponents := getDefaultCoreComponents()
	coreComponents.AppStatusHdl = &appStatusHandlerStub

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents))
	asf := n.GetAppStatusHandler()

	asf.SetUInt64Value(metricKey, uint64(1))

	select {
	case <-setUInt64ValueCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_EncodeDecodeAddressPubkey(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")

	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = mock.NewPubkeyConverterMock(32)
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
	)
	encoded, err := n.EncodeAddressPubkey(buff)
	assert.Nil(t, err)

	recoveredBytes, err := n.DecodeAddressPubkey(encoded)

	assert.Nil(t, err)
	assert.Equal(t, buff, recoveredBytes)
}

func TestNode_EncodeDecodeAddressPubkeyWithNilConverterShouldErr(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")

	coreComponents := getDefaultCoreComponents()
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents))

	coreComponents.AddrPubKeyConv = nil
	encoded, err := n.EncodeAddressPubkey(buff)

	assert.Empty(t, encoded)
	assert.True(t, errors.Is(err, node.ErrNilPubkeyConverter))
}

func TestNode_DecodeAddressPubkeyWithNilConverterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
	)

	coreComponents.AddrPubKeyConv = nil
	recoveredBytes, err := n.DecodeAddressPubkey("")

	assert.True(t, errors.Is(err, node.ErrNilPubkeyConverter))
	assert.Nil(t, recoveredBytes)
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
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		items := strings.Split(string(address), "Shard")
		sId, _ := strconv.ParseUint(items[1], 2, 32)
		return uint32(sId)
	}

	var txsToSend []*transaction.Transaction
	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     10,
		Value:     big.NewInt(15),
		RcvAddr:   []byte("receiverShard1"),
		SndAddr:   []byte("senderShard0"),
		GasPrice:  5,
		GasLimit:  11,
		Data:      []byte(""),
		Signature: []byte("sig0"),
	})

	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     11,
		Value:     big.NewInt(25),
		RcvAddr:   []byte("receiverShard1"),
		SndAddr:   []byte("senderShard0"),
		GasPrice:  6,
		GasLimit:  12,
		Data:      []byte(""),
		Signature: []byte("sig1"),
	})

	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     12,
		Value:     big.NewInt(35),
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

			b := &batch.Batch{}
			err := marshalizer.Unmarshal(b, buff)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			for _, txBuff := range b.Data {
				tx := transaction.Transaction{}
				errMarshal := marshalizer.Unmarshal(&tx, txBuff)
				require.Nil(t, errMarshal)

				mutRecoveredTransactions.Lock()
				sId := shardCoordinator.ComputeId(tx.SndAddr)
				recoveredTransactions[sId] = append(recoveredTransactions[sId], &tx)
				mutRecoveredTransactions.Unlock()

				wg.Done()
			}
			return nil
		},
	}

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return nil
				},
			}
		},
	}
	accAdapter := getAccAdapter(big.NewInt(100))
	keyGen := &mock.KeyGenMock{
		PublicKeyFromByteArrayMock: func(b []byte) (crypto.PublicKey, error) {
			return nil, nil
		},
	}
	feeHandler := &mock.EconomicsHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return 100
		},
		ComputeMoveBalanceFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(100)
		},
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
			return nil
		},
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizer
	coreComponents.VmMarsh = marshalizer
	coreComponents.TxMarsh = marshalizer
	coreComponents.Hash = &mock.HasherMock{}
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	coreComponents.EconomicsHandler = feeHandler
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = accAdapter
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = signer
	cryptoComponents.TxKeyGen = keyGen
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = mes

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithProcessComponents(processComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithTxAccumulator(mock.NewAccumulatorMock()),
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

func TestNode_DirectTrigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	epoch := uint32(47839)
	recoveredEpoch := uint32(0)
	recoveredWithEarlyEndOfEpoch := atomicCore.Flag{}
	hardforkTrigger := &mock.HardforkTriggerStub{
		TriggerCalled: func(epoch uint32, withEarlyEndOfEpoch bool) error {
			wasCalled = true
			atomic.StoreUint32(&recoveredEpoch, epoch)
			recoveredWithEarlyEndOfEpoch.Toggle(withEarlyEndOfEpoch)

			return nil
		},
	}
	n, _ := node.NewNode(
		node.WithHardforkTrigger(hardforkTrigger),
	)

	err := n.DirectTrigger(epoch, true)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, epoch, recoveredEpoch)
	assert.True(t, recoveredWithEarlyEndOfEpoch.IsSet())
}

func TestNode_IsSelfTrigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	hardforkTrigger := &mock.HardforkTriggerStub{
		IsSelfTriggerCalled: func() bool {
			wasCalled = true

			return true
		},
	}
	n, _ := node.NewNode(
		node.WithHardforkTrigger(hardforkTrigger),
	)

	isSelf := n.IsSelfTrigger()

	assert.True(t, isSelf)
	assert.True(t, wasCalled)
}

//------- Query handlers

func TestNode_AddQueryHandlerNilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	err := n.AddQueryHandler("handler", nil)

	assert.True(t, errors.Is(err, node.ErrNilQueryHandler))
}

func TestNode_AddQueryHandlerEmptyNameShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	err := n.AddQueryHandler("", &mock.QueryHandlerStub{})

	assert.True(t, errors.Is(err, node.ErrEmptyQueryHandlerName))
}

func TestNode_AddQueryHandlerExistsShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	err := n.AddQueryHandler("handler", &mock.QueryHandlerStub{})
	assert.Nil(t, err)

	err = n.AddQueryHandler("handler", &mock.QueryHandlerStub{})

	assert.True(t, errors.Is(err, node.ErrQueryHandlerAlreadyExists))
}

func TestNode_GetQueryHandlerNotExistsShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	qh, err := n.GetQueryHandler("handler")

	assert.True(t, check.IfNil(qh))
	assert.True(t, errors.Is(err, node.ErrNilQueryHandler))
}

func TestNode_GetQueryHandlerShouldWork(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	qh := &mock.QueryHandlerStub{}
	handler := "handler"
	_ = n.AddQueryHandler(handler, &mock.QueryHandlerStub{})

	qhRecovered, err := n.GetQueryHandler(handler)

	assert.Equal(t, qhRecovered, qh)
	assert.Nil(t, err)
}

func TestNode_GetPeerInfoUnknownPeerShouldErr(t *testing.T) {
	t.Parallel()

	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = &mock.MessengerStub{
		PeersCalled: func() []core.PeerID {
			return make([]core.PeerID, 0)
		},
	}

	n, _ := node.NewNode(
		node.WithNetworkComponents(networkComponents),
	)

	pid := "pid"
	vals, err := n.GetPeerInfo(pid)

	assert.Nil(t, vals)
	assert.True(t, errors.Is(err, node.ErrUnknownPeerID))
}

func TestNode_ShouldWork(t *testing.T) {
	t.Parallel()

	pid1 := "pid1"
	pid2 := "pid2"
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = &mock.MessengerStub{
		PeersCalled: func() []core.PeerID {
			//return them unsorted
			return []core.PeerID{core.PeerID(pid2), core.PeerID(pid1)}
		},
		PeerAddressesCalled: func(pid core.PeerID) []string {
			return []string{"addr" + string(pid)}
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.ValPubKeyConv = mock.NewPubkeyConverterMock(32)

	n, _ := node.NewNode(
		node.WithNetworkComponents(networkComponents),
		node.WithNetworkShardingCollector(&mock.NetworkShardingCollectorStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				return core.P2PPeerInfo{
					PeerType: 0,
					ShardID:  0,
					PkBytes:  pid.Bytes(),
				}
			},
		}),
		node.WithCoreComponents(coreComponents),
		node.WithPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{
			IsDeniedCalled: func(pid core.PeerID) bool {
				return pid == core.PeerID(pid1)
			},
		}),
	)

	vals, err := n.GetPeerInfo("3sf1k") //will return both pids, sorted

	assert.Nil(t, err)
	require.Equal(t, 2, len(vals))

	expected := []core.QueryP2PPeerInfo{
		{
			Pid:           core.PeerID(pid1).Pretty(),
			Addresses:     []string{"addr" + pid1},
			Pk:            hex.EncodeToString([]byte(pid1)),
			IsBlacklisted: true,
			PeerType:      core.UnknownPeer.String(),
		},
		{
			Pid:           core.PeerID(pid2).Pretty(),
			Addresses:     []string{"addr" + pid2},
			Pk:            hex.EncodeToString([]byte(pid2)),
			IsBlacklisted: false,
			PeerType:      core.UnknownPeer.String(),
		},
	}

	assert.Equal(t, expected, vals)
}
