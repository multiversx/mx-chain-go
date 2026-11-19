package components

import (
	"bytes"
	"encoding/base64"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/heartbeat"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	factorymock "github.com/multiversx/mx-chain-go/node/mock/factory"
	accountsState "github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	"github.com/multiversx/mx-chain-go/testscommon/state"
)

var expectedErr = errors.New("expected error")

func createMockArgsTestOnlyProcessingNode(t *testing.T) ArgsTestOnlyProcessingNode {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:                    3,
		OriginalConfigsPath:            "../../../cmd/node/config/",
		RoundDurationInMillis:          6000,
		SupernovaRoundDurationInMillis: 600,
		TempDir:                        t.TempDir(),
		MinNodesPerShard:               1,
		MetaChainMinNodes:              1,
		ConsensusGroupSize:             1,
		MetaChainConsensusGroupSize:    1,
	})
	require.Nil(t, err)

	return ArgsTestOnlyProcessingNode{
		Configs:             outputConfigs.Configs,
		GasScheduleFilename: outputConfigs.GasScheduleFilename,
		NumShards:           3,

		SyncedBroadcastNetwork:      NewSyncedBroadcastNetwork(),
		ChanStopNodeProcess:         make(chan endProcess.ArgEndProcess),
		APIInterface:                api.NewNoApiInterface(),
		Monitor:                     heartbeat.NewHeartbeatMonitor(),
		ShardIDStr:                  "0",
		RoundDurationInMillis:       6000,
		MinNodesMeta:                1,
		MinNodesPerShard:            1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	}
}

