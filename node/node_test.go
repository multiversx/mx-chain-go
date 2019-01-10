package node_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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

func TestStart_NoPort(t *testing.T) {

	n, _ := node.NewNode()
	err := n.Start()
	defer func() { _ = n.Stop() }()
	assert.NotNil(t, err)
}

func TestStart_NoMarshalizer(t *testing.T) {

	n, _ := node.NewNode(node.WithPort(4000))
	err := n.Start()
	defer func() { _ = n.Stop() }()
	assert.NotNil(t, err)
}

func TestStart_NoHasher(t *testing.T) {

	n, _ := node.NewNode(node.WithPort(4000), node.WithMarshalizer(mock.MarshalizerMock{}))
	err := n.Start()
	defer func() { _ = n.Stop() }()
	assert.NotNil(t, err)
}

func TestStart_NoMaxAllowedPeers(t *testing.T) {

	n, _ := node.NewNode(node.WithPort(4000), node.WithMarshalizer(mock.MarshalizerMock{}), node.WithHasher(mock.HasherMock{}))
	err := n.Start()
	defer func() { _ = n.Stop() }()
	assert.NotNil(t, err)
}

func TestStart_CorrectParams(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
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
	err := n.ApplyOptions(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
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

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n.Start()
	defer func() { _ = n.Stop() }()
	logError(err)

	err = n.ApplyOptions(
		node.WithMaxAllowedPeers(4),
	)

	assert.NotNil(t, err)
	assert.True(t, n.IsRunning())
}

func TestStop_NotStartedYet(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
	)
	err := n.Start()
	defer func() { _ = n.Stop() }()
	logError(err)
	err = n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestStop(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
	)

	err := n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestConnectToAddresses_NodeNotStarted(t *testing.T) {

	n2, _ := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	defer func() { _ = n2.Stop() }()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
	)

	err = n.ConnectToAddresses([]string{addr})
	assert.NotNil(t, err)
}

func TestConnectToAddresses(t *testing.T) {

	n2, _ := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	defer func() { _ = n2.Stop() }()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
	)
	err = n.Start()
	defer func() { _ = n.Stop() }()
	assert.Nil(t, err)

	err = n.ConnectToAddresses([]string{addr})
	assert.Nil(t, err)
}

func TestAddress_NodeNotStarted(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
	)
	_, err := n.Address()
	assert.NotNil(t, err)
}

func TestGetBalance_NoAddrConverterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_CreateAddressFailsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(*big.NewInt(0))
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
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
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
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "could not fetch sender address from provided param", err.Error())
}

func TestGetBalance_GetAccountReturnsNil(t *testing.T) {

	accAdapter := mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, nil
		},
	}
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	balance, err := n.GetBalance("address")
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestGetBalance(t *testing.T) {

	accAdapter := getAccAdapter(*big.NewInt(100))
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	balance, err := n.GetBalance("address")
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(100), balance)
}

func TestGenerateTransaction_NoAddrConverterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoAccAdapterShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoPrivateKeyShouldError(t *testing.T) {

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(&mock.AddressConverterStub{}),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_CreateAddressFailsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(*big.NewInt(0))
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
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountFailsShouldError(t *testing.T) {

	accAdapter := mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, errors.New("error")
		},
	}
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountReturnsNilShouldWork(t *testing.T) {

	accAdapter := mock.AccountsAdapterStub{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, nil
		},
	}
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.Nil(t, err)
}

func TestGenerateTransaction_GetExistingAccountShouldWork(t *testing.T) {

	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.Nil(t, err)
}

func TestGenerateTransaction_MarshalErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	marshalizer := mock.MarshalizerMock{
		MarshalHandler: func(obj interface{}) ([]byte, error) {
			return nil, errors.New("error")
		},
	}
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_SignTxErrorsShouldError(t *testing.T) {

	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	privateKey := mock.PrivateKeyStub{
		SignHandler: func(message []byte) ([]byte, error) {
			return nil, errors.New("error")
		},
	}
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_ShouldSetCorrectSignature(t *testing.T) {

	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	signature := []byte{69}
	privateKey := mock.PrivateKeyStub{
		SignHandler: func(message []byte) ([]byte, error) {
			return signature, nil
		},
	}

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)

	tx, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
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
						Balance: *big.NewInt(0),
					}
				},
			}, nil
		},
	}
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)

	tx, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
}

func TestGenerateTransaction_CorrectParamsShouldNotError(t *testing.T) {

	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.MarshalizerMock{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accAdapter),
		node.WithPrivateKey(privateKey),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.Nil(t, err)
}

func getAccAdapter(balance big.Int) mock.AccountsAdapterStub {
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
		SignHandler: func(message []byte) ([]byte, error) {
			return []byte{2}, nil
		},
	}
}

