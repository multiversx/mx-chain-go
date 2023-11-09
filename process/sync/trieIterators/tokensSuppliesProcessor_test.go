package trieIterators

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	coreEsdt "github.com/multiversx/mx-chain-go/dblookupext/esdtSupply"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	chainStorage "github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func getTokensSuppliesProcessorArgs() ArgsTokensSuppliesProcessor {
	return ArgsTokensSuppliesProcessor{
		StorageService: &genericMocks.ChainStorerMock{},
		Marshaller:     &marshallerMock.MarshalizerMock{},
	}
}

func TestNewTokensSuppliesProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil storage service", func(t *testing.T) {
		t.Parallel()

		args := getTokensSuppliesProcessorArgs()
		args.StorageService = nil

		tsp, err := NewTokensSuppliesProcessor(args)
		require.Nil(t, tsp)
		require.Equal(t, errNilStorageService, err)
	})

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := getTokensSuppliesProcessorArgs()
		args.Marshaller = nil

		tsp, err := NewTokensSuppliesProcessor(args)
		require.Nil(t, tsp)
		require.Equal(t, errNilMarshaller, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getTokensSuppliesProcessorArgs()

		tsp, err := NewTokensSuppliesProcessor(args)
		require.NotNil(t, tsp)
		require.NoError(t, err)
	})
}

func TestTokensSuppliesProcessor_HandleTrieAccountIteration(t *testing.T) {
	t.Parallel()

	t.Run("nil user account", func(t *testing.T) {
		t.Parallel()

		tsp, _ := NewTokensSuppliesProcessor(getTokensSuppliesProcessorArgs())
		err := tsp.HandleTrieAccountIteration(nil)
		require.Equal(t, errNilUserAccount, err)
	})

	t.Run("should skip system account", func(t *testing.T) {
		t.Parallel()

		tsp, _ := NewTokensSuppliesProcessor(getTokensSuppliesProcessorArgs())

		userAcc := stateMock.NewAccountWrapMock(vmcommon.SystemAccountAddress)
		err := tsp.HandleTrieAccountIteration(userAcc)
		require.NoError(t, err)
	})

	t.Run("empty root hash of account", func(t *testing.T) {
		t.Parallel()

		tsp, _ := NewTokensSuppliesProcessor(getTokensSuppliesProcessorArgs())

		userAcc := stateMock.NewAccountWrapMock([]byte("addr"))
		err := tsp.HandleTrieAccountIteration(userAcc)
		require.NoError(t, err)
	})

	t.Run("root hash of account is zero only", func(t *testing.T) {
		t.Parallel()

		tsp, _ := NewTokensSuppliesProcessor(getTokensSuppliesProcessorArgs())

		userAcc := stateMock.NewAccountWrapMock([]byte("addr"))
		userAcc.SetRootHash(bytes.Repeat([]byte{0}, 32))
		err := tsp.HandleTrieAccountIteration(userAcc)
		require.NoError(t, err)
	})

	t.Run("cannot get all leaves on channel", func(t *testing.T) {
		t.Parallel()

		args := getTokensSuppliesProcessorArgs()
		tsp, _ := NewTokensSuppliesProcessor(args)

		expectedErr := errors.New("error")

		userAcc, _ := accounts.NewUserAccount([]byte("addr"), &trie.DataTrieTrackerStub{}, &trie.TrieLeafParserStub{})
		userAcc.SetRootHash([]byte("rootHash"))
		userAcc.SetDataTrie(&trie.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, leafParser common.TrieLeafParser) error {
				return expectedErr
			},
			RootCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
		})

		err := tsp.HandleTrieAccountIteration(userAcc)
		require.ErrorIs(t, err, expectedErr)
		require.Empty(t, tsp.tokensSupplies)
	})

	t.Run("should ignore non-token keys", func(t *testing.T) {
		t.Parallel()

		args := getTokensSuppliesProcessorArgs()
		tsp, _ := NewTokensSuppliesProcessor(args)

		userAcc, _ := accounts.NewUserAccount([]byte("addr"), &trie.DataTrieTrackerStub{}, &trie.TrieLeafParserStub{})
		userAcc.SetRootHash([]byte("rootHash"))
		userAcc.SetDataTrie(&trie.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, leafParser common.TrieLeafParser) error {
				leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("not a token key"), []byte("not a token value"), core.NotSpecified)

				close(leavesChannels.LeavesChan)
				return nil
			},
			RootCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
		})

		err := tsp.HandleTrieAccountIteration(userAcc)
		require.NoError(t, err)
		require.Empty(t, tsp.tokensSupplies)
	})

	t.Run("should not save tokens from the system account", func(t *testing.T) {
		t.Parallel()

		args := getTokensSuppliesProcessorArgs()
		tsp, _ := NewTokensSuppliesProcessor(args)

		userAcc, _ := accounts.NewUserAccount(vmcommon.SystemAccountAddress, &trie.DataTrieTrackerStub{}, &trie.TrieLeafParserStub{})
		userAcc.SetRootHash([]byte("rootHash"))
		userAcc.SetDataTrie(&trie.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, leafParser common.TrieLeafParser) error {
				esToken := &esdt.ESDigitalToken{
					Value: big.NewInt(37),
				}
				esBytes, _ := args.Marshaller.Marshal(esToken)
				tknKey := []byte("ELRONDesdtTKN-00aacc")
				value := append(esBytes, tknKey...)
				value = append(value, []byte("addr")...)
				leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage(tknKey, value, core.NotSpecified)

				close(leavesChannels.LeavesChan)
				return nil
			},
		})

		err := tsp.HandleTrieAccountIteration(userAcc)
		require.NoError(t, err)
		require.Empty(t, tsp.tokensSupplies)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getTokensSuppliesProcessorArgs()
		tsp, _ := NewTokensSuppliesProcessor(args)

		dtt, _ := trackableDataTrie.NewTrackableDataTrie([]byte("addr"), &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		dtlp, _ := parsers.NewDataTrieLeafParser([]byte("addr"), &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		userAcc, _ := accounts.NewUserAccount([]byte("addr"), dtt, dtlp)
		userAcc.SetRootHash([]byte("rootHash"))
		userAcc.SetDataTrie(&trie.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, leafParser common.TrieLeafParser) error {
				esToken := &esdt.ESDigitalToken{
					Value: big.NewInt(37),
				}
				esBytes, _ := args.Marshaller.Marshal(esToken)
				tknKey := []byte("ELRONDesdtTKN-00aacc")
				value := append(esBytes, tknKey...)
				value = append(value, []byte("addr")...)
				leaf, err := leafParser.ParseLeaf(tknKey, value, 0)
				require.Nil(t, err)
				leavesChannels.LeavesChan <- leaf

				sft := &esdt.ESDigitalToken{
					Value: big.NewInt(1),
				}
				sftBytes, _ := args.Marshaller.Marshal(sft)
				sftKey := []byte("ELRONDesdtSFT-00aabb")
				sftKey = append(sftKey, big.NewInt(37).Bytes()...)
				value = append(sftBytes, sftKey...)
				value = append(value, []byte("addr")...)
				leaf, err = leafParser.ParseLeaf(sftKey, value, 0)
				require.Nil(t, err)
				leavesChannels.LeavesChan <- leaf

				close(leavesChannels.LeavesChan)
				return nil
			},
			RootCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
		})

		err := tsp.HandleTrieAccountIteration(userAcc)
		require.NoError(t, err)

		err = tsp.HandleTrieAccountIteration(userAcc)
		require.NoError(t, err)

		expectedSupplies := map[string]*big.Int{
			"SFT-00aabb-25": big.NewInt(2),
			"SFT-00aabb":    big.NewInt(2),
			"TKN-00aacc":    big.NewInt(74),
		}
		require.Equal(t, expectedSupplies, tsp.tokensSupplies)
	})
}

