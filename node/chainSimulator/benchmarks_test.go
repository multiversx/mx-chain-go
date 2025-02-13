package chainSimulator

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"math/big"
	_ "net/http/pprof" // Import pprof
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/stretchr/testify/require"
)

func TestChainSimulatorWithMainnetDBSwaps(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    20,
		},
		ApiInterface:      api.NewNoApiInterface(),
		MinNodesPerShard:  1,
		MetaChainMinNodes: 1,
		InitialRound:      22693505,
		InitialEpoch:      1575,
		InitialNonce:      22675196,
		AlterConfigsFunction: func(cfg *config.Configs) {
			// we need to enable this as this test skips a lot of epoch activations events, and it will fail otherwise
			// because the owner of a BLS key coming from genesis is not set
			// (the owner is not set at genesis anymore because we do not enable the staking v2 in that phase)
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
		},
		TrieStoragePaths: map[string]components.TriePathAndRootHash{
			"1": {
				TriePath: "",
				RootHash: "",
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	rootHashBytes, errD := hex.DecodeString("0565601c40d4285d470a1f03d79475b83bee4beab9c1dbaccb20cfcd6ba21899")
	require.Nil(t, errD)

	err = chainSimulator.GetNodeHandler(1).GetStateComponents().AccountsAdapter().RecreateTrie(holders.NewDefaultRootHashesHolder(rootHashBytes))
	require.Nil(t, err)

	time.Sleep(time.Second)
	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	txs := generateSwapsTransactions(t, chainSimulator, 100_000)
	executeTransaction(t, chainSimulator, txs, "swapsRetrieveValueCacheNewMarshallSliceGrow")
}

func TestChainSimulatorWithMainnetDB(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    20,
		},
		ApiInterface:      api.NewNoApiInterface(),
		MinNodesPerShard:  1,
		MetaChainMinNodes: 1,
		InitialRound:      22693505,
		InitialEpoch:      1575,
		InitialNonce:      22675196,
		AlterConfigsFunction: func(cfg *config.Configs) {
			// we need to enable this as this test skips a lot of epoch activations events, and it will fail otherwise
			// because the owner of a BLS key coming from genesis is not set
			// (the owner is not set at genesis anymore because we do not enable the staking v2 in that phase)
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
		},
		TrieStoragePaths: map[string]components.TriePathAndRootHash{
			"1": {
				TriePath: "",
				RootHash: "",
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	time.Sleep(time.Second)
	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	t.Run("move balances with no accounts", func(t *testing.T) {
		txs := generateMoveBalance(t, "erd1vj3efd5czwearu0gr3vjct8ef53lvtl7vs42vts2kh2qn3cucrnsj7ymqx", chainSimulator, 100_000, false)
		executeTransaction(t, chainSimulator, txs, "moveBalanceNoAccountsNoCache")
	})
	t.Run("move balances with with existing accounts", func(t *testing.T) {
		txs := generateMoveBalance(t, "erd1vj3efd5czwearu0gr3vjct8ef53lvtl7vs42vts2kh2qn3cucrnsj7ymqx", chainSimulator, 100_000, true)
		executeTransaction(t, chainSimulator, txs, "moveBalanceWithExistingAccounts")
	})

	t.Run("esdt transfer no accounts", func(t *testing.T) {
		txs := generateESDTTransfer(t, "erd1ty4pvmjtl3mnsjvnsxgcpedd08fsn83f05tu0v5j23wnfce9p86snlkdyy", chainSimulator, 100_000, false, "ZPAY-247875")
		executeTransaction(t, chainSimulator, txs, "esdtTransferNoAccountsWithCache")
	})

	t.Run("esdt transfer with existing accounts", func(t *testing.T) {
		txs := generateESDTTransfer(t, "erd1ty4pvmjtl3mnsjvnsxgcpedd08fsn83f05tu0v5j23wnfce9p86snlkdyy", chainSimulator, 100_000, true, "ZPAY-247875")
		executeTransaction(t, chainSimulator, txs, "esdtTransferWithExistingAccounts")
	})
}

func executeTransaction(t *testing.T, cs *simulator, txs []tupleTx, profileFileName string) {
	f, err := os.Create(fmt.Sprintf("pprof/%s.prof", profileFileName))
	if err != nil {
		t.Fatalf("Could not create CPU profile: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	err = pprof.StartCPUProfile(f)
	if err != nil {
		t.Fatalf("Could not start CPU profile: %v", err)
	}

	defer pprof.StopCPUProfile()

	goRoutinesFileName := fmt.Sprintf("pprof/%s-goroutines", profileFileName)
	goRoutinesFile, err := os.Create(fmt.Sprintf("%s.prof", goRoutinesFileName))
	if err != nil {
		t.Fatalf("could not create Goroutine profile file: %v", err)

	}
	defer func() {
		_ = goRoutinesFile.Close()
	}()

	err = pprof.Lookup("goroutine").WriteTo(goRoutinesFile, 0)
	if err != nil {
		t.Fatalf("could not write Goroutine profile: %v", err)
	}
	shard1NodeHandler := cs.GetNodeHandler(1)
	for _, txData := range txs {
		shard1NodeHandler.GetProcessComponents().ScheduledTxsExecutionHandler().AddScheduledTx(txData.hash, txData.tx)
	}

	startTime := time.Now()
	err = shard1NodeHandler.GetProcessComponents().ScheduledTxsExecutionHandler().ExecuteAll(func() time.Duration {
		return 1000 * time.Second
	})
	require.Nil(t, err)
	log.Warn(fmt.Sprintf("execution of %d", len(txs)), "took", time.Since(startTime))

	err = profileMemory(profileFileName)
	require.Nil(t, err)
}

type tupleTx struct {
	tx   *transaction.Transaction
	hash []byte
}

func generateESDTTransfer(t *testing.T, initialAddress string, cs *simulator, numTxs uint64, existingAccounts bool, tokenIdentifier string) []tupleTx {
	dataField := fmt.Sprintf("%s@%s@%s", core.BuiltInFunctionESDTTransfer, hex.EncodeToString([]byte(tokenIdentifier)), hex.EncodeToString(big.NewInt(1).Bytes()))

	return generateTransactions(t, initialAddress, cs, numTxs, existingAccounts, []byte(dataField), 500_000)
}

func generateMoveBalance(t *testing.T, initialAddr string, s *simulator, numTxs uint64, existingAccounts bool) []tupleTx {
	return generateTransactions(t, initialAddr, s, numTxs, existingAccounts, nil, 50_000)
}

func generateTransactions(t *testing.T, initialAddr string, s *simulator, numTxs uint64, existingAccounts bool, dataField []byte, gasLimit uint64) []tupleTx {
	account, _, err := s.GetNodeHandler(1).GetFacadeHandler().GetAccount(initialAddr, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)

	sndAddrBytes, err := s.GetNodeHandler(1).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddr)
	require.Nil(t, err)

	shard1NodeHandler := s.GetNodeHandler(1)

	txs := make([]tupleTx, 0, numTxs)
	for i := uint64(0); i < numTxs; i++ {
		var rcv []byte
		if existingAccounts {
			// will generate a new address and also the account is created
			rcvData, errC := s.GenerateAndMintWalletAddress(1, big.NewInt(0))
			require.Nil(t, errC)
			rcv = rcvData.Bytes

		} else {
			// will generate a new address ( account is not created
			rcv = generateAddressInShard(shard1NodeHandler.GetShardCoordinator(), 32)
		}

		value := big.NewInt(1)
		if bytes.HasPrefix(dataField, []byte(core.BuiltInFunctionESDTTransfer)) {
			value = big.NewInt(0)
		}

		tx := &transaction.Transaction{
			Nonce:     account.Nonce + i,
			Value:     value,
			RcvAddr:   rcv,
			SndAddr:   sndAddrBytes,
			GasPrice:  1_000_000_000,
			GasLimit:  gasLimit,
			Data:      dataField,
			ChainID:   []byte(configs.ChainID),
			Version:   1,
			Signature: []byte("dummy"),
		}

		txHash, errG := core.CalculateHash(shard1NodeHandler.GetCoreComponents().InternalMarshalizer(), shard1NodeHandler.GetCoreComponents().Hasher(), tx)
		require.Nil(t, errG)

		txs = append(txs, tupleTx{tx: tx, hash: txHash})
	}
	return txs
}

func generateAddressesWithATokenAndEGLD(t *testing.T, cs *simulator, shardID uint32, numAddresses int, key, value string) [][]byte {
	shardCoordinatorForShard := cs.GetNodeHandler(shardID).GetShardCoordinator()
	converter := cs.GetNodeHandler(shardID).GetCoreComponents().AddressPubKeyConverter()

	addresses := make([][]byte, 0, numAddresses)
	addressesState := make([]*dtos.AddressState, 0, numAddresses)
	for i := 0; i < numAddresses; i++ {
		addr := generateAddressInShard(shardCoordinatorForShard, 32)
		addrBech, _ := converter.Encode(addr)

		addressesState = append(addressesState, &dtos.AddressState{
			Address: addrBech,
			Balance: "1000000000000000000",
			Pairs: map[string]string{
				key: value,
			},
		})

		addresses = append(addresses, addr)
	}

	err := cs.SetStateMultiple(addressesState)
	require.Nil(t, err)

	return addresses
}

func generateSwapsTransactions(t *testing.T, cs *simulator, numTxs int) []tupleTx {
	senderAddresses := generateAddressesWithATokenAndEGLD(t, cs, 1, numTxs, "454c524f4e446573647448544d2d663531643535", "120c000125d0869555f01c300000")
	shard1NodeHandler := cs.GetNodeHandler(1)

	err := cs.GenerateBlocks(1)
	require.Nil(t, err)

	rcv, err := cs.GetNodeHandler(1).GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqpgqxhgs55hpdqll93nnvf0nwnt3wmh62u692jps5wm8uj")
	require.Nil(t, err)

	txs := make([]tupleTx, 0, numTxs)
	for i := 0; i < numTxs; i++ {
		tx := &transaction.Transaction{
			Nonce:     0,
			Value:     big.NewInt(0),
			RcvAddr:   rcv,
			SndAddr:   senderAddresses[i],
			GasPrice:  1_000_000_000,
			GasLimit:  100_000_000,
			Data:      []byte("ESDTTransfer@48544D2D663531643535@01D460162F516F0000@73776170546F6B656E734669786564496E707574@5745474C442D626434643739@01"),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
			Signature: []byte("dummy"),
		}

		txHash, errG := core.CalculateHash(shard1NodeHandler.GetCoreComponents().InternalMarshalizer(), shard1NodeHandler.GetCoreComponents().Hasher(), tx)
		require.Nil(t, errG)

		txs = append(txs, tupleTx{tx: tx, hash: txHash})
	}
	return txs
}

func profileMemory(profileFileName string) error {
	runtime.GC()

	f, err := os.Create(fmt.Sprintf("pprof/%s-memory.prof", profileFileName))
	if err != nil {
		return fmt.Errorf("could not create memory profile file: %w", err)
	}
	defer f.Close()

	err = pprof.Lookup("heap").WriteTo(f, 0)
	if err != nil {
		return fmt.Errorf("could not write memory profile: %w", err)
	}

	return nil
}
