package components

import (
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func createMockArgsTestOnlyProcessingNode(t *testing.T) ArgsTestOnlyProcessingNode {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:           3,
		OriginalConfigsPath:   "../../../cmd/node/config/",
		GenesisTimeStamp:      0,
		RoundDurationInMillis: 6000,
		TempDir:               t.TempDir(),
		MinNodesPerShard:      1,
		MetaChainMinNodes:     1,
	})
	require.Nil(t, err)

	return ArgsTestOnlyProcessingNode{
		Configs:             outputConfigs.Configs,
		GasScheduleFilename: outputConfigs.GasScheduleFilename,
		NumShards:           3,

		SyncedBroadcastNetwork: NewSyncedBroadcastNetwork(),
		ChanStopNodeProcess:    make(chan endProcess.ArgEndProcess),
		APIInterface:           api.NewNoApiInterface(),
		ShardIDStr:             "0",
	}
}

func TestNewTestOnlyProcessingNode(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)
	})
	t.Run("try commit a block", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)

		newHeader, err := node.ProcessComponentsHolder.BlockProcessor().CreateNewHeader(1, 1)
		assert.Nil(t, err)

		err = newHeader.SetPrevHash(node.ChainHandler.GetGenesisHeaderHash())
		assert.Nil(t, err)

		header, block, err := node.ProcessComponentsHolder.BlockProcessor().CreateBlock(newHeader, func() bool {
			return true
		})
		assert.Nil(t, err)
		require.NotNil(t, header)
		require.NotNil(t, block)

		err = node.ProcessComponentsHolder.BlockProcessor().ProcessBlock(header, block, func() time.Duration {
			return 1000
		})
		assert.Nil(t, err)

		err = node.ProcessComponentsHolder.BlockProcessor().CommitBlock(header, block)
		assert.Nil(t, err)
	})
	t.Run("CreateCoreComponents failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.GeneralConfig.Marshalizer.Type = "invalid type"
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateCryptoComponents failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.GeneralConfig.PublicKeyPIDSignature.Type = "invalid type"
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateNetworkComponents failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		args.SyncedBroadcastNetwork = nil
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateBootstrapComponents failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.FlagsConfig.WorkingDir = ""
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateStateComponents failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		args.ShardIDStr = common.MetachainShardName // coverage only
		args.Configs.GeneralConfig.StateTriesConfig.MaxStateTrieLevelInMemory = 0
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateProcessComponents failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.FlagsConfig.Version = ""
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("createFacade failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.EpochConfig.GasSchedule.GasScheduleByEpochs = nil
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
}

func TestTestOnlyProcessingNode_SetKeyValueForAddress(t *testing.T) {
	t.Parallel()

	goodKeyValueMap := map[string]string{
		"01": "02",
	}
	node, err := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
	require.NoError(t, err)

	address := "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj"
	addressBytes, _ := node.CoreComponentsHolder.AddressPubKeyConverter().Decode(address)

	t.Run("should work", func(t *testing.T) {
		_, err = node.StateComponentsHolder.AccountsAdapter().GetExistingAccount(addressBytes)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "account was not found"))

		err = node.SetKeyValueForAddress(addressBytes, goodKeyValueMap)
		require.NoError(t, err)

		_, err = node.StateComponentsHolder.AccountsAdapter().GetExistingAccount(addressBytes)
		require.NoError(t, err)
	})
	t.Run("decode key failure should error", func(t *testing.T) {
		keyValueMap := map[string]string{
			"nonHex": "01",
		}
		err = node.SetKeyValueForAddress(addressBytes, keyValueMap)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "cannot decode key"))
	})
	t.Run("decode value failure should error", func(t *testing.T) {
		keyValueMap := map[string]string{
			"01": "nonHex",
		}
		err = node.SetKeyValueForAddress(addressBytes, keyValueMap)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "cannot decode value"))
	})
	t.Run("LoadAccount failure should error", func(t *testing.T) {
		t.Parallel()

		argsLocal := createMockArgsTestOnlyProcessingNode(t)
		nodeLocal, errLocal := NewTestOnlyProcessingNode(argsLocal)
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
					return nil, expectedErr
				},
			},
		}

		errLocal = nodeLocal.SetKeyValueForAddress(addressBytes, nil)
		require.Equal(t, expectedErr, errLocal)
	})
	t.Run("account un-castable to UserAccountHandler should error", func(t *testing.T) {
		t.Parallel()

		argsLocal := createMockArgsTestOnlyProcessingNode(t)
		nodeLocal, errLocal := NewTestOnlyProcessingNode(argsLocal)
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
					return &state.PeerAccountHandlerMock{}, nil
				},
			},
		}

		errLocal = nodeLocal.SetKeyValueForAddress(addressBytes, nil)
		require.Error(t, errLocal)
		require.Equal(t, "cannot cast AccountHandler to UserAccountHandler", errLocal.Error())
	})
	t.Run("SaveKeyValue failure should error", func(t *testing.T) {
		t.Parallel()

		nodeLocal, errLocal := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
					return &state.UserAccountStub{
						SaveKeyValueCalled: func(key []byte, value []byte) error {
							return expectedErr
						},
					}, nil
				},
			},
		}

		errLocal = nodeLocal.SetKeyValueForAddress(addressBytes, goodKeyValueMap)
		require.Equal(t, expectedErr, errLocal)
	})
	t.Run("SaveAccount failure should error", func(t *testing.T) {
		t.Parallel()

		argsLocal := createMockArgsTestOnlyProcessingNode(t)
		nodeLocal, errLocal := NewTestOnlyProcessingNode(argsLocal)
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				SaveAccountCalled: func(account vmcommon.AccountHandler) error {
					return expectedErr
				},
			},
		}

		errLocal = nodeLocal.SetKeyValueForAddress(addressBytes, goodKeyValueMap)
		require.Equal(t, expectedErr, errLocal)
	})
}