func TestTokensSuppliesProcessor_SaveSupplies(t *testing.T) {
	t.Parallel()

	t.Run("cannot find esdt supplies storer", func(t *testing.T) {
		t.Parallel()

		errStorerNotFound := errors.New("storer not found")
		args := getTokensSuppliesProcessorArgs()
		args.StorageService = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (chainStorage.Storer, error) {
				return nil, errStorerNotFound
			},
		}
		tsp, _ := NewTokensSuppliesProcessor(args)
		err := tsp.SaveSupplies()
		require.Equal(t, errStorerNotFound, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		savedItems := make(map[string][]byte)
		args := getTokensSuppliesProcessorArgs()
		args.StorageService = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (chainStorage.Storer, error) {
				return &storage.StorerStub{
					PutCalled: func(key, data []byte) error {
						savedItems[string(key)] = data
						return nil
					},
				}, nil
			},
		}
		tsp, _ := NewTokensSuppliesProcessor(args)

		supplies := map[string]*big.Int{
			"SFT-00aabb-37": big.NewInt(2),
			"SFT-00aabb":    big.NewInt(2),
			"TKN-00aacc":    big.NewInt(74),
		}
		tsp.tokensSupplies = supplies

		err := tsp.SaveSupplies()
		require.NoError(t, err)

		checkStoredSupply := func(t *testing.T, key string, storedValue []byte, expectedSupply *big.Int) {
			supply := coreEsdt.SupplyESDT{}
			_ = args.Marshaller.Unmarshal(&supply, storedValue)
			require.Equal(t, expectedSupply, supply.Supply)
		}

		require.Len(t, savedItems, 3)
		for key, value := range savedItems {
			checkStoredSupply(t, key, value, supplies[key])
		}
	})
}
