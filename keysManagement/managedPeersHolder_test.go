package keysManagement_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/keysManagement"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	p2pPrivateKey = []byte("p2p private key")

	skBytes0        = []byte("private key 0")
	skBytes1        = []byte("private key 1")
	pkBytes0        = []byte("public key 0")
	pkBytes1        = []byte("public key 1")
	defaultName     = "default node name"
	defaultIdentity = "default node identity"
	p2pPkHex        = "03ca8ec5bd3b84d05e59d5c9cecd548059106649d9e3465f628498628732ab23c9"
	p2pPidString    = "16Uiu2HAmSHgyTYyawhsZv9opxTHX77vKjoPeGkyCYS5fYVMssHjN"
	pid, _          = core.NewPeerID(p2pPidString)
)

func createMockArgsManagedPeersHolder() keysManagement.ArgsManagedPeersHolder {
	return keysManagement.ArgsManagedPeersHolder{
		KeyGenerator: createMockKeyGenerator(),
		P2PKeyGenerator: &cryptoMocks.KeyGenStub{
			GeneratePairStub: func() (crypto.PrivateKey, crypto.PublicKey) {
				return &cryptoMocks.PrivateKeyStub{
						ToByteArrayStub: func() ([]byte, error) {
							return []byte("p2p private key"), nil
						},
					},
					&cryptoMocks.PublicKeyStub{
						ToByteArrayStub: func() ([]byte, error) {
							return hex.DecodeString(p2pPkHex)
						},
					}
			},
		},
		IsMainMachine:                    true,
		MaxRoundsWithoutReceivedMessages: 1,
		PrefsConfig: config.Preferences{
			Preferences: config.PreferencesConfig{
				Identity:        defaultIdentity,
				NodeDisplayName: defaultName,
			},
		},
		P2PKeyConverter: &p2pmocks.P2PKeyConverterStub{
			ConvertPublicKeyToPeerIDCalled: func(pk crypto.PublicKey) (core.PeerID, error) {
				return pid, nil
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

func testManagedKeys(tb testing.TB, result map[string]crypto.PrivateKey, pkBytes ...[]byte) {
	assert.Equal(tb, len(pkBytes), len(result))

	for _, pk := range pkBytes {
		_, found := result[string(pk)]
		assert.True(tb, found)
	}
}

func TestNewManagedPeersHolder(t *testing.T) {
	t.Parallel()

	t.Run("nil key generator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsManagedPeersHolder()
		args.KeyGenerator = nil
		holder, err := keysManagement.NewManagedPeersHolder(args)

		assert.ErrorIs(t, err, keysManagement.ErrNilKeyGenerator)
		assert.Contains(t, err.Error(), "for args.KeyGenerator")
		assert.True(t, check.IfNil(holder))
	})
	t.Run("nil p2p key generator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsManagedPeersHolder()
		args.P2PKeyGenerator = nil
		holder, err := keysManagement.NewManagedPeersHolder(args)

		assert.ErrorIs(t, err, keysManagement.ErrNilKeyGenerator)
		assert.Contains(t, err.Error(), "for args.P2PKeyGenerator")
		assert.True(t, check.IfNil(holder))
	})
	t.Run("invalid MaxRoundsWithoutReceivedMessages should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsManagedPeersHolder()
		args.MaxRoundsWithoutReceivedMessages = -2
		holder, err := keysManagement.NewManagedPeersHolder(args)

		assert.True(t, errors.Is(err, keysManagement.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "MaxRoundsWithoutReceivedMessages"))
		assert.True(t, check.IfNil(holder))
	})
	t.Run("invalid key from config should error", func(t *testing.T) {
		t.Parallel()

		providedInvalidKey := "invalid key"
		args := createMockArgsManagedPeersHolder()
		args.PrefsConfig.NamedIdentity = []config.NamedIdentity{
			{
				BLSKeys: []string{providedInvalidKey},
			},
		}
		holder, err := keysManagement.NewManagedPeersHolder(args)

		assert.True(t, errors.Is(err, keysManagement.ErrInvalidKey))
		assert.True(t, strings.Contains(err.Error(), providedInvalidKey))
		assert.True(t, check.IfNil(holder))
	})
	t.Run("nil p2p key converter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsManagedPeersHolder()
		args.P2PKeyConverter = nil
		holder, err := keysManagement.NewManagedPeersHolder(args)

		assert.True(t, errors.Is(err, keysManagement.ErrNilP2PKeyConverter))
		assert.True(t, check.IfNil(holder))
	})
	t.Run("valid arguments should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsManagedPeersHolder()
		holder, err := keysManagement.NewManagedPeersHolder(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(holder))
	})
}

