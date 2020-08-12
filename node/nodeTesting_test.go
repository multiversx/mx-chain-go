package node_test

import (
	"errors"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSizeCheckDelta = 100

var timeoutWait = time.Second

//------- GenerateAndSendBulkTransactions

func TestGenerateAndSendBulkTransactions_ZeroTxShouldErr(t *testing.T) {
	n, _ := node.NewNode()

	err := n.GenerateAndSendBulkTransactions("", big.NewInt(0), 0, &mock.PrivateKeyStub{}, nil, []byte("chainID"), 1)
	assert.NotNil(t, err)
	assert.Equal(t, "can not generate and broadcast 0 transactions", err.Error())
}

func TestGenerateAndSendBulkTransactions_NilAccountAdapterShouldErr(t *testing.T) {
	marshalizer := &mock.MarshalizerFake{}

	keyGen := &mock.KeyGenMock{}
	sk, _ := keyGen.GeneratePair()
	singleSigner := &mock.SinglesignMock{}

	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.IntMarsh = marshalizer
	coreComponentsMock.AddrPubKeyConv = createMockPubkeyConverter()
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.ShardCoord = mock.NewOneShardCoordinatorMock()
	cryptoComponentsMock := getDefaultCryptoComponents()
	cryptoComponentsMock.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithCryptoComponents(cryptoComponentsMock),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(0), 1, sk, nil, []byte("chainID"), 1)
	assert.Equal(t, node.ErrNilAccountsAdapter, err)
}

func TestGenerateAndSendBulkTransactions_NilSingleSignerShouldErr(t *testing.T) {
	marshalizer := &mock.MarshalizerFake{}

	keyGen := &mock.KeyGenMock{}
	sk, _ := keyGen.GeneratePair()
	accAdapter := getAccAdapter(big.NewInt(0))
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.IntMarsh = marshalizer
	coreComponentsMock.AddrPubKeyConv = createMockPubkeyConverter()
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.ShardCoord = mock.NewOneShardCoordinatorMock()
	stateComponentsMock := getDefaultStateComponents()
	stateComponentsMock.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithStateComponents(stateComponentsMock),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(0), 1, sk, nil, []byte("chainID"), 1)
	assert.Equal(t, node.ErrNilSingleSig, err)
}

func TestGenerateAndSendBulkTransactions_NilShardCoordinatorShouldErr(t *testing.T) {
	marshalizer := &mock.MarshalizerFake{}

	keyGen := &mock.KeyGenMock{}
	sk, _ := keyGen.GeneratePair()
	accAdapter := getAccAdapter(big.NewInt(0))
	singleSigner := &mock.SinglesignMock{}
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.IntMarsh = marshalizer
	coreComponentsMock.AddrPubKeyConv = createMockPubkeyConverter()
	cryptoComponentsMock := getDefaultCryptoComponents()
	cryptoComponentsMock.TxSig = singleSigner
	stateComponentsMock := getDefaultStateComponents()
	stateComponentsMock.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithCryptoComponents(cryptoComponentsMock),
		node.WithStateComponents(stateComponentsMock),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(0), 1, sk, nil, []byte("chainID"), 1)
	assert.Equal(t, node.ErrNilShardCoordinator, err)
}

func TestGenerateAndSendBulkTransactions_NilPubkeyConverterShouldErr(t *testing.T) {
	marshalizer := &mock.MarshalizerFake{}
	accAdapter := getAccAdapter(big.NewInt(0))
	keyGen := &mock.KeyGenMock{}
	sk, _ := keyGen.GeneratePair()
	singleSigner := &mock.SinglesignMock{}
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.IntMarsh = marshalizer
	cryptoComponentsMock := getDefaultCryptoComponents()
	cryptoComponentsMock.TxSig = singleSigner
	stateComponentsMock := getDefaultStateComponents()
	stateComponentsMock.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithCryptoComponents(cryptoComponentsMock),
		node.WithStateComponents(stateComponentsMock),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(0), 1, sk, nil, []byte("chainID"), 1)
	assert.Equal(t, node.ErrNilPubkeyConverter, err)
}

