package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "MultiversX sovereign chain mock notifier"
	app.Usage = "This tool serves as an observer notifier for a sovereign shard. It initiates the transmission of blocks " +
		"starting from an arbitrary nonce, with incoming events occurring every 3 blocks. Each incoming event comprises " +
		"an NFT and an ESDT transfer. The periodic transmission includes 2 NFTs (ASH-a642d1-01 & ASH-a642d1-02) and one " +
		"ESDT (WEGLD-bd4d79). To verify these entities, one can utilize the sovereign proxy at, for example, " +
		fmt.Sprintf("http://127.0.0.1:7950/address/%s/esdt", subscribedAddress) +
		"The blocks are sent with an arbitrary period between them."
	app.Flags = []cli.Flag{
		logLevel,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = startMockNotifier
	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startMockNotifier(ctx *cli.Context) error {
	err := initializeLogger(ctx)
	if err != nil {
		return err
	}

	host, err := createWSHost()
	if err != nil {
		log.Error("cannot create WebSocket server", "error", err)
		return err
	}

	subscribedAddr, err := pubKeyConverter.Decode(subscribedAddress)
	if err != nil {
		return err
	}

	nonce := uint64(10)
	prevHash := generateRandomHash()
	prevRandSeed := generateRandomHash()
	for {
		headerV2 := createHeaderV2(nonce, prevHash, prevRandSeed)
		outportBlock, err := createOutportBlock(headerV2, subscribedAddr)
		if err != nil {
			return err
		}

		headerHash := outportBlock.BlockData.HeaderHash
		log.Info("sending block",
			"nonce", nonce,
			"hash", hex.EncodeToString(headerHash),
			"prev hash", prevHash,
			"rand seed", headerV2.GetRandSeed(),
			"prev rand seed", prevRandSeed)

		err = sendOutportBlock(outportBlock, host)
		log.LogIfError(err)

		time.Sleep(3000 * time.Millisecond)

		err = sendFinalizedBlock(headerHash, host)
		log.LogIfError(err)

		nonce++
		prevHash = headerHash
		prevRandSeed = headerV2.GetRandSeed()
	}
}

func initializeLogger(ctx *cli.Context) error {
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	return logger.SetLogLevel(logLevelFlagValue)
}

func createWSHost() (factoryHost.FullDuplexHost, error) {
	args := factoryHost.ArgsWebSocketHost{
		WebSocketConfig: data.WebSocketConfig{
			URL:                        wsURL,
			Mode:                       data.ModeServer,
			RetryDurationInSec:         1,
			WithAcknowledge:            true,
			BlockingAckOnError:         false,
			DropMessagesIfNoConnection: false,
			AcknowledgeTimeoutInSec:    60,
		},
		Marshaller: marshaller,
		Log:        log,
	}

	return factoryHost.CreateWebSocketHost(args)
}

func generateRandomHash() []byte {
	randomBytes := make([]byte, hashSize)
	_, _ = rand.Read(randomBytes)
	return randomBytes
}

func createHeaderV2(nonce uint64, prevHash []byte, prevRandSeed []byte) *block.HeaderV2 {
	return &block.HeaderV2{
		Header: &block.Header{
			PrevHash:     prevHash,
			Nonce:        nonce,
			Round:        nonce,
			RandSeed:     generateRandomHash(),
			PrevRandSeed: prevRandSeed,
		},
	}
}

func createOutportBlock(headerV2 *block.HeaderV2, subscribedAddr []byte) (*outport.OutportBlock, error) {
	blockData, err := createBlockData(headerV2)
	if err != nil {
		return nil, err
	}
	logs, err := createLogs(subscribedAddr, headerV2.GetNonce())
	if err != nil {
		return nil, err
	}

	return &outport.OutportBlock{
		BlockData: blockData,
		TransactionPool: &outport.TransactionPool{
			Logs: logs,
		},
	}, nil
}

func createLogs(subscribedAddr []byte, ct uint64) ([]*outport.LogData, error) {
	if ct%3 != 0 {
		return make([]*outport.LogData, 0), nil
	}

	topics, err := createTransferTopics(subscribedAddr, int64(ct))
	if err != nil {
		return nil, err
	}

	return []*outport.LogData{
		{
			Log: &transaction.Log{
				Address: nil,
				Events: []*transaction.Event{
					{
						Address:    subscribedAddr,
						Identifier: []byte("deposit"),
						Topics:     topics,
					},
				},
			},
		},
	}, nil
}

func createTransferTopics(addr []byte, ct int64) ([][]byte, error) {
	nftTransferNonce := big.NewInt(ct%2 + 1)
	nftTransferValue := big.NewInt(100)
	nftMetaData, err := createNFTMetaData(nftTransferValue, nftTransferNonce.Uint64(), addr)
	if err != nil {
		return nil, err
	}

	transferNFT := [][]byte{
		[]byte("ASH-a642d1"),     // id
		nftTransferNonce.Bytes(), // nonce != 0
		nftMetaData,              // meta data
	}
	transferESDT := [][]byte{
		[]byte("WEGLD-bd4d79"),      // id
		big.NewInt(0).Bytes(),       // nonce = 0
		big.NewInt(50 + ct).Bytes(), // value
	}

	topic := append([][]byte{addr}, transferNFT...)
	topic = append(topic, transferESDT...)
	return topic, nil
}

func createNFTMetaData(value *big.Int, nonce uint64, creator []byte) ([]byte, error) {
	esdtData := &esdt.ESDigitalToken{
		Type:  uint32(core.NonFungible),
		Value: value,
		TokenMetaData: &esdt.MetaData{
			URIs:       [][]byte{[]byte("uri1"), []byte("uri2"), []byte("uri3")},
			Nonce:      nonce,
			Hash:       []byte("NFT hash"),
			Name:       []byte("name nft"),
			Attributes: []byte("attributes"),
			Creator:    creator,
		},
	}

	return marshaller.Marshal(esdtData)
}

func createBlockData(headerV2 *block.HeaderV2) (*outport.BlockData, error) {
	headerBytes, err := marshaller.Marshal(headerV2)
	if err != nil {
		return nil, err
	}

	headerHash, err := core.CalculateHash(marshaller, hasher, headerV2)
	if err != nil {
		return nil, err
	}

	return &outport.BlockData{
		HeaderBytes: headerBytes,
		HeaderType:  string(core.ShardHeaderV2),
		HeaderHash:  headerHash,
	}, nil
}

func sendOutportBlock(outportBlock *outport.OutportBlock, host factoryHost.FullDuplexHost) error {
	outportBlockBytes, err := marshaller.Marshal(outportBlock)
	if err != nil {
		return err
	}

	sendWithRetrial(host, outportBlockBytes, outport.TopicSaveBlock)
	return nil
}

func sendFinalizedBlock(hash []byte, host factoryHost.FullDuplexHost) error {
	finalizedBlock := &outport.FinalizedBlock{
		HeaderHash: hash,
	}
	finalizedBlockBytes, err := marshaller.Marshal(finalizedBlock)
	if err != nil {
		return err
	}

	sendWithRetrial(host, finalizedBlockBytes, outport.TopicFinalizedBlock)
	return nil
}

var ct uint64

func sendWithRetrial(host factoryHost.FullDuplexHost, data []byte, topic string) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		<-timer.C
		err := host.Send(data, topic)
		if err == nil {
			return
		}

		log.Warn("could not send data", "topic", topic, "error", err)
		timer.Reset(3 * time.Second)
	}
	if ct == 0 {
		time.Sleep(3 * time.Second)
	}

	ct++
}