func TestManagedPeersHolder_AddManagedPeer(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("private key from byte array errors", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.KeyGenerator = &cryptoMocks.KeyGenStub{
			PrivateKeyFromByteArrayStub: func(b []byte) (crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}

		holder, _ := keysManagement.NewManagedPeersHolder(args)
		err := holder.AddManagedPeer([]byte("private key"))

		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("public key from byte array errors", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()

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

		holder, _ := keysManagement.NewManagedPeersHolder(args)
		err := holder.AddManagedPeer([]byte("private key"))

		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("p2p key generation returns an invalid private key creation errors", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.P2PKeyGenerator = &cryptoMocks.KeyGenStub{
			GeneratePairStub: func() (crypto.PrivateKey, crypto.PublicKey) {
				return &cryptoMocks.PrivateKeyStub{
					ToByteArrayStub: func() ([]byte, error) {
						return nil, expectedErr
					},
				}, &cryptoMocks.PublicKeyStub{}
			},
		}

		holder, _ := keysManagement.NewManagedPeersHolder(args)
		err := holder.AddManagedPeer([]byte("private key"))

		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work for a new pk", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()

		holder, _ := keysManagement.NewManagedPeersHolder(args)
		err := holder.AddManagedPeer(skBytes0)
		assert.Nil(t, err)

		pInfo := holder.GetPeerInfo(pkBytes0)
		assert.NotNil(t, pInfo)
		assert.Equal(t, pid, pInfo.Pid())
		assert.Equal(t, p2pPrivateKey, pInfo.P2pPrivateKeyBytes())
		skBytesRecovered, _ := pInfo.PrivateKey().ToByteArray()
		assert.Equal(t, skBytes0, skBytesRecovered)
		assert.Equal(t, 10, len(pInfo.MachineID()))
		assert.Equal(t, defaultIdentity, pInfo.NodeIdentity())
		assert.Equal(t, defaultName+"-00", pInfo.NodeName())
	})
	t.Run("should work for a new pk with identity from config", func(t *testing.T) {
		providedAddress := []byte("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
		providedHex := hex.EncodeToString(providedAddress)
		providedName := "provided name"
		providedIdentity := "provided identity"
		args := createMockArgsManagedPeersHolder()
		args.KeyGenerator = &cryptoMocks.KeyGenStub{
			PrivateKeyFromByteArrayStub: func(b []byte) (crypto.PrivateKey, error) {
				return &cryptoMocks.PrivateKeyStub{
					GeneratePublicStub: func() crypto.PublicKey {
						return &cryptoMocks.PublicKeyStub{
							ToByteArrayStub: func() ([]byte, error) {
								return providedAddress, nil
							},
						}
					},
					ToByteArrayStub: func() ([]byte, error) {
						return providedAddress, nil
					},
				}, nil
			},
		}
		args.PrefsConfig.NamedIdentity = []config.NamedIdentity{
			{
				Identity: providedIdentity,
				NodeName: providedName,
				BLSKeys:  []string{providedHex},
			},
		}

		holder, _ := keysManagement.NewManagedPeersHolder(args)
		err := holder.AddManagedPeer(skBytes0)
		assert.Nil(t, err)

		pInfo := holder.GetPeerInfo(providedAddress)
		assert.NotNil(t, pInfo)
		assert.Equal(t, pid, pInfo.Pid())
		assert.Equal(t, p2pPrivateKey, pInfo.P2pPrivateKeyBytes())
		skBytesRecovered, _ := pInfo.PrivateKey().ToByteArray()
		assert.Equal(t, providedAddress, skBytesRecovered)
		assert.Equal(t, 10, len(pInfo.MachineID()))
		assert.Equal(t, providedIdentity, pInfo.NodeIdentity())
		assert.Equal(t, providedName+"-00", pInfo.NodeName())
	})
	t.Run("should error when trying to add the same pk", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()

		holder, _ := keysManagement.NewManagedPeersHolder(args)
		err := holder.AddManagedPeer(skBytes0)
		assert.Nil(t, err)

		err = holder.AddManagedPeer(skBytes0)
		assert.True(t, errors.Is(err, keysManagement.ErrDuplicatedKey))
	})
	t.Run("should work for 2 new public keys", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()

		holder, _ := keysManagement.NewManagedPeersHolder(args)
		err := holder.AddManagedPeer(skBytes0)
		assert.Nil(t, err)

		err = holder.AddManagedPeer(skBytes1)
		assert.Nil(t, err)

		pInfo0 := holder.GetPeerInfo(pkBytes0)
		assert.NotNil(t, pInfo0)

		pInfo1 := holder.GetPeerInfo(pkBytes1)
		assert.NotNil(t, pInfo1)

		assert.Equal(t, p2pPrivateKey, pInfo0.P2pPrivateKeyBytes())
		assert.Equal(t, p2pPrivateKey, pInfo1.P2pPrivateKeyBytes())

		assert.Equal(t, pid, pInfo0.Pid())
		assert.Equal(t, pid, pInfo1.Pid())

		skBytesRecovered0, _ := pInfo0.PrivateKey().ToByteArray()
		assert.Equal(t, skBytes0, skBytesRecovered0)

		skBytesRecovered1, _ := pInfo1.PrivateKey().ToByteArray()
		assert.Equal(t, skBytes1, skBytesRecovered1)

		assert.NotEqual(t, pInfo0.MachineID(), pInfo1.MachineID())
		assert.Equal(t, 10, len(pInfo0.MachineID()))
		assert.Equal(t, 10, len(pInfo1.MachineID()))
	})
}

func TestManagedPeersHolder_GetPrivateKey(t *testing.T) {
	t.Parallel()

	args := createMockArgsManagedPeersHolder()

	holder, _ := keysManagement.NewManagedPeersHolder(args)
	_ = holder.AddManagedPeer(skBytes0)

	testName := "test name"
	testIdentity := "test identity"

	t.Run("public key not added should error", func(t *testing.T) {
		skRecovered, err := holder.GetPrivateKey(pkBytes1)
		assert.Nil(t, skRecovered)
		assert.True(t, errors.Is(err, keysManagement.ErrMissingPublicKeyDefinition))
	})
	t.Run("identity provided but not the actual key should not provide a nil key", func(t *testing.T) {
		argsLocal := createMockArgsManagedPeersHolder()
		blsKey := "aabb1122"
		namedIdentity := config.NamedIdentity{
			Identity: testIdentity,
			NodeName: testName,
			BLSKeys:  []string{blsKey},
		}

		argsLocal.PrefsConfig.NamedIdentity = append(argsLocal.PrefsConfig.NamedIdentity, namedIdentity)
		holderLocal, err := keysManagement.NewManagedPeersHolder(argsLocal)
		assert.Nil(t, err)

		keyUnhexed, _ := hex.DecodeString(blsKey)
		skRecovered, err := holderLocal.GetPrivateKey(keyUnhexed)
		assert.Nil(t, skRecovered)
		assert.ErrorIs(t, err, keysManagement.ErrMissingPublicKeyDefinition)

		name, identity, err := holderLocal.GetNameAndIdentity(keyUnhexed)
		assert.ErrorIs(t, err, keysManagement.ErrMissingPublicKeyDefinition)
		assert.Empty(t, name)
		assert.Empty(t, identity)
	})
	t.Run("public key exists should return the private key", func(t *testing.T) {
		skRecovered, err := holder.GetPrivateKey(pkBytes0)
		assert.Nil(t, err)

		skBytesRecovered, _ := skRecovered.ToByteArray()
		assert.Equal(t, skBytes0, skBytesRecovered)
	})
	t.Run("identity provided and the actual key should work", func(t *testing.T) {
		argsLocal := createMockArgsManagedPeersHolder()
		namedIdentity := config.NamedIdentity{
			Identity: testIdentity,
			NodeName: testName,
			BLSKeys:  []string{hex.EncodeToString(pkBytes0)},
		}

		argsLocal.PrefsConfig.NamedIdentity = append(argsLocal.PrefsConfig.NamedIdentity, namedIdentity)
		holderLocal, err := keysManagement.NewManagedPeersHolder(argsLocal)
		assert.Nil(t, err)

		_ = holderLocal.AddManagedPeer(skBytes0)
		skRecovered, err := holderLocal.GetPrivateKey(pkBytes0)
		skBytesRecovered, _ := skRecovered.ToByteArray()
		assert.Equal(t, skBytes0, skBytesRecovered)
		assert.Nil(t, err)

		name, identity, err := holderLocal.GetNameAndIdentity(pkBytes0)
		assert.Nil(t, err)
		assert.Equal(t, testName+"-00", name)
		assert.Equal(t, testIdentity, identity)
	})
}

func TestManagedPeersHolder_GetP2PIdentity(t *testing.T) {
	t.Parallel()

	args := createMockArgsManagedPeersHolder()

	holder, _ := keysManagement.NewManagedPeersHolder(args)
	_ = holder.AddManagedPeer(skBytes0)
	t.Run("public key not added should error", func(t *testing.T) {
		p2pPrivateKeyRecovered, pidRecovered, err := holder.GetP2PIdentity(pkBytes1)
		assert.Nil(t, p2pPrivateKeyRecovered)
		assert.Empty(t, pidRecovered)
		assert.True(t, errors.Is(err, keysManagement.ErrMissingPublicKeyDefinition))
	})
	t.Run("public key exists should return the p2p identity", func(t *testing.T) {
		p2pPrivateKeyRecovered, pidRecovered, err := holder.GetP2PIdentity(pkBytes0)
		assert.Nil(t, err)
		assert.Equal(t, p2pPrivateKey, p2pPrivateKeyRecovered)
		assert.Equal(t, pid, pidRecovered)
	})
}

func TestManagedPeersHolder_GetMachineID(t *testing.T) {
	t.Parallel()

	args := createMockArgsManagedPeersHolder()
	holder, _ := keysManagement.NewManagedPeersHolder(args)
	_ = holder.AddManagedPeer(skBytes0)
	t.Run("public key not added should error", func(t *testing.T) {
		machineIDRecovered, err := holder.GetMachineID(pkBytes1)
		assert.Empty(t, machineIDRecovered)
		assert.True(t, errors.Is(err, keysManagement.ErrMissingPublicKeyDefinition))
	})
	t.Run("public key exists should return machine ID", func(t *testing.T) {
		machineIDRecovered, err := holder.GetMachineID(pkBytes0)
		assert.Nil(t, err)
		assert.Equal(t, 10, len(machineIDRecovered))
	})
}

func TestManagedPeersHolder_GetNameAndIdentity(t *testing.T) {
	t.Parallel()

	args := createMockArgsManagedPeersHolder()
	holder, _ := keysManagement.NewManagedPeersHolder(args)
	_ = holder.AddManagedPeer(skBytes0)
	t.Run("public key not added should error", func(t *testing.T) {
		name, identity, err := holder.GetNameAndIdentity(pkBytes1)
		assert.Empty(t, name)
		assert.Empty(t, identity)
		assert.True(t, errors.Is(err, keysManagement.ErrMissingPublicKeyDefinition))
	})
	t.Run("public key exists should return name and identity", func(t *testing.T) {
		name, identity, err := holder.GetNameAndIdentity(pkBytes0)
		assert.Nil(t, err)
		assert.Equal(t, defaultName+"-00", name)
		assert.Equal(t, defaultIdentity, identity)
	})
	t.Run("complex scenarios with multiple identities should work", func(t *testing.T) {
		argsLocal := createMockArgsManagedPeersHolder()
		argsLocal.PrefsConfig.NamedIdentity = []config.NamedIdentity{
			{
				Identity: "identity1",
				NodeName: "name1",
				BLSKeys: []string{
					hex.EncodeToString([]byte("public key 0")),
					"",
					hex.EncodeToString([]byte("public key 1")),
				},
			},
			{
				Identity: "identity2",
				NodeName: "name2",
				BLSKeys: []string{
					"",
					hex.EncodeToString([]byte("public key 2")),
					hex.EncodeToString([]byte("public key 3")),
				},
			},
			{
				Identity: "identity3",
				NodeName: "",
				BLSKeys: []string{
					hex.EncodeToString([]byte("public key 4")),
					"",
					hex.EncodeToString([]byte("public key 5")),
				},
			},
		}

		holderLocal, err := keysManagement.NewManagedPeersHolder(argsLocal)
		require.Nil(t, err)

		for i := 0; i < 10; i++ {
			_ = holderLocal.AddManagedPeer([]byte(fmt.Sprintf("private key %d", i)))
		}

		checkNameIdentity(t, holderLocal, "public key 0", "identity1", "name1-00")
		checkNameIdentity(t, holderLocal, "public key 1", "identity1", "name1-01")

		checkNameIdentity(t, holderLocal, "public key 2", "identity2", "name2-00")
		checkNameIdentity(t, holderLocal, "public key 3", "identity2", "name2-01")

		checkNameIdentity(t, holderLocal, "public key 4", "identity3", "")
		checkNameIdentity(t, holderLocal, "public key 5", "identity3", "")

		checkNameIdentity(t, holderLocal, "public key 6", defaultIdentity, defaultName+"-00")
		checkNameIdentity(t, holderLocal, "public key 7", defaultIdentity, defaultName+"-01")
		checkNameIdentity(t, holderLocal, "public key 8", defaultIdentity, defaultName+"-02")
	})
}

func checkNameIdentity(tb testing.TB, holder common.ManagedPeersHolder, pk string, expectedIdentity string, expectedName string) {
	name, identity, err := holder.GetNameAndIdentity([]byte(pk))
	assert.Nil(tb, err)
	assert.Equal(tb, expectedIdentity, identity)
	assert.Equal(tb, expectedName, name)
}

func TestManagedPeersHolder_IncrementRoundsWithoutReceivedMessages(t *testing.T) {
	t.Parallel()

	t.Run("is main machine should ignore the call", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = true
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)

		t.Run("missing public key should not panic", func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
				}
			}()

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes1)
			pInfoRecovered := holder.GetPeerInfo(pkBytes1)
			assert.Nil(t, pInfoRecovered)
		})
		t.Run("existing public key", func(t *testing.T) {
			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)

			pInfoRecovered := holder.GetPeerInfo(pkBytes0)
			assert.Zero(t, pInfoRecovered.GetRoundsWithoutReceivedMessages())
		})
	})
	t.Run("is secondary machine should increment, if existing", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = false
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)

		t.Run("missing public key should not panic", func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
				}
			}()

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes1)
			pInfoRecovered := holder.GetPeerInfo(pkBytes1)
			assert.Nil(t, pInfoRecovered)
		})
		t.Run("existing public key should increment", func(t *testing.T) {
			pInfoRecovered := holder.GetPeerInfo(pkBytes0)
			assert.Zero(t, pInfoRecovered.GetRoundsWithoutReceivedMessages())

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)

			pInfoRecovered = holder.GetPeerInfo(pkBytes0)
			assert.Equal(t, 1, pInfoRecovered.GetRoundsWithoutReceivedMessages())

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)

			pInfoRecovered = holder.GetPeerInfo(pkBytes0)
			assert.Equal(t, 2, pInfoRecovered.GetRoundsWithoutReceivedMessages())
		})
	})
}

