package processBlock

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	proc "github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	testFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
)

const (
	defaultPathToInitialConfig     = "../../../../node/config/"
	sovereignConfigPath            = "../../../config/"
	eventIDDepositIncomingTransfer = "deposit"
	topicIDDepositIncomingTransfer = "deposit"
	hashSize                       = 32
)

type sovChainBlockTracer interface {
	proc.BlockTracker
	ComputeLongestExtendedShardChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error)
	IsGenesisLastCrossNotarizedHeader() bool
}

// This test will simulate an incoming header.
// At the end of the test the amount of tokens needs to be in the receiver account
func TestSovereignChainSimulator_IncomingHeader(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	token := "TKN-123456"
	amountToTransfer := "123"
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	receiverWallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.ZeroValue)
	require.Nil(t, err)

	headerNonce := uint64(9999999)
	prevHeader := createHeaderV2(headerNonce, generateRandomHash(), generateRandomHash())
	txsEvent := make([]*transaction.Event, 0)

	for i := 0; i < 3; i++ {
		if i == 1 {
			txsEvent = append(txsEvent, createTransactionsEvent(nodeHandler.GetRunTypeComponents().DataCodecHandler(), receiverWallet.Bytes, token, amountToTransfer)...)
		} else {
			txsEvent = nil
		}

		incomingHdr, headerHash := createIncomingHeader(nodeHandler, &headerNonce, prevHeader, txsEvent)
		err = nodeHandler.GetIncomingHeaderSubscriber().AddHeader(headerHash, incomingHdr)
		require.Nil(t, err)

		prevHeader = incomingHdr.Header

		err = cs.GenerateBlocks(1)
		require.Nil(t, err)
	}

	esdts, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(receiverWallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, esdts[token] != nil)
	require.Equal(t, amountToTransfer, esdts[token].Value.String())
}