func TestGenerateAndSendBulkTransactions_NilPrivateKeyShouldErr(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))
	singleSigner := &mock.SinglesignMock{}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return nil
				},
			}
		},
	}
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.IntMarsh = &mock.MarshalizerFake{}
	coreComponentsMock.AddrPubKeyConv = createMockPubkeyConverter()
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.ShardCoord = mock.NewOneShardCoordinatorMock()
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.DataPool = dataPool
	cryptoComponentsMock := getDefaultCryptoComponents()
	cryptoComponentsMock.TxSig = singleSigner
	stateComponentsMock := getDefaultStateComponents()
	stateComponentsMock.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
		node.WithCryptoComponents(cryptoComponentsMock),
		node.WithStateComponents(stateComponentsMock),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(0), 1, nil, nil, []byte("chainID"), 1)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "trying to set nil private key"))
}

func TestGenerateAndSendBulkTransactions_InvalidReceiverAddressShouldErr(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))

	sk := &mock.PrivateKeyStub{GeneratePublicHandler: func() crypto.PublicKey {
		return &mock.PublicKeyMock{
			ToByteArrayHandler: func() (bytes []byte, err error) {
				return []byte("key"), nil
			},
		}
	}}
	singleSigner := &mock.SinglesignMock{}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return nil
				},
			}
		},
	}
	expectedErr := errors.New("expected error")
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.AddrPubKeyConv = &mock.PubkeyConverterStub{
		DecodeCalled: func(humanReadable string) ([]byte, error) {
			if len(humanReadable) == 0 {
				return nil, expectedErr
			}

			return []byte("1234"), nil
		},
	}
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.ShardCoord = mock.NewOneShardCoordinatorMock()
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.DataPool = dataPool
	cryptoComponentsMock := getDefaultCryptoComponents()
	cryptoComponentsMock.TxSig = singleSigner
	stateComponentsMock := getDefaultStateComponents()
	stateComponentsMock.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
		node.WithCryptoComponents(cryptoComponentsMock),
		node.WithStateComponents(stateComponentsMock),
	)

	err := n.GenerateAndSendBulkTransactions("", big.NewInt(0), 1, sk, nil, []byte("chainID"), 1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "could not create receiver address from provided param")
}

func TestGenerateAndSendBulkTransactions_MarshalizerErrorsShouldErr(t *testing.T) {
	accAdapter := getAccAdapter(big.NewInt(0))
	marshalizer := &mock.MarshalizerFake{}
	marshalizer.Fail = true
	sk := &mock.PrivateKeyStub{GeneratePublicHandler: func() crypto.PublicKey {
		return &mock.PublicKeyMock{
			ToByteArrayHandler: func() (bytes []byte, err error) {
				return []byte("key"), nil
			},
		}
	}}
	singleSigner := &mock.SinglesignMock{}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return nil
				},
			}
		},
	}

	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.IntMarsh = marshalizer
	coreComponentsMock.AddrPubKeyConv = createMockPubkeyConverter()
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.ShardCoord = mock.NewOneShardCoordinatorMock()
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.DataPool = dataPool
	cryptoComponentsMock := getDefaultCryptoComponents()
	cryptoComponentsMock.TxSig = singleSigner
	stateComponentsMock := getDefaultStateComponents()
	stateComponentsMock.Accounts = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
		node.WithCryptoComponents(cryptoComponentsMock),
		node.WithStateComponents(stateComponentsMock),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(1), 1, sk, nil, []byte("chainID"), 1)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "could not marshal transaction"))
}

