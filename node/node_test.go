package node_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestNewNode(t *testing.T) {
	t.Parallel()
	n, err := node.NewNode()
	assert.NotNil(t, n)
	assert.Nil(t, err)
}

func TestNewNode_NotRunning(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode()
	assert.False(t, n.IsRunning())
}

func TestNewNode_NilOptionShouldError(t *testing.T) {
	t.Parallel()
	_, err := node.NewNode(node.WithAccountsAdapter(nil))
	assert.NotNil(t, err)
}

func TestNewNode_ApplyNilOptionShouldError(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode()
	err := n.ApplyOptions(node.WithAccountsAdapter(nil))
	assert.NotNil(t, err)
}

func TestStart_NoPort(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode()
	err := n.Start()
	assert.NotNil(t, err)
}

func TestStart_NoMarshalizer(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(node.WithPort(4000))
	err := n.Start()
	assert.NotNil(t, err)
}

func TestStart_NoHasher(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(node.WithPort(4000), node.WithMarshalizer(mock.Marshalizer{}))
	err := n.Start()
	assert.NotNil(t, err)
}

func TestStart_NoMaxAllowedPeers(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(node.WithPort(4000), node.WithMarshalizer(mock.Marshalizer{}), node.WithHasher(mock.Hasher{}))
	err := n.Start()
	assert.NotNil(t, err)
}

func TestStart_CorrectParams(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(&mock.AddressConverter{}),
		node.WithAccountsAdapter(&mock.AccountsAdapter{}),
	)
	err := n.Start()
	assert.Nil(t, err)
	assert.True(t, n.IsRunning())
}

func TestStart_CorrectParamsApplyingOptions(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode()
	err := n.ApplyOptions(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithAddressConverter(&mock.AddressConverter{}),
		node.WithAccountsAdapter(&mock.AccountsAdapter{}),
	)

	logError(err)

	err = n.Start()
	assert.Nil(t, err)
	assert.True(t, n.IsRunning())
}

func TestApplyOptions_NodeStarted(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n.Start()
	logError(err)

	err = n.ApplyOptions(
		node.WithMaxAllowedPeers(4),
	)

	assert.NotNil(t, err)
	assert.True(t, n.IsRunning())
}

func TestStop_NotStartedYet(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
	)
	err := n.Start()
	logError(err)
	err = n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestStop(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
	)

	err := n.Stop()
	assert.Nil(t, err)
	assert.False(t, n.IsRunning())
}

func TestConnectToInitialAddresses_NodeNotStarted(t *testing.T) {
	t.Parallel()
	n2, _ := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithInitialNodeAddresses([]string{addr}),
	)

	err = n.ConnectToInitialAddresses()
	assert.NotNil(t, err)
}

func TestConnectToNilInitialAddresses(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n.Start()
	assert.Nil(t, err)
	err = n.ConnectToInitialAddresses()
	assert.NotNil(t, err)
}

func TestConnectToInitialAddresses(t *testing.T) {
	t.Parallel()
	n2, _ := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithInitialNodeAddresses([]string{addr}),
	)
	err = n.Start()
	assert.Nil(t, err)

	err = n.ConnectToInitialAddresses()
	assert.Nil(t, err)
}

func TestConnectToAddresses_NodeNotStarted(t *testing.T) {
	t.Parallel()
	n2, _ := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)

	err = n.ConnectToAddresses([]string{addr})
	assert.NotNil(t, err)
}

func TestConnectToAddresses(t *testing.T) {
	t.Parallel()
	n2, _ := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err := n2.Start()
	assert.Nil(t, err)
	addr, _ := n2.Address()

	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	err = n.Start()
	assert.Nil(t, err)

	err = n.ConnectToAddresses([]string{addr})
	assert.Nil(t, err)
}

func TestAddress_NodeNotStarted(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
	)
	_, err := n.Address()
	assert.NotNil(t, err)
}

