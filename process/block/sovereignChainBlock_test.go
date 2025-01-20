package block_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	atomicCore "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	sovereignCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
)

func createSovereignChainShardTrackerMockArguments() track.ArgShardTracker {
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      &hashingMocks.HasherMock{},
		Marshalizer: &marshallerMock.MarshalizerMock{},
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgShardTracker{
		ArgBaseTracker: track.ArgBaseTracker{
			Hasher:           &hashingMocks.HasherMock{},
			HeaderValidator:  headerValidator,
			Marshalizer:      &marshallerMock.MarshalizerStub{},
			RequestHandler:   &testscommon.ExtendedShardHeaderRequestHandlerStub{},
			RoundHandler:     &testscommon.RoundHandlerMock{},
			ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
			Store:            &storage.ChainStorerStub{},
			StartHeaders:     createGenesisBlocks(&testscommon.ShardsCoordinatorMock{NoShards: 1}),
			PoolsHolder:      dataRetrieverMock.NewPoolsHolderMock(),
			WhitelistHandler: &testscommon.WhiteListHandlerStub{},
			FeeHandler:       &economicsmocks.EconomicsHandlerStub{},
		},
	}

	return arguments
}

func createShardBlockProcessorArgsForSovereign(
	coreComp *mock.CoreComponentsMock,
	dataComp *mock.DataComponentsMock,
	bootstrapComp *mock.BootstrapComponentsMock,
	statusComp *mock.StatusComponentsMock,
) blproc.ArgShardProcessor {
	shardArguments := createSovereignChainShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	rrh, _ := requestHandlers.NewResolverRequestHandler(
		&dataRetrieverMock.RequestersFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	arguments := CreateMockArguments(coreComp, dataComp, bootstrapComp, statusComp)
	arguments.BootstrapComponents = &mock.BootstrapComponentsMock{
		Coordinator:          mock.NewOneShardCoordinatorMock(),
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32) data.HeaderHandler {
				return &block.SovereignChainHeader{Header: &block.Header{}}
			},
		},
	}
	arguments.AccountsDB[state.PeerAccountsState] = &stateMock.AccountsStub{
		CommitCalled: func() ([]byte, error) {
			return nil, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return nil, nil
		},
	}

	arguments.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
	arguments.RequestHandler, _ = requestHandlers.NewSovereignResolverRequestHandler(rrh)

	return arguments
}

func createArgsSovereignChainBlockProcessor(baseArgs blproc.ArgShardProcessor) blproc.ArgsSovereignChainBlockProcessor {
	sp, _ := blproc.NewShardProcessor(baseArgs)
	return blproc.ArgsSovereignChainBlockProcessor{
		ShardProcessor:                  sp,
		ValidatorStatisticsProcessor:    &testscommon.ValidatorStatisticsProcessorStub{},
		OutgoingOperationsFormatter:     &sovereign.OutgoingOperationsFormatterMock{},
		OutGoingOperationsPool:          &sovereign.OutGoingOperationsPoolMock{},
		OperationsHasher:                &testscommon.HasherStub{},
		EpochStartDataCreator:           &mock.EpochStartDataCreatorStub{},
		EpochRewardsCreator:             &testscommon.RewardsCreatorStub{},
		ValidatorInfoCreator:            &testscommon.EpochValidatorInfoCreatorStub{},
		EpochSystemSCProcessor:          &testscommon.EpochStartSystemSCStub{},
		EpochEconomics:                  &mock.EpochEconomicsStub{},
		SCToProtocol:                    &mock.SCToProtocolStub{},
		MainChainNotarizationStartRound: 11,
	}
}

func createSovChainBlockProcessorArgs() blproc.ArgsSovereignChainBlockProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	return createArgsSovereignChainBlockProcessor(arguments)
}