// In this test we simulate:
// - a sovereign chain with the same round time as mainnet
// - for each generated block in sovereign chain, we receive an incoming header from mainnet
func TestSovereignChainSimulator_AddIncomingHeaderCase1(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startRound := uint64(101)
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch: core.OptionalUint64{
				Value:    20,
				HasValue: true,
			},
			ApiInterface:     api.NewNoApiInterface(),
			MinNodesPerShard: 2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.GeneralConfig.SovereignConfig.MainChainNotarization.MainChainNotarizationStartRound = startRound
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	incomingHdrNonce := startRound - 5
	prevIncomingHeader := createHeaderV2(incomingHdrNonce, generateRandomHash(), generateRandomHash())

	sovBlockTracker := getSovereignBlockTracker(t, nodeHandler)

	var previousExtendedHeaderHash []byte
	var prevSovHdr data.SovereignChainHeaderHandler
	var previousExtendedHeader data.HeaderHandler

	for currIncomingHeaderRound := startRound - 5; currIncomingHeaderRound < startRound+200; currIncomingHeaderRound++ {

		// Handlers are notified on go routines; wait a bit so that pools are updated
		incomingHdr := addIncomingHeader(t, nodeHandler, &incomingHdrNonce, prevIncomingHeader)
		time.Sleep(time.Millisecond * 10)

		// We just received header in pool and notified all subscribed components, header has not been processed + committed.
		// We check how leader will compute the longest incoming header chain
		extendedHeaderHash := getExtendedHeaderHash(t, nodeHandler, incomingHdr)
		longestChain, longestChainHdrHashes, err := sovBlockTracker.ComputeLongestExtendedShardChainFromLastNotarized()
		require.Nil(t, err)

		if currIncomingHeaderRound < startRound {
			require.Empty(t, longestChain)
			require.Empty(t, longestChainHdrHashes)

		} else {
			currentExtendedHeader := getExtendedHeader(t, nodeHandler, incomingHdr)

			// On sovereign epoch start block processing we do not process incoming headers.
			// This means that we accumulate incoming headers in pool and in next sovereign header we need to include
			// accumulating headers
			if prevSovHdr.IsStartOfEpochBlock() {
				require.Equal(t, []data.HeaderHandler{previousExtendedHeader, currentExtendedHeader}, longestChain)
				require.Equal(t, [][]byte{previousExtendedHeaderHash, extendedHeaderHash}, longestChainHdrHashes)
			} else {
				require.Equal(t, []data.HeaderHandler{currentExtendedHeader}, longestChain)
				require.Equal(t, [][]byte{extendedHeaderHash}, longestChainHdrHashes)
			}
		}

		// Process + commit sovereign block with received incoming header
		err = cs.GenerateBlocks(1)
		require.Nil(t, err)

		currentSovHeader := common.GetCurrentSovereignHeader(nodeHandler)
		lastCrossNotarizedHeader, _, err := sovBlockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)
		require.Nil(t, err)

		// Check tracker and blockchain hook state for incoming processed data
		if currIncomingHeaderRound <= 99 {
			require.Zero(t, lastCrossNotarizedHeader.GetRound())
			require.True(t, sovBlockTracker.IsGenesisLastCrossNotarizedHeader())
			require.Empty(t, currentSovHeader.GetExtendedShardHeaderHashes())
		} else if currIncomingHeaderRound == 100 { // pre-genesis incoming header is notarized
			require.Equal(t, uint64(100), lastCrossNotarizedHeader.GetRound())
			require.False(t, sovBlockTracker.IsGenesisLastCrossNotarizedHeader())
			require.Empty(t, currentSovHeader.GetExtendedShardHeaderHashes())
		} else { // since genesis main-chain header, each incoming header is instantly notarized (0 block finality)
			if currentSovHeader.IsStartOfEpochBlock() { // epoch start block, no incoming header process is added to sovereign block
				require.Equal(t, currIncomingHeaderRound-1, lastCrossNotarizedHeader.GetRound())
				require.False(t, sovBlockTracker.IsGenesisLastCrossNotarizedHeader())
				require.Empty(t, currentSovHeader.GetExtendedShardHeaderHashes())
			} else if prevSovHdr.IsStartOfEpochBlock() { // prev sov block was epoch start, should have 2 accumulated incoming headers
				require.Equal(t, currIncomingHeaderRound, lastCrossNotarizedHeader.GetRound())
				require.False(t, sovBlockTracker.IsGenesisLastCrossNotarizedHeader())
				require.Equal(t, [][]byte{previousExtendedHeaderHash, extendedHeaderHash}, currentSovHeader.GetExtendedShardHeaderHashes())
			} else { // normal processing, in each sovereign block, there is an extended header hash
				require.Equal(t, currIncomingHeaderRound, lastCrossNotarizedHeader.GetRound())
				require.False(t, sovBlockTracker.IsGenesisLastCrossNotarizedHeader())
				require.Equal(t, [][]byte{extendedHeaderHash}, currentSovHeader.GetExtendedShardHeaderHashes())
			}
		}

		prevSovHdr = currentSovHeader
		prevIncomingHeader = incomingHdr.Header
		previousExtendedHeaderHash = extendedHeaderHash
		previousExtendedHeader = getExtendedHeader(t, nodeHandler, incomingHdr)
	}

	require.Equal(t, uint32(9), nodeHandler.GetCoreComponents().EpochNotifier().CurrentEpoch())
}