func TestNewTestOnlyProcessingNode(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("should work", func(t *testing.T) {
		args := createMockArgsTestOnlyProcessingNode(t)
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)
	})
	t.Run("try commit a block", func(t *testing.T) {
		args := createMockArgsTestOnlyProcessingNode(t)
		genesisTime := time.Now()
		args.GenesisTime = genesisTime
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)

		newHeader, err := node.ProcessComponentsHolder.BlockProcessor().CreateNewHeader(1, 1)
		assert.Nil(t, err)

		rootHash, err := node.GetStateComponents().AccountsAdapter().RootHash()
		if err != nil {
			log.Error("node.GetStateComponents().AccountsAdapter().RootHash()", "err", err)
		}

		genesisHeader := node.GetDataComponents().Blockchain().GetGenesisHeader()
		err = genesisHeader.SetRootHash(rootHash)
		require.Nil(t, err)

		err = node.GetDataComponents().Blockchain().SetGenesisHeader(genesisHeader)
		require.Nil(t, err)

		err = node.GetDataComponents().Datapool().Transactions().OnExecutedBlock(genesisHeader, rootHash)
		require.Nil(t, err)

		err = newHeader.SetPrevHash(node.ChainHandler.GetGenesisHeaderHash())
		assert.Nil(t, err)

		// Set timestamp before processing (pre-supernova: header stores seconds)
		expectedTimestampMs := node.GetCoreComponents().RoundHandler().GetTimeStampForRound(1)
		err = newHeader.SetTimeStamp(expectedTimestampMs / 1000)
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
		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.GeneralConfig.Marshalizer.Type = "invalid type"
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateCryptoComponents failure should error", func(t *testing.T) {
		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.GeneralConfig.PublicKeyPIDSignature.Type = "invalid type"
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateNetworkComponents failure should error", func(t *testing.T) {
		args := createMockArgsTestOnlyProcessingNode(t)
		args.SyncedBroadcastNetwork = nil
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateBootstrapComponents failure should error", func(t *testing.T) {
		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.FlagsConfig.WorkingDir = ""
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateStateComponents failure should error", func(t *testing.T) {
		args := createMockArgsTestOnlyProcessingNode(t)
		args.ShardIDStr = common.MetachainShardName // coverage only
		args.Configs.GeneralConfig.TrieStorageManagerConfig = config.TrieStorageManagerConfig{}
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("CreateProcessComponents failure should error", func(t *testing.T) {
		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.FlagsConfig.Version = ""
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
	t.Run("createFacade failure should error", func(t *testing.T) {
		args := createMockArgsTestOnlyProcessingNode(t)
		args.Configs.EpochConfig.GasSchedule.GasScheduleByEpochs = nil
		node, err := NewTestOnlyProcessingNode(args)
		require.Error(t, err)
		require.Nil(t, node)
	})
}

func TestTestOnlyProcessingNode_SetKeyValueForAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	node, err := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
	require.NoError(t, err)
	nonce := uint64(100)

	address := "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj"
	scAddress := "erd1qqqqqqqqqqqqqpgqrchxzx5uu8sv3ceg8nx8cxc0gesezure5awqn46gtd"
	addressBytes, _ := node.CoreComponentsHolder.AddressPubKeyConverter().Decode(address)
	scAddressBytes, _ := node.CoreComponentsHolder.AddressPubKeyConverter().Decode(scAddress)
	addressState := &dtos.AddressState{
		Address: "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj",
		Nonce:   &nonce,
		Balance: "1000000000000000000",
		Pairs: map[string]string{
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
		require.Equal(t, *addressState.Nonce, account.GetNonce())
	})
	t.Run("LoadAccount failure should error", func(t *testing.T) {
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
		nodeLocal, errLocal := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
		require.NoError(t, errLocal)

		nodeLocal.StateComponentsHolder = &factory.StateComponentsMock{
			Accounts: &state.AccountsStub{
				LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
					return &state.UserAccountStub{
						AddToBalanceCalled: func(value *big.Int) error {
							return expectedErr
						},
						Balance: big.NewInt(0),
					}, nil
				},
			},
		}

		errLocal = nodeLocal.SetStateForAddress([]byte("address"), addressState)
		require.Equal(t, expectedErr, errLocal)
	})
	t.Run("SaveKeyValue failure should error", func(t *testing.T) {
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
						Balance: big.NewInt(0),
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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var node *testOnlyProcessingNode
	require.True(t, node.IsInterfaceNil())

	node, _ = NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
	require.False(t, node.IsInterfaceNil())
}

func TestTestOnlyProcessingNode_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	node, err := NewTestOnlyProcessingNode(createMockArgsTestOnlyProcessingNode(t))
	require.NoError(t, err)

	require.NoError(t, node.Close())
}

func TestTestOnlyProcessingNode_Getters(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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

func TestTestOnlyProcessingNode_SetBlockchainRootHashWithMetaExecutionResult(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := createMockArgsTestOnlyProcessingNode(t)
	args.Configs.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)] = config.ActivationRoundByName{
		Round:   "0",
		Options: []string{},
	}
	node, err := NewTestOnlyProcessingNode(args)
	require.NoError(t, err)

	expectedRootHash := []byte("new-root-hash")

	metaResult := &block.MetaExecutionResult{
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("header-hash"),
				HeaderNonce: 1,
				HeaderRound: 1,
				RootHash:    []byte("initial-root-hash"),
			},
		},
	}

	var capturedResult data.BaseExecutionResultHandler
	chainHandlerStub := &testscommon.ChainHandlerStub{
		GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
			return nil
		},
		GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
			return 0, nil, nil
		},
		SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, blockHash []byte, rootHash []byte) {
		},
		GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
			return metaResult
		},
		SetLastExecutionInfoCalled: func(header data.HeaderHandler, result data.BaseExecutionResultHandler) {
			capturedResult = result
		},
	}
	node.ChainHandler = chainHandlerStub

	node.setBlockchainRootHashIfSupernovaIsActive(expectedRootHash)

	require.NotNil(t, capturedResult)
	capturedMetaResult, ok := capturedResult.(*block.MetaExecutionResult)
	require.True(t, ok)
	require.Equal(t, expectedRootHash, capturedMetaResult.ExecutionResult.BaseExecutionResult.RootHash)
}