func TestManagedPeersHolder_ResetRoundsWithoutReceivedMessages(t *testing.T) {
	t.Parallel()

	t.Run("is main machine should ignore the call", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = true
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)

		t.Run("missing public key should not panic", func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
				}
			}()

			holder.ResetRoundsWithoutReceivedMessages(pkBytes1)
		})
		t.Run("existing public key", func(t *testing.T) {
			holder.ResetRoundsWithoutReceivedMessages(pkBytes0)

			pInfoRecovered := holder.GetPeerInfo(pkBytes0)
			assert.Zero(t, pInfoRecovered.GetRoundsWithoutReceivedMessages())
		})
	})
	t.Run("is secondary machine should reset, if existing", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = false
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)

		t.Run("missing public key should not panic", func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
				}
			}()

			holder.ResetRoundsWithoutReceivedMessages(pkBytes1)
		})
		t.Run("existing public key should reset", func(t *testing.T) {
			pInfoRecovered := holder.GetPeerInfo(pkBytes0)
			assert.Zero(t, pInfoRecovered.GetRoundsWithoutReceivedMessages())

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
			pInfoRecovered = holder.GetPeerInfo(pkBytes0)
			assert.Equal(t, 1, pInfoRecovered.GetRoundsWithoutReceivedMessages())

			holder.ResetRoundsWithoutReceivedMessages(pkBytes0)

			pInfoRecovered = holder.GetPeerInfo(pkBytes0)
			assert.Equal(t, 0, pInfoRecovered.GetRoundsWithoutReceivedMessages())

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
			pInfoRecovered = holder.GetPeerInfo(pkBytes0)
			assert.Equal(t, 3, pInfoRecovered.GetRoundsWithoutReceivedMessages())

			holder.ResetRoundsWithoutReceivedMessages(pkBytes0)

			pInfoRecovered = holder.GetPeerInfo(pkBytes0)
			assert.Equal(t, 0, pInfoRecovered.GetRoundsWithoutReceivedMessages())
		})
	})
}