func TestGetBalance_NoAddrConverterShouldError(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAccountsAdapter(&mock.AccountsAdapter{}),
		node.WithPrivateKey(&mock.PrivateKey{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_NoAccAdapterShouldError(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(&mock.AddressConverter{}),
		node.WithPrivateKey(&mock.PrivateKey{}),
	)
	_, err := n.GetBalance("address")
	assert.NotNil(t, err)
	assert.Equal(t, "initialize AccountsAdapter and AddressConverter first", err.Error())
}

func TestGetBalance_CreateAddressFailsShouldError(t *testing.T) {
	t.Parallel()
	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := mock.AddressConverter{
		CreateAddressFromHexHandler: func(hexAddress string) (state.AddressContainer, error) {
			// Return that will result in a correct run of GenerateTransaction -> will fail test
			/*return mock.AddressContainer{
			}, nil*/

			return nil, errors.New("error")
		},
	}
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	accAdapter := mock.AccountsAdapter{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, errors.New("error")
		},
	}
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	accAdapter := mock.AccountsAdapter{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, nil
		},
	}
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	accAdapter := getAccAdapter(*big.NewInt(100))
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAccountsAdapter(&mock.AccountsAdapter{}),
		node.WithPrivateKey(&mock.PrivateKey{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoAccAdapterShouldError(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(&mock.AddressConverter{}),
		node.WithPrivateKey(&mock.PrivateKey{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoPrivateKeyShouldError(t *testing.T) {
	t.Parallel()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAddressConverter(&mock.AddressConverter{}),
		node.WithAccountsAdapter(&mock.AccountsAdapter{}),
	)
	_, err := n.GenerateTransaction("sender", "receiver", *big.NewInt(10), "code")
	assert.NotNil(t, err)
}

func TestGenerateTransaction_CreateAddressFailsShouldError(t *testing.T) {
	t.Parallel()
	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := mock.AddressConverter{
		CreateAddressFromHexHandler: func(hexAddress string) (state.AddressContainer, error) {
			// Return that will result in a correct run of GenerateTransaction -> will fail test
			/*return mock.AddressContainer{
			}, nil*/

			return nil, errors.New("error")
		},
	}
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	accAdapter := mock.AccountsAdapter{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, errors.New("error")
		},
	}
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	accAdapter := mock.AccountsAdapter{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return nil, nil
		},
	}
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	marshalizer := mock.Marshalizer{
		MarshalHandler: func(obj interface{}) ([]byte, error) {
			return nil, errors.New("error")
		},
	}
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(mock.Hasher{}),
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
	t.Parallel()
	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	privateKey := mock.PrivateKey{
		SignHandler: func(message []byte) ([]byte, error) {
			return nil, errors.New("error")
		},
	}
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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

func TestGenerateTransaction_CorrectParamsShouldNotError(t *testing.T) {
	t.Parallel()
	accAdapter := getAccAdapter(*big.NewInt(0))
	addrConverter := getAddressConverter()
	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(mock.Marshalizer{}),
		node.WithHasher(mock.Hasher{}),
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

func getAccAdapter(balance big.Int) mock.AccountsAdapter {
	return mock.AccountsAdapter{
		GetExistingAccountHandler: func(addrContainer state.AddressContainer) (state.AccountWrapper, error) {
			return mock.AccountWrapper{
				BaseAccountHandler: func() *state.Account {
					return &state.Account{
						Nonce: 1,
						Balance: balance,
					}
				},
			}, nil
		},
	}
}

func getPrivateKey() mock.PrivateKey {
	return mock.PrivateKey{
		SignHandler: func(message []byte) ([]byte, error) {
			return []byte{1}, nil
		},
	}
}

func getAddressConverter() mock.AddressConverter {
	return mock.AddressConverter{
		CreateAddressFromHexHandler: func(hexAddress string) (state.AddressContainer, error) {
			// Return that will result in a correct run of GenerateTransaction -> will fail test
			return mock.AddressContainer{
			}, nil
		},
	}
}