func getAddressConverter() mock.AddressConverterStub {
	return mock.AddressConverterStub{
		CreateAddressFromHexHandler: func(hexAddress string) (state.AddressContainer, error) {
			// Return that will result in a correct run of GenerateTransaction -> will fail test
			return mock.AddressContainerStub{}, nil
		},
	}
}

func TestBindInterceptorsResolvers_NodeNotStartedShouldErr(t *testing.T) {
	n, _ := node.NewNode()

	err := n.BindInterceptorsResolvers()

	assert.Equal(t, "node is not started yet", err.Error())
}

func TestBindInterceptorsResolvers_ShouldWork(t *testing.T) {
	n, _ := node.NewNode(
		node.WithDataPool(createDataPoolMock()),
		node.WithAddressConverter(mock.AddressConverterStub{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithSingleSignKeyGenerator(&mock.SingleSignKeyGenMock{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithMarshalizer(&mock.MarshalizerMock{}),
		node.WithBlockChain(createStubBlockchain()),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
	)

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	prepareMessenger(mes)

	err := n.BindInterceptorsResolvers()

	assert.Nil(t, err)
}

func createDataPoolMock() *mock.TransientDataPoolMock {
	dataPool := &mock.TransientDataPoolMock{}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	return dataPool
}

func prepareMessenger(mes *mock.MessengerStub) {
	registration := func(v pubsub.Validator) error {
		return nil
	}

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTx.RegisterTopicValidator = registration
	topicHdr := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicHdr.RegisterTopicValidator = registration
	topicTxBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTxBlk.RegisterTopicValidator = registration
	topicPeerBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicPeerBlk.RegisterTopicValidator = registration
	topicStateBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicStateBlk.RegisterTopicValidator = registration

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(node.TransactionTopic):
			return topicTx
		case string(node.HeadersTopic):
			return topicHdr
		case string(node.TxBlockBodyTopic):
			return topicTxBlk
		case string(node.PeerChBodyTopic):
			return topicPeerBlk
		case string(node.StateBodyTopic):
			return topicStateBlk
		}

		return nil
	}
}

func createStubBlockchain() *blockchain.BlockChain {
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{})

	return blkc
}

func TestBindInterceptorsResolvers_CreateInterceptorFailsShouldErr(t *testing.T) {
	n, _ := node.NewNode(
		node.WithDataPool(createDataPoolMock()),
		node.WithHasher(mock.HasherMock{}),
		node.WithSingleSignKeyGenerator(&mock.SingleSignKeyGenMock{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithMarshalizer(&mock.MarshalizerMock{}),
		node.WithBlockChain(createStubBlockchain()),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
	)

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	prepareMessenger(mes)

	err := n.BindInterceptorsResolvers()

	assert.Equal(t, "nil AddressConverter", err.Error())
}

func TestBindInterceptorsResolvers_CreateResolversFailsShouldErr(t *testing.T) {
	n, _ := node.NewNode(
		node.WithDataPool(createDataPoolMock()),
		node.WithAddressConverter(mock.AddressConverterStub{}),
		node.WithHasher(mock.HasherMock{}),
		node.WithSingleSignKeyGenerator(&mock.SingleSignKeyGenMock{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithMarshalizer(&mock.MarshalizerMock{}),
		node.WithBlockChain(createStubBlockchain()),
	)

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	prepareMessenger(mes)

	err := n.BindInterceptorsResolvers()

	assert.Equal(t, "nil nonce converter", err.Error())
}

func TestSendTransaction_TopicDoesNotExistsShouldErr(t *testing.T) {
	n, _ := node.NewNode()

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	nonce := uint64(50)
	value := *big.NewInt(567)
	sender := "sender"
	receiver := "receiver"
	txData := "data"
	signature := "signature"

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
	n, _ := node.NewNode()

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	broadcastErr := errors.New("failure")

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTx.SendData = func(data []byte) error {
		return broadcastErr
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(node.TransactionTopic) {
			return topicTx
		}

		return nil
	}

	nonce := uint64(50)
	value := *big.NewInt(567)
	sender := "sender"
	receiver := "receiver"
	txData := "data"
	signature := "signature"

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
	n, _ := node.NewNode()

	mes := mock.NewMessengerStub()
	n.SetMessenger(mes)

	txSent := false

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTx.SendData = func(data []byte) error {
		txSent = true
		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(node.TransactionTopic) {
			return topicTx
		}

		return nil
	}

	nonce := uint64(50)
	value := *big.NewInt(567)
	sender := "sender"
	receiver := "receiver"
	txData := "data"
	signature := "signature"

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