func TestSovereignBlockProcessor_NewSovereignChainBlockProcessorShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should error when shard processor is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.ShardProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrNilBlockProcessor)
	})

	t.Run("should error when validator statistics is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.ValidatorStatisticsProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrNilValidatorStatistics)
	})

	t.Run("should error when outgoing operations formatter is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.OutgoingOperationsFormatter = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, errors.ErrNilOutgoingOperationsFormatter)
	})

	t.Run("should error when outgoing operation pool is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.OutGoingOperationsPool = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, errors.ErrNilOutGoingOperationsPool, err)
	})

	t.Run("should error when operations hasher is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.OperationsHasher = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, errors.ErrNilOperationsHasher, err)
	})

	t.Run("should error when epoch start data creator is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.EpochStartDataCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartDataCreator, err)
	})

	t.Run("should error when epoch rewards creator is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.EpochRewardsCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilRewardsCreator, err)
	})

	t.Run("should error when epoch validator infor creator is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.ValidatorInfoCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartValidatorInfoCreator, err)
	})

	t.Run("should error when epoch start system sc processor is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.EpochSystemSCProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartSystemSCProcessor, err)
	})

	t.Run("should error when epoch economics is nil", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.EpochEconomics = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochEconomics, err)
	})

	t.Run("should error when type assertion to extendedShardHeaderTrackHandler fails", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(args)

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.ShardProcessor = sp
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should error when type assertion to extendedShardHeaderRequestHandler fails", func(t *testing.T) {
		t.Parallel()

		shardArguments := createSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		args := CreateMockArguments(createComponentHolderMocks())
		args.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
		sp, _ := blproc.NewShardProcessor(args)

		sovArgs := createSovChainBlockProcessorArgs()
		sovArgs.ShardProcessor = sp
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		sovArgs := createSovChainBlockProcessorArgs()
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)

		require.NotNil(t, scbp)
		require.Nil(t, err)
	})
}

func TestSovereignChainBlockProcessor_createAndSetOutGoingMiniBlock(t *testing.T) {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	expectedLogs := []*data.LogData{
		{
			TxHash: "txHash1",
		},
	}
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		GetAllCurrentLogsCalled: func() []*data.LogData {
			return expectedLogs
		},
	}
	bridgeOp1 := []byte("bridgeOp@123@rcv1@token1@val1")
	bridgeOp2 := []byte("bridgeOp@124@rcv2@token2@val2")

	outgoingOpsHasher := &mock.HasherStub{}
	bridgeOp1Hash := outgoingOpsHasher.Compute(string(bridgeOp1))
	bridgeOp2Hash := outgoingOpsHasher.Compute(string(bridgeOp2))
	bridgeOpsHash := outgoingOpsHasher.Compute(string(append(bridgeOp1Hash, bridgeOp2Hash...)))

	outgoingOperationsFormatter := &sovereign.OutgoingOperationsFormatterMock{
		CreateOutgoingTxDataCalled: func(logs []*data.LogData) ([][]byte, error) {
			require.Equal(t, expectedLogs, logs)
			return [][]byte{bridgeOp1, bridgeOp2}, nil
		},
	}

	poolAddCt := 0
	outGoingOperationsPool := &sovereign.OutGoingOperationsPoolMock{
		AddCalled: func(data *sovereignCore.BridgeOutGoingData) {
			defer func() {
				poolAddCt++
			}()

			switch poolAddCt {
			case 0:
				require.Equal(t, &sovereignCore.BridgeOutGoingData{
					Hash: bridgeOpsHash,
					OutGoingOperations: []*sovereignCore.OutGoingOperation{
						{
							Hash: bridgeOp1Hash,
							Data: bridgeOp1,
						},
						{
							Hash: bridgeOp2Hash,
							Data: bridgeOp2,
						},
					},
				}, data)
			default:
				require.Fail(t, "should not add in pool any other operation")
			}
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	scbp, _ := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
		ShardProcessor:               sp,
		ValidatorStatisticsProcessor: &testscommon.ValidatorStatisticsProcessorStub{},
		OutgoingOperationsFormatter:  outgoingOperationsFormatter,
		OutGoingOperationsPool:       outGoingOperationsPool,
		OperationsHasher:             outgoingOpsHasher,
		EpochStartDataCreator:        &mock.EpochStartDataCreatorStub{},
		EpochRewardsCreator:          &testscommon.RewardsCreatorStub{},
		ValidatorInfoCreator:         &testscommon.EpochValidatorInfoCreatorStub{},
		EpochSystemSCProcessor:       &testscommon.EpochStartSystemSCStub{},
		EpochEconomics:               &mock.EpochEconomicsStub{},
		SCToProtocol:                 &mock.SCToProtocolStub{},
	})

	sovChainHdr := &block.SovereignChainHeader{}
	processedMb := &block.MiniBlock{
		ReceiverShardID: core.SovereignChainShardId,
		SenderShardID:   core.MainChainShardId,
	}
	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{processedMb},
	}

	err := scbp.CreateAndSetOutGoingMiniBlock(sovChainHdr, blockBody)
	require.Nil(t, err)
	require.Equal(t, 1, poolAddCt)

	expectedOutGoingMb := &block.MiniBlock{
		TxHashes:        [][]byte{bridgeOp1Hash, bridgeOp2Hash},
		ReceiverShardID: core.MainChainShardId,
		SenderShardID:   arguments.BootstrapComponents.ShardCoordinator().SelfId(),
	}
	expectedBlockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{processedMb, expectedOutGoingMb},
	}
	require.Equal(t, expectedBlockBody, blockBody)

	expectedOutGoingMbHash, err := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), expectedOutGoingMb)
	require.Nil(t, err)

	expectedSovChainHeader := &block.SovereignChainHeader{
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			Hash:                   expectedOutGoingMbHash,
			OutGoingOperationsHash: bridgeOpsHash,
		},
	}
	require.Equal(t, expectedSovChainHeader, sovChainHdr)
}

