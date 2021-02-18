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

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elrond-go/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/core"
	atomicCore "github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go/core/versioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
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
	_, err := node.NewNode(node.WithAccountsAdapter(nil))
	assert.NotNil(t, err)
}

func TestNewNode_ApplyNilOptionShouldError(t *testing.T) {
	n, _ := node.NewNode()
	err := n.ApplyOptions(node.WithAccountsAdapter(nil))
	assert.NotNil(t, err)
}

func TestGetBalance_NoAddrConverterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and PubkeyConverter first", err.Error())
}

func TestGetBalance_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
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
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
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
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
	)
	balance, err := n.GetBalance(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestGetBalance(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(100))
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
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
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accDB),
	)
	username, err := n.GetUsername(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, string(expectedUsername), username)
}

func TestNode_GetKeyValuePairs(t *testing.T) {
	acc, _ := state.NewUserAccount([]byte("newaddress"))

	k1, v1 := []byte("key1"), []byte("value1")
	k2, v2 := []byte("key2"), []byte("value2")

	accDB := &mock.AccountsStub{}
	acc.DataTrieTracker().SetDataTrie(
		&mock.TrieStub{
			GetAllLeavesOnChannelCalled: func(rootHash []byte) (chan core.KeyValueHolder, error) {
				ch := make(chan core.KeyValueHolder)

				go func() {
					trieLeaf := keyValStorage.NewKeyValStorage(k1, v1)
					ch <- trieLeaf

					trieLeaf2 := keyValStorage.NewKeyValStorage(k2, v2)
					ch <- trieLeaf2
					close(ch)
				}()

				return ch, nil
			},
		})

	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acc, nil
	}

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accDB),
	)

	pairs, err := n.GetKeyValuePairs(createDummyHexAddress(64))
	assert.Nil(t, err)
	resV1, ok := pairs[hex.EncodeToString(k1)]
	assert.True(t, ok)
	assert.Equal(t, hex.EncodeToString(v1), resV1)

	resV2, ok := pairs[hex.EncodeToString(k2)]
	assert.True(t, ok)
	assert.Equal(t, hex.EncodeToString(v2), resV2)
}

func TestNode_GetValueForKey(t *testing.T) {
	acc, _ := state.NewUserAccount([]byte("newaddress"))

	k1, v1 := []byte("key1"), []byte("value1")
	_ = acc.DataTrieTracker().SaveKeyValue(k1, v1)

	accDB := &mock.AccountsStub{}

	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acc, nil
	}

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accDB),
	)

	value, err := n.GetValueForKey(createDummyHexAddress(64), hex.EncodeToString(k1))
	assert.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(v1), value)
}

func TestNode_GetESDTBalance(t *testing.T) {
	acc, _ := state.NewUserAccount([]byte("newaddress"))
	esdtToken := "newToken"
	esdtKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + esdtToken)

	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(10)}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	_ = acc.DataTrieTracker().SaveKeyValue(esdtKey, marshalledData)

	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acc, nil
	}
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accDB),
	)

	value, _, err := n.GetESDTBalance(createDummyHexAddress(64), esdtToken)
	assert.Nil(t, err)
	assert.Equal(t, esdtData.Value.String(), value)
}

func TestNode_GetAllESDTTokens(t *testing.T) {
	acc, _ := state.NewUserAccount([]byte("newaddress"))
	esdtToken := "newToken"
	esdtKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + esdtToken)

	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(10)}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	_ = acc.DataTrieTracker().SaveKeyValue(esdtKey, marshalledData)

	acc.DataTrieTracker().SetDataTrie(
		&mock.TrieStub{
			GetAllLeavesOnChannelCalled: func(rootHash []byte) (chan core.KeyValueHolder, error) {
				ch := make(chan core.KeyValueHolder)

				go func() {
					trieLeaf := keyValStorage.NewKeyValStorage(esdtKey, marshalledData)
					ch <- trieLeaf
					close(ch)
				}()

				return ch, nil
			},
		})

	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acc, nil
	}
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accDB),
	)

	value, err := n.GetAllESDTTokens(createDummyHexAddress(64))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, esdtToken, value[0])
}

//------- GenerateTransaction

func TestGenerateTransaction_NoAddrConverterShouldError(t *testing.T) {
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", &mock.PrivateKeyStub{}, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoPrivateKeyShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", nil, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_CreateAddressFailsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
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
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(&mock.SinglesignMock{}),
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

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
}