func TestManagedPeersHolder_GetManagedKeysByCurrentNode(t *testing.T) {
	t.Parallel()

	t.Run("main machine should return all keys, always", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = true
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)
		_ = holder.AddManagedPeer(skBytes1)

		for i := 0; i < 10; i++ {
			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
		}

		result := holder.GetManagedKeysByCurrentNode()
		testManagedKeys(t, result, pkBytes0, pkBytes1)
	})
	t.Run("is secondary machine should return managed keys", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = false
		args.MaxRoundsWithoutReceivedMessages = 2
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)
		_ = holder.AddManagedPeer(skBytes1)

		t.Run("MaxRoundsWithoutReceivedMessages not reached should return none", func(t *testing.T) {
			result := holder.GetManagedKeysByCurrentNode()
			testManagedKeys(t, result)

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)

			result = holder.GetManagedKeysByCurrentNode()
			testManagedKeys(t, result)
		})
		t.Run("MaxRoundsWithoutReceivedMessages reached, should return failed pk", func(t *testing.T) {
			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)

			result := holder.GetManagedKeysByCurrentNode()
			testManagedKeys(t, result, pkBytes0)

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
			result = holder.GetManagedKeysByCurrentNode()
			testManagedKeys(t, result, pkBytes0)
		})
	})
}