func TestSovereignShardProcessor_CreateNewBlockExpectCheckRoundCalled(t *testing.T) {
	t.Parallel()

	round := uint64(4)
	checkRoundCt := atomicCore.Counter{}

	roundsNotifier := &epochNotifier.RoundNotifierStub{
		CheckRoundCalled: func(header data.HeaderHandler) {
			checkRoundCt.Increment()
			require.Equal(t, round, header.GetRound())
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.RoundNotifierField = roundsNotifier
	arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sovArgs := createArgsSovereignChainBlockProcessor(arguments)
	scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
	require.Nil(t, err)

	headerHandler, err := scbp.CreateNewHeader(round, 1)
	require.Nil(t, err)
	require.Equal(t, &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce:           1,
			Round:           4,
			AccumulatedFees: big.NewInt(0),
			DeveloperFees:   big.NewInt(0),
		},
		AccumulatedFeesInEpoch: big.NewInt(0),
	}, headerHandler)
	require.Equal(t, int64(1), checkRoundCt.Get())
}

func TestSovereignShardProcessor_CreateNewHeaderValsOK(t *testing.T) {
	t.Parallel()

	round := uint64(7)
	nonce := uint64(5)

	sovArgs := createSovChainBlockProcessorArgs()
	scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
	require.Nil(t, err)

	h, err := scbp.CreateNewHeader(round, nonce)
	require.Nil(t, err)
	require.Equal(t, &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce:           5,
			Round:           7,
			AccumulatedFees: big.NewInt(0),
			DeveloperFees:   big.NewInt(0),
		},
		AccumulatedFeesInEpoch: big.NewInt(0),
	}, h)
}

