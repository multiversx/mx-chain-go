package keysManagement

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/keysManagement/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/assert"
)

var (
	p2pPrivateKey = []byte("p2p private key")
	pid           = core.PeerID("pid")
	skBytes0      = []byte("private key 0")
	skBytes1      = []byte("private key 1")
	pkBytes0      = []byte("public key 0")
	pkBytes1      = []byte("public key 1")
)

func createMockArgsVirtualPeersHolder() ArgsVirtualPeersHolder {
	return ArgsVirtualPeersHolder{
		KeyGenerator: &cryptoMocks.KeyGenStub{},
		P2PIdentityGenerator: &mock.IdentityGeneratorStub{
			CreateRandomP2PIdentityCalled: func() ([]byte, core.PeerID, error) {
				return p2pPrivateKey, pid, nil
			},
		},
	}
}

func createMockKeyGenerator() crypto.KeyGenerator {
	return &cryptoMocks.KeyGenStub{
		PrivateKeyFromByteArrayStub: func(b []byte) (crypto.PrivateKey, error) {
			return &cryptoMocks.PrivateKeyStub{
				GeneratePublicStub: func() crypto.PublicKey {
					pk := &cryptoMocks.PublicKeyStub{
						ToByteArrayStub: func() ([]byte, error) {
							pkBytes := bytes.Replace(b, []byte("private"), []byte("public"), -1)

							return pkBytes, nil
						},
					}

					return pk
				},
				ToByteArrayStub: func() ([]byte, error) {
					return b, nil
				},
			}, nil
		},
	}

}

func TestNewVirtualPeersHolder(t *testing.T) {
	t.Parallel()

	t.Run("nil key generator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsVirtualPeersHolder()
		args.KeyGenerator = nil
		holder, err := NewVirtualPeersHolder(args)

		assert.Equal(t, errNilKeyGenerator, err)
		assert.True(t, check.IfNil(holder))
	})
	t.Run("nil identity generator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsVirtualPeersHolder()
		args.P2PIdentityGenerator = nil
		holder, err := NewVirtualPeersHolder(args)

		assert.Equal(t, errNilP2PIdentityGenerator, err)
		assert.True(t, check.IfNil(holder))
	})
	t.Run("valid arguments should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsVirtualPeersHolder()
		holder, err := NewVirtualPeersHolder(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(holder))
	})
}