func TestManagedPeersHolder_IsKeyManagedByCurrentNode(t *testing.T) {
	t.Parallel()

	t.Run("main machine", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = true
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)

		t.Run("foreign public key should return false", func(t *testing.T) {
			isManaged := holder.IsKeyManagedByCurrentNode(pkBytes1)
			assert.False(t, isManaged)
		})
		t.Run("managed key should return true", func(t *testing.T) {
			isManaged := holder.IsKeyManagedByCurrentNode(pkBytes0)
			assert.True(t, isManaged)
		})
	})
	t.Run("secondary machine", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = false
		args.MaxRoundsWithoutReceivedMessages = 2
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)

		t.Run("foreign public key should return false", func(t *testing.T) {
			isManaged := holder.IsKeyManagedByCurrentNode(pkBytes1)
			assert.False(t, isManaged)
		})
		t.Run("managed key should return false while MaxRoundsWithoutReceivedMessages is not reached", func(t *testing.T) {
			isManaged := holder.IsKeyManagedByCurrentNode(pkBytes0)
			assert.False(t, isManaged)

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
			isManaged = holder.IsKeyManagedByCurrentNode(pkBytes0)
			assert.False(t, isManaged)

			holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
			isManaged = holder.IsKeyManagedByCurrentNode(pkBytes0)
			assert.True(t, isManaged)

			holder.ResetRoundsWithoutReceivedMessages(pkBytes0)
			isManaged = holder.IsKeyManagedByCurrentNode(pkBytes0)
			assert.False(t, isManaged)
		})
	})
}