func TestSovereignShardProcessor_CreateBlock(t *testing.T) {
	t.Parallel()

	doesHaveTime := func() bool {
		return true
	}

	t.Run("nil block should error", func(t *testing.T) {
		sovArgs := createSovChainBlockProcessorArgs()
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		hdr, body, err := scbp.CreateBlock(nil, doesHaveTime)
		require.True(t, check.IfNil(body))
		require.True(t, check.IfNil(hdr))
		require.Equal(t, process.ErrNilBlockHeader, err)
	})
	t.Run("wrong header type should error", func(t *testing.T) {
		sovArgs := createSovChainBlockProcessorArgs()
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		meta := &block.MetaBlock{}

		hdr, body, err := scbp.CreateBlock(meta, doesHaveTime)
		require.True(t, check.IfNil(body))
		require.True(t, check.IfNil(hdr))
		require.ErrorContains(t, err, process.ErrWrongTypeAssertion.Error())
	})
	t.Run("account state dirty should error", func(t *testing.T) {
		journalLen := func() int { return 3 }
		revToSnapshot := func(snapshot int) error { return nil }
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}

		busyIdleCalled := make([]string, 0)
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.ProcessStatusHandlerField = &testscommon.ProcessStatusHandlerStub{
			SetIdleCalled: func() {
				busyIdleCalled = append(busyIdleCalled, idleIdentifier)
			},
			SetBusyCalled: func(reason string) {
				busyIdleCalled = append(busyIdleCalled, busyIdentifier)
			},
		}
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		}

		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		sovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce:         1,
				PubKeysBitmap: []byte("0100101"),
				PrevHash:      []byte(""),
				PrevRandSeed:  []byte("rand seed"),
				Signature:     []byte("signature"),
				RootHash:      []byte("roothash"),
			},
		}

		hdr, body, err := scbp.CreateBlock(sovHeader, doesHaveTime)
		require.True(t, check.IfNil(body))
		require.True(t, check.IfNil(hdr))
		require.Equal(t, process.ErrAccountStateDirty, err)
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
	t.Run("create block started should error because add intermediate txs in tx coordinator returns error", func(t *testing.T) {
		expectedErr := fmt.Errorf("createBlockStarted error")
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}

		busyIdleCalled := make([]string, 0)
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.ProcessStatusHandlerField = &testscommon.ProcessStatusHandlerStub{
			SetIdleCalled: func() {
				busyIdleCalled = append(busyIdleCalled, idleIdentifier)
			},
			SetBusyCalled: func(reason string) {
				busyIdleCalled = append(busyIdleCalled, busyIdentifier)
			},
		}
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			AddIntermediateTransactionsCalled: func(_ map[block.Type][]data.TransactionHandler, _ []byte) error {
				return expectedErr
			},
		}
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)

		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 37,
				Round: 38,
				Epoch: 1,
			},
		}

		hdr, bodyHandler, err := scbp.CreateBlock(expectedSovHeader, doesHaveTime)
		require.True(t, check.IfNil(bodyHandler))
		require.True(t, check.IfNil(hdr))
		require.Equal(t, expectedErr, err)
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
	t.Run("should work with sovereign header and will change epoch", func(t *testing.T) {
		currentEpoch := uint32(1)
		nextEpoch := uint32(2)
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}

		busyIdleCalled := make([]string, 0)
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.ProcessStatusHandlerField = &testscommon.ProcessStatusHandlerStub{
			SetIdleCalled: func() {
				busyIdleCalled = append(busyIdleCalled, idleIdentifier)
			},
			SetBusyCalled: func(reason string) {
				busyIdleCalled = append(busyIdleCalled, busyIdentifier)
			},
		}
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				return true
			},
			MetaEpochCalled: func() uint32 {
				return nextEpoch
			},
		}
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)

		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		sovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 37,
				Round: 38,
				Epoch: currentEpoch,
			},
		}
		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 37,
				Round: 38,
				Epoch: nextEpoch,
			},
			IsStartOfEpoch: true,
		}

		hdr, bodyHandler, err := scbp.CreateBlock(sovHeader, doesHaveTime)
		require.False(t, check.IfNil(bodyHandler))
		body, ok := bodyHandler.(*block.Body)
		require.True(t, ok)
		require.Zero(t, len(body.MiniBlocks))
		require.Equal(t, expectedSovHeader, hdr)
		require.Nil(t, err)
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
	t.Run("should work with sovereign header", func(t *testing.T) {
		currentEpoch := uint32(1)
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}

		busyIdleCalled := make([]string, 0)
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.ProcessStatusHandlerField = &testscommon.ProcessStatusHandlerStub{
			SetIdleCalled: func() {
				busyIdleCalled = append(busyIdleCalled, idleIdentifier)
			},
			SetBusyCalled: func(reason string) {
				busyIdleCalled = append(busyIdleCalled, busyIdentifier)
			},
		}
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return currentEpoch
			}}
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)

		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 37,
				Round: 38,
				Epoch: currentEpoch,
			},
		}
		hdr, bodyHandler, err := scbp.CreateBlock(expectedSovHeader, doesHaveTime)
		require.False(t, check.IfNil(bodyHandler))
		body, ok := bodyHandler.(*block.Body)
		require.True(t, ok)
		require.Zero(t, len(body.MiniBlocks))
		require.Equal(t, expectedSovHeader, hdr)
		require.Nil(t, err)
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
}