func TestTestOnlyProcessingNode_SetBlockchainRootHashWithShardExecutionResult(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := createMockArgsTestOnlyProcessingNode(t)
	args.Configs.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)] = config.ActivationRoundByName{
		Round:   "0",
		Options: []string{},
	}
	node, err := NewTestOnlyProcessingNode(args)
	require.NoError(t, err)

	expectedRootHash := []byte("new-root-hash")

	shardResult := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("header-hash"),
			HeaderNonce: 1,
			HeaderRound: 1,
			RootHash:    []byte("initial-root-hash"),
		},
	}

	var capturedResult data.BaseExecutionResultHandler
	chainHandlerStub := &testscommon.ChainHandlerStub{
		GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
			return nil
		},
		GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
			return 0, nil, nil
		},
		SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, blockHash []byte, rootHash []byte) {
		},
		GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
			return shardResult
		},
		SetLastExecutionInfoCalled: func(header data.HeaderHandler, result data.BaseExecutionResultHandler) {
			capturedResult = result
		},
	}
	node.ChainHandler = chainHandlerStub

	node.setBlockchainRootHashIfSupernovaIsActive(expectedRootHash)

	require.NotNil(t, capturedResult)
	capturedShardResult, ok := capturedResult.(*block.ExecutionResult)
	require.True(t, ok)
	require.Equal(t, expectedRootHash, capturedShardResult.BaseExecutionResult.RootHash)
}

func newNodeForSetStateRootHashTest(t *testing.T) (*testOnlyProcessingNode, *accountsState.AccountsDB) {
	t.Helper()
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.FixDataTrieLeafValidationFlag)
	accounts := integrationtests.CreateAccountsDB(integrationtests.CreateMemUnit(), flags)
	t.Cleanup(func() {
		require.NoError(t, accounts.Close())
	})
	node := &testOnlyProcessingNode{
		StateComponentsHolder: &factory.StateComponentsMock{Accounts: accounts},
		CoreComponentsHolder: &factorymock.CoreComponentsMock{
			EnableRoundsHandlerField: testscommon.NewEnableRoundsHandlerStub(),
		},
	}
	return node, accounts
}

func setStateRootHashTestContractAddress() []byte {
	address := bytes.Repeat([]byte{0x07}, 32)
	clear(address[:8])
	address[8], address[9] = 0x05, 0x00
	return address
}

func TestSetStateForAddress_InvalidRootDoesNotCreateAccount(t *testing.T) {
	tests := []struct {
		name     string
		rootHash string
		wantErr  string
	}{
		{
			name:     "malformed",
			rootHash: "invalid root hash",
			wantErr:  "illegal base64 data",
		},
		{
			name:     "mismatch",
			rootHash: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0xaa}, 32)),
			wantErr:  "account root hash does not match",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node, accounts := newNodeForSetStateRootHashTest(t)
			address := setStateRootHashTestContractAddress()
			before, err := accounts.RootHash()
			require.NoError(t, err)

			nonce := uint64(9)
			err = node.SetStateForAddress(address, &dtos.AddressState{
				Nonce:    &nonce,
				Balance:  "123",
				Code:     "abcdef",
				Pairs:    map[string]string{"01": "02"},
				RootHash: test.rootHash,
			})
			require.ErrorContains(t, err, test.wantErr)

			after, err := accounts.RootHash()
			require.NoError(t, err)
			require.Equal(t, before, after)
			require.Zero(t, accounts.JournalLen())

			unrelated := bytes.Repeat([]byte{0x02}, 32)
			require.NoError(t, node.SetStateForAddress(unrelated, &dtos.AddressState{Balance: "1"}))
			committed, err := accounts.RootHash()
			require.NoError(t, err)
			require.NoError(t, accounts.RecreateTrie(holders.NewDefaultRootHashesHolder(committed)))
			_, err = accounts.GetExistingAccount(address)
			require.ErrorIs(t, err, accountsState.ErrAccNotFound)
			codeHash := integrationtests.TestHasher.Compute(string([]byte{0xab, 0xcd, 0xef}))
			require.Empty(t, accounts.GetCode(codeHash))
		})
	}
}