func TestManagedPeersHolder_IsKeyRegistered(t *testing.T) {
	t.Parallel()

	t.Run("main machine", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = true
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)

		t.Run("foreign public key should return false", func(t *testing.T) {
			isManaged := holder.IsKeyRegistered(pkBytes1)
			assert.False(t, isManaged)
		})
		t.Run("registered key should return true", func(t *testing.T) {
			isManaged := holder.IsKeyRegistered(pkBytes0)
			assert.True(t, isManaged)
		})
	})
	t.Run("secondary machine", func(t *testing.T) {
		args := createMockArgsManagedPeersHolder()
		args.IsMainMachine = false
		holder, _ := keysManagement.NewManagedPeersHolder(args)
		_ = holder.AddManagedPeer(skBytes0)

		t.Run("foreign public key should return false", func(t *testing.T) {
			isManaged := holder.IsKeyRegistered(pkBytes1)
			assert.False(t, isManaged)
		})
		t.Run("registered key should return true", func(t *testing.T) {
			isManaged := holder.IsKeyRegistered(pkBytes0)
			assert.True(t, isManaged)
		})
	})
}

func TestManagedPeersHolder_IsPidManagedByCurrentNode(t *testing.T) {
	t.Parallel()

	args := createMockArgsManagedPeersHolder()
	args.IsMainMachine = true
	holder, _ := keysManagement.NewManagedPeersHolder(args)

	t.Run("empty holder should return false", func(t *testing.T) {
		isManaged := holder.IsPidManagedByCurrentNode(pid)
		assert.False(t, isManaged)
	})

	_ = holder.AddManagedPeer(skBytes0)

	t.Run("pid not managed by current should return false", func(t *testing.T) {
		isManaged := holder.IsPidManagedByCurrentNode("other pid")
		assert.False(t, isManaged)
	})
	t.Run("pid managed by current should return true", func(t *testing.T) {
		isManaged := holder.IsPidManagedByCurrentNode(pid)
		assert.True(t, isManaged)
	})
}

