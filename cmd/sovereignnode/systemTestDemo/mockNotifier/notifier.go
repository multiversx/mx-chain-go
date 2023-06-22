package main

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"os"
	"time"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

var (
	marshaller, _      = factory.NewMarshalizer("gogo protobuf")
	log                = logger.GetOrCreate("sovereign-mock-notifier")
	hasher             = blake2b.NewBlake2b()
	pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	url                = "localhost:22111"
	subscribedAddress  = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
)

func createEsdtMetaData(value *big.Int, nonce uint64, creator []byte) []byte {
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

	marshalledData, err := marshaller.Marshal(esdtData)
	log.LogIfError(err)
	return marshalledData
}

func createTransfer(addr []byte, ct int64) [][]byte {
	nftTransferNonce := big.NewInt(1 + ct)
	nftTransferValue := big.NewInt(100 + ct)

	transfer1 := [][]byte{
		[]byte("ASH-a642d1"),
		nftTransferNonce.Bytes(),
		createEsdtMetaData(nftTransferValue, nftTransferNonce.Uint64(), addr),
	}
	transfer2 := [][]byte{
		[]byte("WEGLD-bd4d79"),
		big.NewInt(0).Bytes(),
		big.NewInt(50 + ct).Bytes(),
	}
	topic1 := append([][]byte{addr}, transfer1...)
	topic1 = append(topic1, transfer2...)

	return topic1
}

func createLogs(subscribedAddr []byte, ct int) []*outport.LogData {
	return []*outport.LogData{
		{
			Log: &transaction.Log{
				Address: nil,
				Events: []*transaction.Event{
					{
						Address:    subscribedAddr,
						Identifier: []byte("deposit"),
						Topics:     createTransfer(subscribedAddr, int64(ct)),
					},
				},
			},
		},
	}
}

func generateRandomHash() []byte {
	randomBytes := make([]byte, 32)
	_, _ = rand.Read(randomBytes)
	return randomBytes
}

func main() {
	app := cli.NewApp()
	app.Name = "MultiversX sovereign chain notifier"
	app.Usage = "This tool will communicate with an observer/light client connected to mx-chain via " +
		"websocket outport driver and listen to incoming transaction to the specified sovereign chain. If such transactions" +
		"are found, it will format them and forward them to the sovereign shard."
	app.Flags = []cli.Flag{
		logLevel,
		logSaveFile,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = startNotifier

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startNotifier(ctx *cli.Context) error {
	wsServer, err := createWSHost()
	if err != nil {
		log.Error("cannot create WebSocket server", "error", err)
		return err
	}

	prevHash := generateRandomHash()
	prevRandSeed := generateRandomHash()
	nonce := uint64(10)
	subscribedAddr, err := pubKeyConverter.Decode(subscribedAddress)
	log.LogIfError(err)

	for {
		headerV2 := createHeaderV2(nonce, prevHash, prevRandSeed)
		outportBlock, err := createOutportBlock(headerV2, nonce, subscribedAddr)
		if err != nil {
			return err
		}

		outportBlockBytes, err := marshaller.Marshal(outportBlock)
		if err != nil {
			return err
		}

		log.Info("sending block",
			"nonce", nonce,
			"hash", hex.EncodeToString(outportBlock.BlockData.HeaderHash),
			"prev hash", prevHash,
			"rand seed", headerV2.GetRandSeed(),
			"prev rand seed", prevRandSeed)

		err = wsServer.Send(outportBlockBytes, outport.TopicSaveBlock)
		log.LogIfError(err)

		time.Sleep(3000 * time.Millisecond)

		prevHash = outportBlock.BlockData.HeaderHash
		err = sendFinalizedBlock(prevHash, wsServer)
		log.LogIfError(err)

		prevRandSeed = headerV2.GetRandSeed()
		nonce++
	}
}

func createOutportBlock(headerV2 *block.HeaderV2, nonce uint64, subscribedAddr []byte) (*outport.OutportBlock, error) {
	logs := make([]*outport.LogData, 0)

	if nonce%3 == 0 {
		logs = createLogs(subscribedAddr, int(nonce))
	}

	blockData, err := createBlockData(headerV2)
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

func sendFinalizedBlock(hash []byte, wsServer factoryHost.FullDuplexHost) error {
	finalizedBlock := &outport.FinalizedBlock{
		HeaderHash: hash,
	}
	finalizedBlockBytes, err := marshaller.Marshal(finalizedBlock)
	if err != nil {
		return err
	}

	return wsServer.Send(finalizedBlockBytes, outport.TopicFinalizedBlock)
}

func createWSHost() (factoryHost.FullDuplexHost, error) {
	args := factoryHost.ArgsWebSocketHost{
		WebSocketConfig: data.WebSocketConfig{
			URL:                        url,
			Mode:                       data.ModeServer,
			RetryDurationInSec:         1,
			WithAcknowledge:            true,
			BlockingAckOnError:         false,
			DropMessagesIfNoConnection: false,
		},
		Marshaller: marshaller,
		Log:        log,
	}

	return factoryHost.CreateWebSocketHost(args)
}