func TestSetStateForAddress_RootMismatchPreservesPendingWrites(t *testing.T) {
	node, accounts := newNodeForSetStateRootHashTest(t)
	address := setStateRootHashTestContractAddress()
	require.NoError(t, node.SetStateForAddress(address, &dtos.AddressState{
		Balance: "5",
		Code:    "abcdef",
		Pairs:   map[string]string{"01": "02"},
	}))

	pendingAddress := bytes.Repeat([]byte{0x03}, 32)
	pending, err := node.getUserAccount(pendingAddress)
	require.NoError(t, err)
	require.NoError(t, pending.AddToBalance(big.NewInt(4)))
	require.NoError(t, pending.SaveKeyValue([]byte{0x01}, []byte{0x05}))
	require.NoError(t, accounts.SaveAccount(pending))
	snapshot := accounts.JournalLen()
	require.Positive(t, snapshot)
	before, err := accounts.RootHash()
	require.NoError(t, err)

	nonce := uint64(9)
	err = node.SetStateForAddress(address, &dtos.AddressState{
		Nonce:    &nonce,
		Balance:  "123",
		Code:     "fedcba",
		Pairs:    map[string]string{"01": "03", "02": "04"},
		RootHash: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0xaa}, 32)),
	})
	require.ErrorContains(t, err, "account root hash does not match")
	require.Equal(t, snapshot, accounts.JournalLen())
	after, err := accounts.RootHash()
	require.NoError(t, err)
	require.Equal(t, before, after)

	committed, err := accounts.Commit()
	require.NoError(t, err)
	require.NoError(t, accounts.RecreateTrie(holders.NewDefaultRootHashesHolder(committed)))
	existing, err := accounts.GetExistingAccount(address)
	require.NoError(t, err)
	contract := existing.(accountsState.UserAccountHandler)
	require.Zero(t, contract.GetNonce())
	require.Equal(t, big.NewInt(5), contract.GetBalance())
	require.Equal(t, []byte{0xab, 0xcd, 0xef}, accounts.GetCode(contract.GetCodeHash()))
	value, _, err := contract.RetrieveValue([]byte{0x01})
	require.NoError(t, err)
	require.Equal(t, []byte{0x02}, value)

	existing, err = accounts.GetExistingAccount(pendingAddress)
	require.NoError(t, err)
	persisted := existing.(accountsState.UserAccountHandler)
	require.Equal(t, big.NewInt(4), persisted.GetBalance())
	value, _, err = persisted.RetrieveValue([]byte{0x01})
	require.NoError(t, err)
	require.Equal(t, []byte{0x05}, value)
}

func TestSetStateForAddress_MatchingRootCommits(t *testing.T) {
	node, accounts := newNodeForSetStateRootHashTest(t)
	address := bytes.Repeat([]byte{0x06}, 32)
	require.NoError(t, node.SetStateForAddress(address, &dtos.AddressState{
		Balance: "1",
		Pairs:   map[string]string{"01": "02"},
	}))

	existing, err := accounts.GetExistingAccount(address)
	require.NoError(t, err)
	root := existing.(accountsState.UserAccountHandler).GetRootHash()
	require.NotEmpty(t, root)
	require.NoError(t, node.SetStateForAddress(address, &dtos.AddressState{
		Balance:  "2",
		RootHash: base64.StdEncoding.EncodeToString(root),
	}))
	require.Zero(t, accounts.JournalLen())

	committed, err := accounts.RootHash()
	require.NoError(t, err)
	require.NoError(t, accounts.RecreateTrie(holders.NewDefaultRootHashesHolder(committed)))
	existing, err = accounts.GetExistingAccount(address)
	require.NoError(t, err)
	account := existing.(accountsState.UserAccountHandler)
	require.Equal(t, big.NewInt(2), account.GetBalance())
	require.Equal(t, root, account.GetRootHash())
}