func TestVirtualPeersHolder_AddVirtualPeer(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("private key from byte array errors", func(t *testing.T) {
		args := createMockArgsVirtualPeersHolder()
		args.KeyGenerator = &cryptoMocks.KeyGenStub{
			PrivateKeyFromByteArrayStub: func(b []byte) (crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}

		holder, _ := NewVirtualPeersHolder(args)
		err := holder.AddVirtualPeer([]byte("private key"))

		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("public key from byte array errors", func(t *testing.T) {
		args := createMockArgsVirtualPeersHolder()

		pk := &cryptoMocks.PublicKeyStub{
			ToByteArrayStub: func() ([]byte, error) {
				return nil, expectedErr
			},
		}

		sk := &cryptoMocks.PrivateKeyStub{
			GeneratePublicStub: func() crypto.PublicKey {
				return pk
			},
		}

		args.KeyGenerator = &cryptoMocks.KeyGenStub{
			PrivateKeyFromByteArrayStub: func(b []byte) (crypto.PrivateKey, error) {
				return sk, nil
			},
		}

		holder, _ := NewVirtualPeersHolder(args)
		err := holder.AddVirtualPeer([]byte("private key"))

		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("identity creation errors should errors", func(t *testing.T) {
		args := createMockArgsVirtualPeersHolder()
		args.P2PIdentityGenerator = &mock.IdentityGeneratorStub{
			CreateRandomP2PIdentityCalled: func() ([]byte, core.PeerID, error) {
				return nil, "", expectedErr
			},
		}

		holder, _ := NewVirtualPeersHolder(args)
		err := holder.AddVirtualPeer([]byte("private key"))

		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work for a new pk", func(t *testing.T) {
		args := createMockArgsVirtualPeersHolder()
		args.KeyGenerator = createMockKeyGenerator()

		holder, _ := NewVirtualPeersHolder(args)
		err := holder.AddVirtualPeer(skBytes0)
		assert.Nil(t, err)

		pInfo := holder.getPeerInfo(pkBytes0)
		assert.NotNil(t, pInfo)
		assert.Equal(t, pid, pInfo.pid)
		assert.Equal(t, p2pPrivateKey, pInfo.p2pPrivateKeyBytes)
		skBytesRecovered, _ := pInfo.privateKey.ToByteArray()
		assert.Equal(t, skBytes0, skBytesRecovered)
		assert.Equal(t, 10, len(pInfo.machineID))
	})
	t.Run("should error when trying to add the same pk", func(t *testing.T) {
		args := createMockArgsVirtualPeersHolder()
		args.KeyGenerator = createMockKeyGenerator()

		holder, _ := NewVirtualPeersHolder(args)
		err := holder.AddVirtualPeer(skBytes0)
		assert.Nil(t, err)

		err = holder.AddVirtualPeer(skBytes0)
		assert.True(t, errors.Is(err, errDuplicatedKey))
	})
	t.Run("should work for 2 new public keys", func(t *testing.T) {
		args := createMockArgsVirtualPeersHolder()
		args.KeyGenerator = createMockKeyGenerator()

		holder, _ := NewVirtualPeersHolder(args)
		err := holder.AddVirtualPeer(skBytes0)
		assert.Nil(t, err)

		err = holder.AddVirtualPeer(skBytes1)
		assert.Nil(t, err)

		pInfo0 := holder.getPeerInfo(pkBytes0)
		assert.NotNil(t, pInfo0)

		pInfo1 := holder.getPeerInfo(pkBytes1)
		assert.NotNil(t, pInfo1)

		assert.Equal(t, p2pPrivateKey, pInfo0.p2pPrivateKeyBytes)
		assert.Equal(t, p2pPrivateKey, pInfo1.p2pPrivateKeyBytes)

		assert.Equal(t, pid, pInfo0.pid)
		assert.Equal(t, pid, pInfo1.pid)

		skBytesRecovered0, _ := pInfo0.privateKey.ToByteArray()
		assert.Equal(t, skBytes0, skBytesRecovered0)

		skBytesRecovered1, _ := pInfo1.privateKey.ToByteArray()
		assert.Equal(t, skBytes1, skBytesRecovered1)

		assert.NotEqual(t, pInfo0.machineID, pInfo1.machineID)
		assert.Equal(t, 10, len(pInfo0.machineID))
		assert.Equal(t, 10, len(pInfo1.machineID))
	})
}

func TestVirtualPeersHolder_GetPrivateKey(t *testing.T) {
	t.Parallel()

	args := createMockArgsVirtualPeersHolder()
	args.KeyGenerator = createMockKeyGenerator()

	holder, _ := NewVirtualPeersHolder(args)
	_ = holder.AddVirtualPeer(skBytes0)
	t.Run("public key not added should error", func(t *testing.T) {
		skRecovered, err := holder.GetPrivateKey(pkBytes1)
		assert.Nil(t, skRecovered)
		assert.True(t, errors.Is(err, errMissingPublicKeyDefinition))
	})
	t.Run("public key exists should return the private key", func(t *testing.T) {
		skRecovered, err := holder.GetPrivateKey(pkBytes0)
		assert.Nil(t, err)

		skBytesRecovered, _ := skRecovered.ToByteArray()
		assert.Equal(t, skBytes0, skBytesRecovered)
	})
}

func TestVirtualPeersHolder_GetP2PIdentity(t *testing.T) {
	t.Parallel()

	args := createMockArgsVirtualPeersHolder()
	args.KeyGenerator = createMockKeyGenerator()

	holder, _ := NewVirtualPeersHolder(args)
	_ = holder.AddVirtualPeer(skBytes0)
	t.Run("public key not added should error", func(t *testing.T) {
		p2pPrivateKeyRecovered, pidRecovered, err := holder.GetP2PIdentity(pkBytes1)
		assert.Nil(t, p2pPrivateKeyRecovered)
		assert.Empty(t, pidRecovered)
		assert.True(t, errors.Is(err, errMissingPublicKeyDefinition))
	})
	t.Run("public key exists should return the p2p identity", func(t *testing.T) {
		p2pPrivateKeyRecovered, pidRecovered, err := holder.GetP2PIdentity(pkBytes0)
		assert.Nil(t, err)
		assert.Equal(t, p2pPrivateKey, p2pPrivateKeyRecovered)
		assert.Equal(t, pid, pidRecovered)
	})
}

func TestVirtualPeersHolder_GetMachineID(t *testing.T) {
	t.Parallel()

	args := createMockArgsVirtualPeersHolder()
	args.KeyGenerator = createMockKeyGenerator()

	holder, _ := NewVirtualPeersHolder(args)
	_ = holder.AddVirtualPeer(skBytes0)
	t.Run("public key not added should error", func(t *testing.T) {
		machineIDRecovered, err := holder.GetMachineID(pkBytes1)
		assert.Empty(t, machineIDRecovered)
		assert.True(t, errors.Is(err, errMissingPublicKeyDefinition))
	})
	t.Run("public key exists should return machine ID", func(t *testing.T) {
		machineIDRecovered, err := holder.GetMachineID(pkBytes0)
		assert.Nil(t, err)
		assert.Equal(t, 10, len(machineIDRecovered))
	})
}

func TestVirtualPeersHolder_ParallelOperationsShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	args := createMockArgsVirtualPeersHolder()
	args.KeyGenerator = createMockKeyGenerator()
	holder, _ := NewVirtualPeersHolder(args)

	numOperations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(idOperation int) {
			time.Sleep(time.Millisecond * 10) // increase the chance of concurrent operations

			switch idOperation {
			case 0:
				randomBytes := make([]byte, 32)
				_, _ = rand.Read(randomBytes)
				_ = holder.AddVirtualPeer(randomBytes)
			case 1:
				_, _ = holder.GetMachineID(pkBytes1)
			case 2:
				_, _, _ = holder.GetP2PIdentity(pkBytes1)
			case 3:
				_, _ = holder.GetPrivateKey(pkBytes1)
			}

			wg.Done()
		}(i % 4)
	}

	wg.Wait()
}