// In this test we simulate:
// - a sovereign chain with a lower round time than mainnet
// - we will receive a mainnet block at every 3 generated sovereign blocks
func TestSovereignChainSimulator_AddIncomingHeaderCase2(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startRound := uint64(101)
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch: core.OptionalUint64{
				Value:    20,
				HasValue: true,
			},
			ApiInterface:     api.NewNoApiInterface(),
			MinNodesPerShard: 2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.GeneralConfig.SovereignConfig.MainChainNotarization.MainChainNotarizationStartRound = startRound
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	sovBlockTracker := getSovereignBlockTracker(t, nodeHandler)
	require.True(t, sovBlockTracker.IsGenesisLastCrossNotarizedHeader())

	incomingHdrNonce := startRound - 1
	prevHeader := createHeaderV2(incomingHdrNonce, generateRandomHash(), generateRandomHash())
	incomingHdr := addIncomingHeader(t, nodeHandler, &incomingHdrNonce, prevHeader)
	require.Nil(t, err)

	lastCrossNotarizedRound := uint64(100)
	// Pre genesis header was added before, but no extended header will be notarized in the next 3 sovereign blocks
	for i := 0; i < 3; i++ {
		err = cs.GenerateBlocks(1)
		require.Nil(t, err)

		currentSovBlock := common.GetCurrentSovereignHeader(nodeHandler)
		require.Empty(t, currentSovBlock.GetExtendedShardHeaderHashes())

		checkLastCrossNotarizedRound(t, sovBlockTracker, lastCrossNotarizedRound)
	}

	// From now on, every 3 sovereign blocks we add one incoming header
	for i := 0; i < 50; i++ {
		if i%3 == 0 {
			prevHeader = incomingHdr.Header
			incomingHdr = addIncomingHeader(t, nodeHandler, &incomingHdrNonce, prevHeader)
			lastCrossNotarizedRound++
		}

		err = cs.GenerateBlocks(1)
		require.Nil(t, err)

		checkLastCrossNotarizedRound(t, sovBlockTracker, lastCrossNotarizedRound)

		currentSovBlock := common.GetCurrentSovereignHeader(nodeHandler)
		if i%3 == 0 {
			extendedHeaderHash := getExtendedHeaderHash(t, nodeHandler, incomingHdr)
			require.Equal(t, [][]byte{extendedHeaderHash}, currentSovBlock.GetExtendedShardHeaderHashes())
		} else {
			require.Empty(t, currentSovBlock.GetExtendedShardHeaderHashes())
		}
	}

}