func TestGenerateTransaction_GetExistingAccountShouldWork(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
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
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(marshalizer, testSizeCheckDelta),
		node.WithVmMarshalizer(marshalizer),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_SignTxErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignFailMock{}

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_ShouldSetCorrectSignature(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	signature := []byte("signed")
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
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

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
}

func TestGenerateTransaction_CorrectParamsShouldNotError(t *testing.T) {

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithTxSingleSigner(singleSigner),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
}

func TestCreateTransaction_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	chainID := []byte("chain id")
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithChainID(chainID),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := ""
	sender := ""
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "-"

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, string(chainID), 1, 0)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.Equal(t, node.ErrNilPubkeyConverter, err)
}

func TestCreateTransaction_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	chainID := "chain id"
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
			},
		),
		node.WithChainID([]byte(chainID)),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := ""
	sender := ""
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := "-"

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, chainID, 1, 0)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.Equal(t, node.ErrNilAccountsAdapter, err)
}

func TestCreateTransaction_InvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
			},
		),
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

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, "chainID", 1, 0)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.NotNil(t, err)
}

func TestCreateTransaction_ChainIDFieldChecks(t *testing.T) {
	t.Parallel()

	chainID := "chain id"
	expectedHash := []byte("expected hash")
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
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithTxFeeHandler(&mock.FeeHandlerStub{}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithChainID([]byte(chainID)),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString([]byte(strings.Repeat("s", 10)))

	emptyChainID := ""
	_, _, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, emptyChainID, 1, 0)
	assert.Equal(t, node.ErrInvalidChainIDInTransaction, err)

	for i := 1; i < len(chainID); i++ {
		newChainID := strings.Repeat("c", i)
		_, _, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, newChainID, 1, 0)
		assert.NoError(t, err)
	}

	newChainID := chainID + "additional text"
	_, _, err = n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, newChainID, 1, 0)
	assert.Equal(t, node.ErrInvalidChainIDInTransaction, err)
}

func TestCreateTransaction_InvalidTxVersionShouldErr(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
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
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
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
	_, _, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, "", 0, 0)
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
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
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
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, string(chainID), version, 0)
	assert.NotNil(t, tx)
	assert.Equal(t, expectedHash, txHash)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
	assert.Equal(t, value, tx.Value)
	assert.True(t, bytes.Equal([]byte(receiver), tx.RcvAddr))

	err = n.ValidateTransaction(tx)
	assert.True(t, errors.Is(err, node.ErrDifferentSenderShardId))
}

func TestCreateTransaction_SignatureLengthChecks(t *testing.T) {
	t.Parallel()

	maxValueLength := 7
	signatureLength := 10
	chainID := "chain id"
	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxFeeHandler(
			&mock.FeeHandlerStub{
				GenesisTotalSupplyCalled: func() *big.Int {
					str := strings.Repeat("1", maxValueLength)
					bi := big.NewInt(0)
					bi.SetString(str, 10)
					return bi
				},
			}),
		node.WithChainID([]byte(chainID)),
		node.WithAddressSignatureSize(signatureLength),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithHasher(&mock.HasherMock{}),
	)

	nonce := uint64(0)
	value := "1" + strings.Repeat("0", maxValueLength)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")

	for i := 0; i <= signatureLength; i++ {
		signatureBytes := []byte(strings.Repeat("a", i))
		signatureHex := hex.EncodeToString(signatureBytes)
		tx, _, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signatureHex, chainID, 1, 0)
		assert.NotNil(t, tx)
		assert.NoError(t, err)
		assert.Equal(t, signatureBytes, tx.Signature)
	}

	signature := hex.EncodeToString([]byte(strings.Repeat("a", signatureLength+1)))
	tx, txHash, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signature, chainID, 1, 0)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Equal(t, node.ErrInvalidSignatureLength, err)
}

func TestCreateTransaction_SenderLengthChecks(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	encodedAddressLen := 5
	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return encodedAddressLen
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxFeeHandler(
			&mock.FeeHandlerStub{
				GenesisTotalSupplyCalled: func() *big.Int {
					str := strings.Repeat("1", maxLength)
					bi := big.NewInt(0)
					bi.SetString(str, 10)
					return bi
				},
			}),
		node.WithChainID([]byte(chainID)),
		node.WithAddressSignatureSize(10),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithHasher(&mock.HasherMock{}),
		node.WithTxFeeHandler(&mock.FeeHandlerStub{}),
	)

	nonce := uint64(0)
	value := "10"
	receiver := "rcv"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	for i := 0; i <= encodedAddressLen; i++ {
		sender := strings.Repeat("s", i)
		_, _, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signature, chainID, 1, 0)
		assert.NoError(t, err)
	}

	sender := strings.Repeat("s", encodedAddressLen) + "additional"
	tx, txHash, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signature, chainID, 1, 0)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, node.ErrInvalidAddressLength))
}