func TestGenerateAndSendBulkTransactions_ShouldWork(t *testing.T) {
	marshalizer := &mock.MarshalizerFake{}

	noOfTx := 1000
	mutRecoveredTransactions := &sync.RWMutex{}
	recoveredTransactions := make(map[uint64]*transaction.Transaction)
	signer := &mock.SinglesignMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	wg := sync.WaitGroup{}
	wg.Add(noOfTx)

	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		chDone <- struct{}{}
	}()

	mes := &mock.MessengerStub{
		BroadcastOnChannelBlockingCalled: func(pipe string, topic string, buff []byte) error {
			identifier := factory.TransactionTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())

			if topic == identifier {
				//handler to capture sent data
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
					recoveredTransactions[tx.Nonce] = &tx
					mutRecoveredTransactions.Unlock()

					wg.Done()
				}
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
	accAdapter := getAccAdapter(big.NewInt(0))
	sk := &mock.PrivateKeyStub{GeneratePublicHandler: func() crypto.PublicKey {
		return &mock.PublicKeyMock{
			ToByteArrayHandler: func() (bytes []byte, err error) {
				return []byte("key"), nil
			},
		}
	}}
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.IntMarsh = marshalizer
	coreComponentsMock.TxMarsh = marshalizer
	coreComponentsMock.AddrPubKeyConv = createMockPubkeyConverter()
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.ShardCoord = shardCoordinator
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.DataPool = dataPool
	cryptoComponentsMock := getDefaultCryptoComponents()
	cryptoComponentsMock.TxSig = signer
	stateComponentsMock := getDefaultStateComponents()
	stateComponentsMock.Accounts = accAdapter
	networkComponentsMock := getDefaultNetworkComponents()
	networkComponentsMock.Messenger = mes

	n, _ := node.NewNode(
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
		node.WithCryptoComponents(cryptoComponentsMock),
		node.WithStateComponents(stateComponentsMock),
		node.WithNetworkComponents(networkComponentsMock),
	)

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(1), uint64(noOfTx), sk, nil, []byte("chainID"), 1)
	assert.Nil(t, err)

	select {
	case <-chDone:
	case <-time.After(timeoutWait):
		assert.Fail(t, "timout while waiting the broadcast of the generated transactions")
		return
	}

	mutRecoveredTransactions.RLock()
	assert.Equal(t, noOfTx, len(recoveredTransactions))
	mutRecoveredTransactions.RUnlock()
}

func getDefaultCryptoComponents() *mock.CryptoComponentsMock {
	return &mock.CryptoComponentsMock{
		PubKey:          &mock.PublicKeyMock{},
		PrivKey:         &mock.PrivateKeyStub{},
		PubKeyString:    "pubKey",
		PrivKeyBytes:    []byte("privKey"),
		PubKeyBytes:     []byte("pubKey"),
		BlockSig:        &mock.SingleSignerMock{},
		TxSig:           &mock.SingleSignerMock{},
		MultiSig:        &mock.MultisignMock{},
		PeerSignHandler: &mock.PeerSignatureHandler{},
		BlKeyGen:        &mock.KeyGenMock{},
		TxKeyGen:        &mock.KeyGenMock{},
		MsgSigVerifier:  &testscommon.MessageSignVerifierMock{},
	}
}

func getDefaultStateComponents() *testscommon.StateComponentsMock {
	return &testscommon.StateComponentsMock{
		PeersAcc:        &mock.AccountsStub{},
		Accounts:        &mock.AccountsStub{},
		Tries:           &mock.TriesHolderStub{},
		StorageManagers: map[string]data.StorageManager{"0": &mock.StorageManagerStub{}},
	}
}

func getDefaultNetworkComponents() *mock.NetworkComponentsMock {
	return &mock.NetworkComponentsMock{
		Messenger:       &mock.MessengerStub{},
		InputAntiFlood:  &mock.P2PAntifloodHandlerStub{},
		OutputAntiFlood: &mock.P2PAntifloodHandlerStub{},
		PeerBlackList:   &mock.PeerBlackListHandlerStub{},
	}
}