func TestTestOnlyProcessingNode_SetStateForAddress(t *testing.T) {
	t.Parallel()

	node, err := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
	require.NoError(t, err)

	address := "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj"
	scAddress := "erd1qqqqqqqqqqqqqpgqrchxzx5uu8sv3ceg8nx8cxc0gesezure5awqn46gtd"
	addressBytes, _ := node.CoreComponentsHolder.AddressPubKeyConverter().Decode(address)
	scAddressBytes, _ := node.CoreComponentsHolder.AddressPubKeyConverter().Decode(scAddress)
	addressState := &dtos.AddressState{
		Address: "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj",
		Nonce:   100,
		Balance: "1000000000000000000",
		Keys: map[string]string{
			"01": "02",
		},
	}

	t.Run("should work", func(t *testing.T) {
		_, err = node.StateComponentsHolder.AccountsAdapter().GetExistingAccount(addressBytes)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "account was not found"))

		err = node.SetStateForAddress(addressBytes, addressState)
		require.NoError(t, err)

		account, err := node.StateComponentsHolder.AccountsAdapter().GetExistingAccount(addressBytes)
		require.NoError(t, err)
		require.Equal(t, addressState.Nonce, account.GetNonce())
	})
	t.Run("LoadAccount failure should error", func(t *testing.T) {
		t.Parallel()

		nodeLocal, errLocal := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
					return nil, expectedErr
				},
			},
		}

		errLocal = nodeLocal.SetStateForAddress([]byte("address"), nil)
		require.Equal(t, expectedErr, errLocal)
	})
	t.Run("state balance invalid should error", func(t *testing.T) {
		addressStateCopy := *addressState
		addressStateCopy.Balance = "invalid balance"
		err = node.SetStateForAddress(addressBytes, &addressStateCopy)
		require.Error(t, err)
		require.Equal(t, "cannot convert string balance to *big.Int", err.Error())
	})
	t.Run("AddToBalance failure should error", func(t *testing.T) {
		t.Parallel()

		nodeLocal, errLocal := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
					return &state.UserAccountStub{
						AddToBalanceCalled: func(value *big.Int) error {
							return expectedErr
						},
					}, nil
				},
			},
		}

		errLocal = nodeLocal.SetStateForAddress([]byte("address"), addressState)
		require.Equal(t, expectedErr, errLocal)
	})
	t.Run("SaveKeyValue failure should error", func(t *testing.T) {
		t.Parallel()

		argsLocal := createMockArgsTestOnlyProcessingNode(t)
		nodeLocal, errLocal := NewTestOnlyProcessingNode(argsLocal)
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
					return &state.UserAccountStub{
						SaveKeyValueCalled: func(key []byte, value []byte) error {
							return expectedErr
						},
					}, nil
				},
			},
		}

		errLocal = nodeLocal.SetStateForAddress(addressBytes, addressState)
		require.Equal(t, expectedErr, errLocal)
	})
	t.Run("invalid sc code should error", func(t *testing.T) {
		addressStateCopy := *addressState
		addressStateCopy.Address = scAddress
		addressStateCopy.Code = "invalid code"

		err = node.SetStateForAddress(scAddressBytes, &addressStateCopy)
		require.Error(t, err)
	})
	t.Run("invalid sc code hash should error", func(t *testing.T) {
		addressStateCopy := *addressState
		addressStateCopy.Address = scAddress
		addressStateCopy.CodeHash = "invalid code hash"

		err = node.SetStateForAddress(scAddressBytes, &addressStateCopy)
		require.Error(t, err)
	})
	t.Run("invalid sc code metadata should error", func(t *testing.T) {
		addressStateCopy := *addressState
		addressStateCopy.Address = scAddress
		addressStateCopy.CodeMetadata = "invalid code metadata"

		err = node.SetStateForAddress(scAddressBytes, &addressStateCopy)
		require.Error(t, err)
	})
	t.Run("invalid sc owner should error", func(t *testing.T) {
		addressStateCopy := *addressState
		addressStateCopy.Address = scAddress
		addressStateCopy.Owner = "invalid owner"

		err = node.SetStateForAddress(scAddressBytes, &addressStateCopy)
		require.Error(t, err)
	})
	t.Run("invalid sc dev rewards should error", func(t *testing.T) {
		addressStateCopy := *addressState
		addressStateCopy.Owner = address
		addressStateCopy.Address = scAddress
		addressStateCopy.DeveloperRewards = "invalid dev rewards"

		err = node.SetStateForAddress(scAddressBytes, &addressStateCopy)
		require.Error(t, err)
	})
	t.Run("invalid root hash should error", func(t *testing.T) {
		addressStateCopy := *addressState
		addressStateCopy.Owner = address
		addressStateCopy.Address = scAddress // coverage
		addressStateCopy.DeveloperRewards = "1000000"
		addressStateCopy.RootHash = "invalid root hash"

		err = node.SetStateForAddress(scAddressBytes, &addressStateCopy)
		require.Error(t, err)
	})
	t.Run("SaveAccount failure should error", func(t *testing.T) {
		t.Parallel()

		argsLocal := createMockArgsTestOnlyProcessingNode(t)
		nodeLocal, errLocal := NewTestOnlyProcessingNode(argsLocal)
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				SaveAccountCalled: func(account vmcommon.AccountHandler) error {
					return expectedErr
				},
			},
		}

		errLocal = nodeLocal.SetStateForAddress(addressBytes, addressState)
		require.Equal(t, expectedErr, errLocal)
	})
}

