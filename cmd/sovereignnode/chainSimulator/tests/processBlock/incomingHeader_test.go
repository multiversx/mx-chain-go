package processBlock

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"
)

const (
	defaultPathToInitialConfig     = "../../../../node/config/"
	sovereignConfigPath            = "../../../config/"
	eventIDDepositIncomingTransfer = "deposit"
	topicIDDepositIncomingTransfer = "deposit"
	hashSize                       = 32
)

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

		incomingHeader, headerHash := createIncomingHeader(nodeHandler, &headerNonce, prevHeader, txsEvent)
		err = nodeHandler.GetIncomingHeaderSubscriber().AddHeader(headerHash, incomingHeader)
		require.Nil(t, err)

		prevHeader = incomingHeader.Header

		err = cs.GenerateBlocks(1)
		require.Nil(t, err)
	}

	esdts, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(receiverWallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, esdts[token] != nil)
	require.Equal(t, amountToTransfer, esdts[token].Value.String())
}

func createIncomingHeader(nodeHandler process.NodeHandler, headerNonce *uint64, prevHeader *block.HeaderV2, txsEvent []*transaction.Event) (*sovereign.IncomingHeader, []byte) {
	*headerNonce++
	prevHeaderHash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), prevHeader)
	incomingHeader := &sovereign.IncomingHeader{
		Header:         createHeaderV2(*headerNonce, prevHeaderHash, prevHeader.GetRandSeed()),
		IncomingEvents: txsEvent,
	}
	headerHash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), incomingHeader.Header)

	return incomingHeader, headerHash
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

func createTransactionsEvent(dataCodecHandler dataCodec.SovereignDataCodec, receiver []byte, token string, amountToTransfer string) []*transaction.Event {
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