func TestManagedPeersHolder_IsKeyValidator(t *testing.T) {
	t.Parallel()

	holder, _ := keysManagement.NewManagedPeersHolder(createMockArgsManagedPeersHolder())

	t.Run("missing key should return false", func(t *testing.T) {
		isValidator := holder.IsKeyValidator(pkBytes0)
		assert.False(t, isValidator)
	})

	_ = holder.AddManagedPeer(skBytes0)

	t.Run("key found, but not validator should return false", func(t *testing.T) {
		isValidator := holder.IsKeyValidator(pkBytes0)
		assert.False(t, isValidator)
	})
	t.Run("key found and validator should return true", func(t *testing.T) {
		holder.SetValidatorState(pkBytes0, true)
		isValidator := holder.IsKeyValidator(pkBytes0)
		assert.True(t, isValidator)
	})
}

func TestManagedPeersHolder_GetNextPeerAuthenticationTime(t *testing.T) {
	t.Parallel()

	holder, _ := keysManagement.NewManagedPeersHolder(createMockArgsManagedPeersHolder())

	t.Run("missing key should return error", func(t *testing.T) {
		timeBefore := time.Now()
		nextTime, err := holder.GetNextPeerAuthenticationTime(pkBytes0)
		timeAfter := time.Now()
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, keysManagement.ErrMissingPublicKeyDefinition))
		assert.True(t, strings.Contains(err.Error(), hex.EncodeToString(pkBytes0)))
		assert.LessOrEqual(t, nextTime, timeAfter)
		assert.Greater(t, nextTime, timeBefore)
	})

	_ = holder.AddManagedPeer(skBytes0)

	t.Run("key found should work", func(t *testing.T) {
		expectedNextTime := time.Now().Add(time.Hour)
		holder.SetNextPeerAuthenticationTime(pkBytes0, expectedNextTime)
		nextTime, err := holder.GetNextPeerAuthenticationTime(pkBytes0)
		assert.Nil(t, err)
		assert.Equal(t, expectedNextTime, nextTime)
	})
}