func TestCreateTransaction_ReceiverLengthChecks(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	encodedAddressLen := 5
	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return encodedAddressLen
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxFeeHandler(
			&mock.FeeHandlerStub{
				GenesisTotalSupplyCalled: func() *big.Int {
					str := strings.Repeat("1", maxLength)
					bi := big.NewInt(0)
					bi.SetString(str, 10)
					return bi
				},
			}),
		node.WithChainID([]byte(chainID)),
		node.WithAddressSignatureSize(10),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithHasher(&mock.HasherMock{}),
		node.WithTxFeeHandler(&mock.FeeHandlerStub{}),
	)

	nonce := uint64(0)
	value := "10"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	for i := 0; i <= encodedAddressLen; i++ {
		receiver := strings.Repeat("r", i)
		_, _, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signature, chainID, 1, 0)
		assert.NoError(t, err)
	}

	receiver := strings.Repeat("r", encodedAddressLen) + "additional"
	tx, txHash, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signature, chainID, 1, 0)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, node.ErrInvalidAddressLength))
}

func TestCreateTransaction_TooBigSenderUsernameShouldErr(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxFeeHandler(
			&mock.FeeHandlerStub{
				GenesisTotalSupplyCalled: func() *big.Int {
					str := strings.Repeat("1", maxLength)
					bi := big.NewInt(0)
					bi.SetString(str, 10)
					return bi
				},
			}),
		node.WithChainID([]byte(chainID)),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := "1" + strings.Repeat("0", maxLength+1)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	senderUsername := bytes.Repeat([]byte{0}, core.MaxUserNameLength+1)

	tx, txHash, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, senderUsername, gasPrice, gasLimit, txData, signature, chainID, 1, 0)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.Equal(t, node.ErrInvalidSenderUsernameLength, err)
}

func TestCreateTransaction_TooBigReceiverUsernameShouldErr(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxFeeHandler(
			&mock.FeeHandlerStub{
				GenesisTotalSupplyCalled: func() *big.Int {
					str := strings.Repeat("1", maxLength)
					bi := big.NewInt(0)
					bi.SetString(str, 10)
					return bi
				},
			}),
		node.WithChainID([]byte(chainID)),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := "1" + strings.Repeat("0", maxLength+1)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	receiverUsername := bytes.Repeat([]byte{0}, core.MaxUserNameLength+1)

	tx, txHash, err := n.CreateTransaction(nonce, value, receiver, receiverUsername, sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signature, chainID, 1, 0)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.Equal(t, node.ErrInvalidReceiverUsernameLength, err)
}

func TestCreateTransaction_DataFieldSizeExceedsMaxShouldErr(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxFeeHandler(
			&mock.FeeHandlerStub{
				GenesisTotalSupplyCalled: func() *big.Int {
					str := strings.Repeat("1", maxLength)
					bi := big.NewInt(0)
					bi.SetString(str, 10)
					return bi
				},
			}),
		node.WithChainID([]byte(chainID)),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := "1" + strings.Repeat("0", maxLength+1)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := bytes.Repeat([]byte{0}, core.MegabyteSize+1)
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	tx, txHash, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signature, chainID, 1, 0)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.Equal(t, node.ErrDataFieldTooBig, err)
}

func TestCreateTransaction_TooLargeValueFieldShouldErr(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
					return []byte(hexAddress), nil
				},
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithTxFeeHandler(
			&mock.FeeHandlerStub{
				GenesisTotalSupplyCalled: func() *big.Int {
					str := strings.Repeat("1", maxLength)
					bi := big.NewInt(0)
					bi.SetString(str, 10)
					return bi
				},
			}),
		node.WithChainID([]byte(chainID)),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := "1" + strings.Repeat("0", maxLength+1)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	tx, txHash, err := n.CreateTransaction(nonce, value, receiver, []byte("rcvrUsername"), sender, []byte("sndrUsername"), gasPrice, gasLimit, txData, signature, chainID, 1, 0)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.Equal(t, node.ErrTransactionValueLengthTooBig, err)
}