func TestSovereignShardProcessor_ProcessBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil header handler should error", func(t *testing.T) {
		sovArgs := createSovChainBlockProcessorArgs()
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		_, _, err = scbp.ProcessBlock(nil, &block.Body{}, haveTime)
		require.Equal(t, process.ErrNilBlockHeader, err)
	})
	t.Run("nil body handler should error", func(t *testing.T) {
		sovArgs := createSovChainBlockProcessorArgs()
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		_, _, err = scbp.ProcessBlock(&block.SovereignChainHeader{}, nil, haveTime)
		require.Equal(t, process.ErrNilBlockBody, err)
	})
	t.Run("not enough time should error", func(t *testing.T) {
		sovArgs := createSovChainBlockProcessorArgs()
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		_, _, err = scbp.ProcessBlock(nil, nil, nil)
		require.Equal(t, process.ErrNilHaveTimeHandler, err)
	})
	t.Run("process header with incorrect epoch should error", func(t *testing.T) {
		randSeed := []byte("rand seed")
		blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
		_ = blkc.SetCurrentBlockHeaderAndRootHash(
			&block.Header{
				Nonce:    0,
				Round:    0,
				Epoch:    1,
				RandSeed: randSeed,
			}, []byte("root hash"),
		)
		_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
		blkc.SetCurrentBlockHeaderHash([]byte("zzz"))

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = blkc
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		hdr := &block.Header{
			Nonce:         1,
			Round:         1,
			Epoch:         0,
			PubKeysBitmap: []byte("0100101"),
			PrevHash:      []byte("zzz"),
			PrevRandSeed:  randSeed,
			Signature:     []byte("signature"),
			RootHash:      []byte("root hash"),
		}
		_, _, err = scbp.ProcessBlock(hdr, &block.Body{}, haveTime)
		require.Equal(t, process.ErrEpochDoesNotMatch, err)
	})
	t.Run("create block started should error because add intermediate txs in tx coordinator returns error", func(t *testing.T) {
		expectedErr := fmt.Errorf("createBlockStarted error")

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			AddIntermediateTransactionsCalled: func(_ map[block.Type][]data.TransactionHandler, _ []byte) error {
				return expectedErr
			},
		}
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)

		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 1,
				Round: 1,
				Epoch: 0,
			},
		}

		hdr, bodyHandler, err := scbp.ProcessBlock(expectedSovHeader, &block.Body{}, haveTime)
		require.True(t, check.IfNil(bodyHandler))
		require.True(t, check.IfNil(hdr))
		require.Equal(t, expectedErr, err)
	})
	t.Run("data not prepared for processing should error should error", func(t *testing.T) {
		expectedErr := fmt.Errorf("expected error")

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			IsDataPreparedForProcessingCalled: func(_ func() time.Duration) error {
				return expectedErr
			},
		}
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 1,
				Round: 1,
				Epoch: 0,
			},
		}

		hdr, bodyHandler, err := scbp.ProcessBlock(expectedSovHeader, &block.Body{}, haveTime)
		require.True(t, check.IfNil(bodyHandler))
		require.True(t, check.IfNil(hdr))
		require.Equal(t, expectedErr, err)
	})
	t.Run("account state dirty should error", func(t *testing.T) {
		journalLen := func() int { return 3 }
		revToSnapshot := func(snapshot int) error { return nil }

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		}

		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		sovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce:         1,
				PubKeysBitmap: []byte("0100101"),
				PrevHash:      []byte(""),
				PrevRandSeed:  []byte("rand seed"),
				Signature:     []byte("signature"),
				RootHash:      []byte("roothash"),
			},
		}

		_, _, err = scbp.ProcessBlock(sovHeader, &block.Body{}, haveTime)
		require.NotNil(t, err)
		require.Equal(t, process.ErrAccountStateDirty, err)
	})
	t.Run("process block should work", func(t *testing.T) {
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}
		randSeed := []byte("rand seed")
		blockHash := []byte("block hash")
		blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
		_ = blkc.SetCurrentBlockHeaderAndRootHash(
			&block.SovereignChainHeader{
				Header: &block.Header{
					Nonce:    4,
					Round:    4,
					Epoch:    0,
					RandSeed: randSeed,
				},
				AccumulatedFeesInEpoch: big.NewInt(1),
				DevFeesInEpoch:         big.NewInt(1),
			},
			[]byte("root hash"),
		)
		_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
		blkc.SetCurrentBlockHeaderHash(blockHash)

		busyIdleCalled := make([]string, 0)
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = blkc
		coreComponents.ProcessStatusHandlerField = &testscommon.ProcessStatusHandlerStub{
			SetIdleCalled: func() {
				busyIdleCalled = append(busyIdleCalled, idleIdentifier)
			},
			SetBusyCalled: func(reason string) {
				busyIdleCalled = append(busyIdleCalled, busyIdentifier)
			},
		}
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce:         5,
				Round:         5,
				Epoch:         0,
				PubKeysBitmap: []byte("0100101"),
				PrevHash:      blockHash,
				PrevRandSeed:  randSeed,
				Signature:     []byte("signature"),
				RootHash:      []byte("root hash"),
			},
		}

		hdr, bodyHandler, err := scbp.ProcessBlock(expectedSovHeader, &block.Body{}, haveTime)
		require.Nil(t, err)
		require.Equal(t, expectedSovHeader, hdr)
		require.False(t, check.IfNil(bodyHandler))
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
	t.Run("process block start of epoch should work", func(t *testing.T) {
		expectedBusyIdleSequencePerCall := []string{busyIdentifier, idleIdentifier}
		randSeed := []byte("rand seed")
		blockHash := []byte("block hash")
		blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
		_ = blkc.SetCurrentBlockHeaderAndRootHash(
			&block.SovereignChainHeader{
				Header: &block.Header{
					Nonce:           4,
					Round:           4,
					Epoch:           0,
					RandSeed:        randSeed,
					AccumulatedFees: big.NewInt(0),
					DeveloperFees:   big.NewInt(0),
				},
				AccumulatedFeesInEpoch: big.NewInt(0),
				DevFeesInEpoch:         big.NewInt(0),
			},
			[]byte("root hash"),
		)
		_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
		blkc.SetCurrentBlockHeaderHash(blockHash)

		busyIdleCalled := make([]string, 0)
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = blkc
		coreComponents.ProcessStatusHandlerField = &testscommon.ProcessStatusHandlerStub{
			SetIdleCalled: func() {
				busyIdleCalled = append(busyIdleCalled, idleIdentifier)
			},
			SetBusyCalled: func(reason string) {
				busyIdleCalled = append(busyIdleCalled, busyIdentifier)
			},
		}
		arguments := createShardBlockProcessorArgsForSovereign(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sovArgs := createArgsSovereignChainBlockProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sovArgs)
		require.Nil(t, err)

		expectedSovHeader := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce:           5,
				Round:           5,
				Epoch:           0,
				PubKeysBitmap:   []byte("0100101"),
				PrevHash:        blockHash,
				PrevRandSeed:    randSeed,
				Signature:       []byte("signature"),
				RootHash:        []byte("root hash"),
				AccumulatedFees: big.NewInt(0),
				DeveloperFees:   big.NewInt(0),
			},
			IsStartOfEpoch: true,
		}

		hdr, bodyHandler, err := scbp.ProcessBlock(expectedSovHeader, &block.Body{}, haveTime)
		require.Nil(t, err)
		require.Equal(t, expectedSovHeader, hdr)
		require.False(t, check.IfNil(bodyHandler))
		require.Equal(t, expectedBusyIdleSequencePerCall, busyIdleCalled)
	})
}