// In this test we simulate:
// - a sovereign chain with a higher round time than mainnet
// - we will receive 3 mainnet block for every generated sovereign blocks
func TestSovereignChainSimulator_AddIncomingHeaderCase3(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startRound := uint64(101)
	sovConfig := config.SovereignConfig{}

	sovRequestHandler := &testscommon.ExtendedShardHeaderRequestHandlerStub{
		RequestExtendedShardHeaderCalled: func(hash []byte) {
			require.Fail(t, "should not request any extended header")
		},
		RequestExtendedShardHeaderByNonceCalled: func(nonce uint64) {
			require.Fail(t, "should not request any extended header")
		},
	}
	sovRequestHandlerFactory := &testFactory.RequestHandlerFactoryMock{
		CreateRequestHandlerCalled: func(args requestHandlers.RequestHandlerArgs) (proc.RequestHandler, error) {
			return sovRequestHandler, nil
		},
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch: core.OptionalUint64{
				Value:    20,
				HasValue: true,
			},
			ApiInterface:     api.NewNoApiInterface(),
			MinNodesPerShard: 1,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.GeneralConfig.SovereignConfig.MainChainNotarization.MainChainNotarizationStartRound = startRound
				sovConfig = cfg.GeneralConfig.SovereignConfig
			},
			CreateRunTypeComponents: func(args runType.ArgsRunTypeComponents) (factory.RunTypeComponentsHolder, error) {
				runTypeComps, err := common.CreateSovereignRunTypeComponents(args, sovConfig)
				require.Nil(t, err)

				runTypeCompsHolder := components.GetRunTypeComponentsStub(runTypeComps)
				runTypeCompsHolder.RequestHandlerFactory = sovRequestHandlerFactory

				return runTypeCompsHolder, nil
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	sovBlockTracker := getSovereignBlockTracker(t, nodeHandler)
	require.True(t, sovBlockTracker.IsGenesisLastCrossNotarizedHeader())

	incomingHdrNonce := startRound - 3
	prevHeader := createHeaderV2(incomingHdrNonce, generateRandomHash(), generateRandomHash())

	// Fill pool with incoming headers up until pre-genesis
	for i := 0; i < 3; i++ {
		incomingHdr := addIncomingHeader(t, nodeHandler, &incomingHdrNonce, prevHeader)
		prevHeader = incomingHdr.Header
	}

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// We process one sovereign block, no incoming header should be added yet
	lastCrossNotarizedRound := uint64(100)
	checkLastCrossNotarizedRound(t, sovBlockTracker, lastCrossNotarizedRound)

	currentSovBlock := common.GetCurrentSovereignHeader(nodeHandler)
	require.Empty(t, currentSovBlock.GetExtendedShardHeaderHashes())

	prevSovBlock := currentSovBlock
	extendedHeaderHashes := make([][]byte, 0)
	// From now on, we generate 3 incoming headers per sovereign block
	for i := 1; i < 300; i++ {
		incomingHdr := addIncomingHeader(t, nodeHandler, &incomingHdrNonce, prevHeader)
		extendedHeaderHashes = append(extendedHeaderHashes, getExtendedHeaderHash(t, nodeHandler, incomingHdr))

		if i%3 == 0 {
			time.Sleep(time.Millisecond * 50)

			err = cs.GenerateBlocks(1)
			require.Nil(t, err)

			currentSovBlock = common.GetCurrentSovereignHeader(nodeHandler)

			if currentSovBlock.IsStartOfEpochBlock() {
				require.Empty(t, currentSovBlock.GetExtendedShardHeaderHashes())
			} else if prevSovBlock.IsStartOfEpochBlock() {
				require.Len(t, currentSovBlock.GetExtendedShardHeaderHashes(), 6)
				require.Equal(t, extendedHeaderHashes, currentSovBlock.GetExtendedShardHeaderHashes())

				lastCrossNotarizedRound += 6
				extendedHeaderHashes = make([][]byte, 0)
			} else {
				require.Len(t, currentSovBlock.GetExtendedShardHeaderHashes(), 3)
				require.Equal(t, extendedHeaderHashes, currentSovBlock.GetExtendedShardHeaderHashes())

				lastCrossNotarizedRound += 3
				extendedHeaderHashes = make([][]byte, 0)
			}

		}

		checkLastCrossNotarizedRound(t, sovBlockTracker, lastCrossNotarizedRound)
		prevHeader = incomingHdr.Header
		prevSovBlock = currentSovBlock
	}

	require.Equal(t, uint32(4), nodeHandler.GetCoreComponents().EpochNotifier().CurrentEpoch())

}

func getExtendedHeader(t *testing.T, nodeHandler process.NodeHandler, incomingHdr *sovereign.IncomingHeader) data.HeaderHandler {
	extendedHeader, err := nodeHandler.GetIncomingHeaderSubscriber().CreateExtendedHeader(incomingHdr)
	require.Nil(t, err)

	return extendedHeader
}

func getExtendedHeaderHash(t *testing.T, nodeHandler process.NodeHandler, incomingHdr *sovereign.IncomingHeader) []byte {
	extendedHeader := getExtendedHeader(t, nodeHandler, incomingHdr)
	extendedHeaderHash, err := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), extendedHeader)
	require.Nil(t, err)

	return extendedHeaderHash
}

func createIncomingHeader(nodeHandler process.NodeHandler, headerNonce *uint64, prevHeader *block.HeaderV2, txsEvent []*transaction.Event) (*sovereign.IncomingHeader, []byte) {
	*headerNonce++
	prevHeaderHash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), prevHeader)
	incomingHdr := &sovereign.IncomingHeader{
		Header:         createHeaderV2(*headerNonce, prevHeaderHash, prevHeader.GetRandSeed()),
		IncomingEvents: txsEvent,
	}
	headerHash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), incomingHdr.Header)

	return incomingHdr, headerHash
}