func TestTestOnlyProcessingNode_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var node *testOnlyProcessingNode
	require.True(t, node.IsInterfaceNil())

	node, _ = NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
	require.False(t, node.IsInterfaceNil())
}

func TestTestOnlyProcessingNode_Close(t *testing.T) {
	t.Parallel()

	node, err := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
	require.NoError(t, err)

	require.NoError(t, node.Close())
}

func TestTestOnlyProcessingNode_Getters(t *testing.T) {
	t.Parallel()

	node := &testOnlyProcessingNode{}
	require.Nil(t, node.GetProcessComponents())
	require.Nil(t, node.GetChainHandler())
	require.Nil(t, node.GetBroadcastMessenger())
	require.Nil(t, node.GetCryptoComponents())
	require.Nil(t, node.GetCoreComponents())
	require.Nil(t, node.GetStateComponents())
	require.Nil(t, node.GetFacadeHandler())
	require.Nil(t, node.GetStatusCoreComponents())

	node, err := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
	require.Nil(t, err)

	require.NotNil(t, node.GetProcessComponents())
	require.NotNil(t, node.GetChainHandler())
	require.NotNil(t, node.GetBroadcastMessenger())
	require.NotNil(t, node.GetShardCoordinator())
	require.NotNil(t, node.GetCryptoComponents())
	require.NotNil(t, node.GetCoreComponents())
	require.NotNil(t, node.GetStateComponents())
	require.NotNil(t, node.GetFacadeHandler())
	require.NotNil(t, node.GetStatusCoreComponents())
}