func TestManagedPeersHolder_IsMultiKeyMode(t *testing.T) {
	t.Parallel()

	args := createMockArgsManagedPeersHolder()
	holder, _ := keysManagement.NewManagedPeersHolder(args)
	assert.False(t, holder.IsMultiKeyMode())

	_ = holder.AddManagedPeer(skBytes0)
	assert.True(t, holder.IsMultiKeyMode())
}

func TestManagedPeersHolder_ParallelOperationsShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	args := createMockArgsManagedPeersHolder()
	holder, _ := keysManagement.NewManagedPeersHolder(args)

	numOperations := 1500
	wg := sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(idOperation int) {
			time.Sleep(time.Millisecond * 10) // increase the chance of concurrent operations

			switch idOperation {
			case 0:
				randomBytes := make([]byte, 32)
				_, _ = rand.Read(randomBytes)
				_ = holder.AddManagedPeer(randomBytes)
			case 1:
				_, _ = holder.GetMachineID(pkBytes1)
			case 2:
				_, _, _ = holder.GetP2PIdentity(pkBytes1)
			case 3:
				_, _ = holder.GetPrivateKey(pkBytes1)
			case 4:
				holder.IncrementRoundsWithoutReceivedMessages(pkBytes0)
			case 5:
				holder.ResetRoundsWithoutReceivedMessages(pkBytes0)
			case 6:
				_ = holder.GetManagedKeysByCurrentNode()
			case 7:
				_ = holder.IsKeyManagedByCurrentNode(pkBytes0)
			case 8:
				_ = holder.IsKeyRegistered(pkBytes0)
			case 9:
				_ = holder.IsPidManagedByCurrentNode("pid")
			case 10:
				_ = holder.IsKeyValidator(pkBytes0)
			case 11:
				holder.SetValidatorState(pkBytes0, true)
			case 12:
				_, _ = holder.GetNextPeerAuthenticationTime(pkBytes0)
			case 13:
				holder.SetNextPeerAuthenticationTime(pkBytes0, time.Now())
			}

			wg.Done()
		}(i % 14)
	}

	wg.Wait()
}