func addIncomingHeader(t *testing.T, nodeHandler process.NodeHandler, headerNonce *uint64, prevHeader *block.HeaderV2) *sovereign.IncomingHeader {
	incomingHdr, incomingHeaderHash := createSimpleIncomingHeader(nodeHandler, headerNonce, prevHeader)
	err := nodeHandler.GetIncomingHeaderSubscriber().AddHeader(incomingHeaderHash, incomingHdr)
	require.Nil(t, err)

	return incomingHdr
}

func createSimpleIncomingHeader(nodeHandler process.NodeHandler, headerNonce *uint64, prevHeader *block.HeaderV2) (*sovereign.IncomingHeader, []byte) {
	prevHeaderHash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), prevHeader)
	incomingHdr := &sovereign.IncomingHeader{
		Header:         createHeaderV2(*headerNonce, prevHeaderHash, prevHeader.GetRandSeed()),
		IncomingEvents: []*transaction.Event{},
	}
	headerHash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), incomingHdr.Header)
	*headerNonce++

	return incomingHdr, headerHash
}

func createHeaderV2(nonce uint64, prevHash []byte, prevRandSeed []byte) *block.HeaderV2 {
	return &block.HeaderV2{
		Header: &block.Header{
			PrevHash:     prevHash,
			Nonce:        nonce,
			Round:        nonce,
			RandSeed:     generateRandomHash(),
			PrevRandSeed: prevRandSeed,
			ChainID:      []byte(configs.ChainID),
		},
	}
}

func generateRandomHash() []byte {
	randomBytes := make([]byte, hashSize)
	_, _ = rand.Read(randomBytes)
	return randomBytes
}

func createTransactionsEvent(dataCodecHandler incomingHeader.SovereignDataCodec, receiver []byte, token string, amountToTransfer string) []*transaction.Event {
	tokenData, _ := dataCodecHandler.SerializeTokenData(createTokenData(amountToTransfer))
	eventData, _ := dataCodecHandler.SerializeEventData(createEventData())

	events := make([]*transaction.Event, 0)
	return append(events, &transaction.Event{
		Identifier: []byte(eventIDDepositIncomingTransfer),
		Topics:     [][]byte{[]byte(topicIDDepositIncomingTransfer), receiver, []byte(token), []byte(""), tokenData},
		Data:       eventData,
	})
}

func createTokenData(amountToTransfer string) sovereign.EsdtTokenData {
	creator, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	amount, _ := big.NewInt(0).SetString(amountToTransfer, 10)

	return sovereign.EsdtTokenData{
		TokenType:  core.Fungible,
		Amount:     amount,
		Frozen:     false,
		Hash:       make([]byte, 0),
		Name:       []byte(""),
		Attributes: make([]byte, 0),
		Creator:    creator,
		Royalties:  big.NewInt(0),
		Uris:       make([][]byte, 0),
	}
}

func createEventData() sovereign.EventData {
	sender, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	return sovereign.EventData{
		Nonce:        1,
		Sender:       sender,
		TransferData: nil,
	}
}

func getSovereignBlockTracker(t *testing.T, nodeHandler process.NodeHandler) sovChainBlockTracer {
	sovBlockTracker, castOk := nodeHandler.GetProcessComponents().BlockTracker().(sovChainBlockTracer)
	require.True(t, castOk)

	return sovBlockTracker
}

func checkLastCrossNotarizedRound(t *testing.T, sovBlockTracker sovChainBlockTracer, lastCrossNotarizedRound uint64) {
	lastCrossNotarizedHeader, _, err := sovBlockTracker.GetLastCrossNotarizedHeader(core.MainChainShardId)
	require.Nil(t, err)
	require.Equal(t, lastCrossNotarizedRound, lastCrossNotarizedHeader.GetRound())
	require.False(t, sovBlockTracker.IsGenesisLastCrossNotarizedHeader())
}