func TestCreateTransaction_OkValsShouldWork(t *testing.T) {
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
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{
			ComputeIdCalled: func(i []byte) uint32 {
				return crtShardID
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
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return 1
			},
		}),
		node.WithTxSignHasher(&mock.HasherMock{}),
		node.WithTxVersionChecker(versioning.NewTxVersionChecker(version)),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	tx, txHash, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, string(chainID), version, 0)
	assert.NotNil(t, tx)
	assert.Equal(t, expectedHash, txHash)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
	assert.Equal(t, value, tx.Value)
	assert.True(t, bytes.Equal([]byte(receiver), tx.RcvAddr))

	err = n.ValidateTransaction(tx)
	assert.Nil(t, err)
}

func TestCreateTransaction_TxSignedWithHashShouldErrVersionShoudBe2(t *testing.T) {
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
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{
			ComputeIdCalled: func(i []byte) uint32 {
				return crtShardID
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
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return 1
			},
		}),
		node.WithEnableSignTxWithHashEpoch(2),
		node.WithTxSignHasher(&mock.HasherMock{}),
		node.WithTxVersionChecker(versioning.NewTxVersionChecker(version)),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	options := versioning.MaskSignedWithHash
	tx, _, err := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, string(chainID), version, options)
	require.Nil(t, err)
	err = n.ValidateTransaction(tx)
	assert.Equal(t, process.ErrInvalidTransactionVersion, err)
}

func TestCreateTransaction_TxSignedWithHashNoEnabledShouldErr(t *testing.T) {
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
				EncodeCalled: func(pkBytes []byte) string {
					return string(pkBytes)
				},
				LenCalled: func() int {
					return 3
				},
			}),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{
			ComputeIdCalled: func(i []byte) uint32 {
				return crtShardID
			},
			SelfShardId: crtShardID,
		}),
		node.WithWhiteListHandler(&mock.WhiteListHandlerStub{}),
		node.WithWhiteListHandlerVerified(&mock.WhiteListHandlerStub{
			IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
				return false
			},
		}),
		node.WithKeyGenForAccounts(&mock.KeyGenMock{
			PublicKeyFromByteArrayMock: func(b []byte) (crypto.PublicKey, error) {
				return nil, nil
			},
		}),
		node.WithTxSingleSigner(&mock.SingleSignerMock{}),
		node.WithTxFeeHandler(&mock.FeeHandlerStub{
			CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
				return nil
			},
		}),
		node.WithChainID(chainID),
		node.WithMinTransactionVersion(version),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return 1
			},
		}),
		node.WithEnableSignTxWithHashEpoch(2),
		node.WithTxSignHasher(&mock.HasherMock{}),
		node.WithTxVersionChecker(versioning.NewTxVersionChecker(version)),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"
	sender := "snd"
	gasPrice := uint64(10)
	gasLimit := uint64(20)
	txData := []byte("-")
	signature := hex.EncodeToString(bytes.Repeat([]byte{0}, 10))

	options := versioning.MaskSignedWithHash
	tx, _, _ := n.CreateTransaction(nonce, value.String(), receiver, nil, sender, nil, gasPrice, gasLimit, txData, signature, string(chainID), version+1, options)

	err := n.ValidateTransaction(tx)
	assert.Equal(t, process.ErrTransactionSignedWithHashIsNotEnabled, err)
}

func TestSendBulkTransactions_NoTxShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}
	marshalizer := &mock.MarshalizerFake{}
	hasher := &mock.HasherFake{}
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(marshalizer, testSizeCheckDelta),
		node.WithVmMarshalizer(marshalizer),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithMessenger(mes),
		node.WithHasher(hasher),
	)
	txs := make([]*transaction.Transaction, 0)

	numOfTxsProcessed, err := n.SendBulkTransactions(txs)
	assert.Equal(t, uint64(0), numOfTxsProcessed)
	assert.Equal(t, node.ErrNoTxToProcess, err)
}

func TestCreateShardedStores_NilShardCoordinatorShouldError(t *testing.T) {
	messenger := getMessenger()
	dataPool := testscommon.NewPoolsHolderStub()

	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)

	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil shard coordinator")
}

func TestCreateShardedStores_NilDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)

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
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
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
	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
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

	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithShardCoordinator(shardCoordinator),
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(getMarshalizer(), testSizeCheckDelta),
		node.WithVmMarshalizer(getMarshalizer()),
		node.WithTxSignMarshalizer(getMarshalizer()),
		node.WithHasher(getHasher()),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(&mock.AccountsStub{}),
	)

	err := n.CreateShardedStores()
	assert.Nil(t, err)
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
			HasTopicCalled: func(name string) bool {
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

	n, _ := node.NewNode(
		node.WithInitialNodesPubKeys(initialPubKeys),
		node.WithValidatorStatistics(vsp),
		node.WithValidatorsProvider(validatorProvider),
	)

	expectedData := &state.ValidatorApiResponse{}
	validatorsData, err := n.ValidatorStatisticsApi()
	require.Equal(t, expectedData, validatorsData[hex.EncodeToString([]byte(keys[2][0]))])
	require.Nil(t, err)
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
		node.WithIndexer(&mock.IndexerMock{}),
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
		node.WithDataPool(&testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

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
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
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
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithIndexer(&mock.IndexerMock{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, state.ErrNilAccountsAdapter, err)
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
			return &mock.StorerStub{}
		},
	}

	tr := &mock.TrieStub{
		GetStorageManagerCalled: func() data.StorageManager {
			return &mock.StorageManagerStub{
				DatabaseCalled: func() data.DBWriteCacher {
					return &mock.StorerMock{}
				},
			}
		},
	}
	accountDb, _ := state.NewAccountsDB(tr, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithBootStorer(&mock.BoostrapStorerMock{}),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithTxSignMarshalizer(&mock.MarshalizerMock{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithIndexer(&mock.IndexerMock{}),
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
		node.WithSyncer(&mock.SyncTimerStub{}),
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
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithTxSignMarshalizer(&mock.MarshalizerMock{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithPendingMiniBlocksHandler(&mock.PendingMiniBlocksHandlerStub{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithIndexer(&mock.IndexerMock{}),
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
	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(shardingCoordinator),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithDataPool(testscommon.NewPoolsHolderStub()),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithIndexer(&mock.IndexerMock{}),
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

	tr := &mock.TrieStub{
		GetStorageManagerCalled: func() data.StorageManager {
			return &mock.StorageManagerStub{
				DatabaseCalled: func() data.DBWriteCacher {
					return &mock.StorerMock{}
				},
			}
		},
	}
	accountDb, _ := state.NewAccountsDB(tr, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	localErr := errors.New("err")
	n, _ := node.NewNode(
		node.WithDataPool(&testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

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
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockBlackListHandler(&mock.TimeCacheStub{}),
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
		node.WithRequestedItemsHandler(&mock.TimeCacheStub{}),
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
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithIndexer(elasticIndexer.NewNilIndexer()),
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

	tr := &mock.TrieStub{
		GetStorageManagerCalled: func() data.StorageManager {
			return &mock.StorageManagerStub{
				DatabaseCalled: func() data.DBWriteCacher {
					return &mock.StorerMock{}
				},
			}
		},
	}
	accountDb, _ := state.NewAccountsDB(tr, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithDataPool(&testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

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
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockBlackListHandler(&mock.TimeCacheStub{}),
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
		node.WithRequestedItemsHandler(&mock.TimeCacheStub{}),
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
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithIndexer(elasticIndexer.NewNilIndexer()),
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

	tr := &mock.TrieStub{
		GetStorageManagerCalled: func() data.StorageManager {
			return &mock.StorageManagerStub{
				DatabaseCalled: func() data.DBWriteCacher {
					return &mock.StorerMock{}
				},
			}
		},
	}
	accountDb, _ := state.NewAccountsDB(tr, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithDataPool(&testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

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
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{
			CheckForkCalled: func() *process.ForkInfo {
				return &process.ForkInfo{}
			},
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
		}),
		node.WithBlockBlackListHandler(&mock.TimeCacheStub{}),
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
		node.WithRequestedItemsHandler(&mock.TimeCacheStub{}),
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
		node.WithEpochStartEventNotifier(&mock.EpochStartNotifierStub{}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithNetworkShardingCollector(&mock.NetworkShardingCollectorStub{}),
		node.WithInputAntifloodHandler(&mock.P2PAntifloodHandlerStub{}),
		node.WithHeaderIntegrityVerifier(&mock.HeaderIntegrityVerifierStub{}),
		node.WithPeerHonestyHandler(&testscommon.PeerHonestyHandlerStub{}),
		node.WithFallbackHeaderValidator(&testscommon.FallBackHeaderValidatorStub{}),
		node.WithHardforkTrigger(&mock.HardforkTriggerStub{}),
		node.WithInterceptorsContainer(&mock.InterceptorsContainerStub{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithPeerSignatureHandler(&mock.PeerSignatureHandler{}),
		node.WithIndexer(elasticIndexer.NewNilIndexer()),
	)

	err := n.StartConsensus()
	assert.Nil(t, err)
}

//------- GetAccount

func TestNode_GetAccountWithNilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
	)

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

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
	)

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
	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressPubkeyConverter(
			&mock.PubkeyConverterStub{
				DecodeCalled: func(hexAddress string) ([]byte, error) {
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
		GetExistingAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
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

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
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

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
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

func TestNode_EncodeDecodeAddressPubkey(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")
	n, _ := node.NewNode(
		node.WithAddressPubkeyConverter(mock.NewPubkeyConverterMock(32)),
	)
	encoded, err := n.EncodeAddressPubkey(buff)
	assert.Nil(t, err)

	recoveredBytes, err := n.DecodeAddressPubkey(encoded)

	assert.Nil(t, err)
	assert.Equal(t, buff, recoveredBytes)
}

func TestNode_EncodeDecodeAddressPubkeyWithNilCoberterShouldErr(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")
	n, _ := node.NewNode()
	encoded, err := n.EncodeAddressPubkey(buff)

	assert.Empty(t, encoded)
	assert.True(t, errors.Is(err, node.ErrNilPubkeyConverter))
}

func TestNode_DecodeAddressPubkeyWithNilConverterShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

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
	feeHandler := &mock.FeeHandlerStub{
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
	n, _ := node.NewNode(
		node.WithInternalMarshalizer(marshalizer, testSizeCheckDelta),
		node.WithVmMarshalizer(marshalizer),
		node.WithTxSignMarshalizer(marshalizer),
		node.WithHasher(&mock.HasherMock{}),
		node.WithAddressPubkeyConverter(createMockPubkeyConverter()),
		node.WithAccountsAdapter(accAdapter),
		node.WithKeyGenForAccounts(keyGen),
		node.WithTxSingleSigner(signer),
		node.WithShardCoordinator(shardCoordinator),
		node.WithMessenger(mes),
		node.WithDataPool(dataPool),
		node.WithTxFeeHandler(feeHandler),
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

	n, _ := node.NewNode(
		node.WithMessenger(&mock.MessengerStub{
			PeersCalled: func() []core.PeerID {
				return make([]core.PeerID, 0)
			},
		}),
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
	n, _ := node.NewNode(
		node.WithMessenger(&mock.MessengerStub{
			PeersCalled: func() []core.PeerID {
				//return them unsorted
				return []core.PeerID{core.PeerID(pid2), core.PeerID(pid1)}
			},
			PeerAddressesCalled: func(pid core.PeerID) []string {
				return []string{"addr" + string(pid)}
			},
		}),
		node.WithNetworkShardingCollector(&mock.NetworkShardingCollectorStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				return core.P2PPeerInfo{
					PeerType: 0,
					ShardID:  0,
					PkBytes:  pid.Bytes(),
				}
			},
		}),
		node.WithValidatorPubkeyConverter(mock.NewPubkeyConverterMock(32)),
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

func TestNode_ValidateTransactionForSimulation_BypassSignatureTrue(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithAccountsAdapter(&mock.AccountsStub{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithWhiteListHandler(&mock.WhiteListHandlerStub{}),
		node.WithWhiteListHandlerVerified(&mock.WhiteListHandlerStub{}),
		node.WithAddressPubkeyConverter(mock.NewPubkeyConverterMock(3)),
		node.WithTxSignHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 10),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithTxSignMarshalizer(&mock.MarshalizerFake{}),
		node.WithHasher(&mock.HasherMock{}),
		node.WithKeyGenForAccounts(&mock.KeyGenMock{}),
		node.WithTxFeeHandler(&mock.FeeHandlerStub{}),
		node.WithChainID([]byte("a")),
		node.WithTxVersionChecker(versioning.NewTxVersionChecker(0)),
	)

	tx := &transaction.Transaction{
		Nonce:     11,
		Value:     big.NewInt(25),
		RcvAddr:   []byte("rec"),
		SndAddr:   []byte("snd"),
		GasPrice:  6,
		GasLimit:  12,
		Data:      []byte(""),
		Signature: []byte("sig1"),
		ChainID:   []byte("a"),
	}

	err := n.ValidateTransactionForSimulation(tx, true)
	require.NoError(t, err)
}
