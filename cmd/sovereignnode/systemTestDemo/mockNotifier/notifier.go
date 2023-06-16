package main

import (
	"encoding/hex"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	marshaller, _ = factory.NewMarshalizer("gogo protobuf")
	log           = logger.GetOrCreate("server")
	url           = "localhost:22111"
)

func createTransfer(addr []byte) [][]byte {
	transfer1 := [][]byte{
		[]byte("ASH-a642d1"),
		big.NewInt(4).Bytes(),
		big.NewInt(100).Bytes(),
	}
	transfer2 := [][]byte{
		[]byte("WEGLD-bd4d79"),
		big.NewInt(0).Bytes(),
		big.NewInt(50).Bytes(),
	}
	topic1 := append([][]byte{addr}, transfer1...)
	topic1 = append(topic1, transfer2...)

	return topic1
}

func main() {
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

	wsServer, err := factoryHost.CreateWebSocketHost(args)
	if err != nil {
		log.Error("cannot create WebSocket server", "error", err)
		return
	}

	prevHash, err := hex.DecodeString("c6d5b27501261f1e871214ab5faaba8b7770a185c5b7e146882dbfc8fca9b2ef")
	log.LogIfError(err)

	prevRandSeed, err := hex.DecodeString("e400abed092753418b3c23411dfa4b05abd082180a817fccf3dd2e5d669d1e3f")
	log.LogIfError(err)

	hasher := blake2b.NewBlake2b()
	nonce := uint64(10)
	for {
		pubKeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
		log.LogIfError(err)

		subscribedAddr, err := pubKeyConverter.Decode("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
		log.LogIfError(err)

		outportBlock := &outport.OutportBlock{
			BlockData: &outport.BlockData{},

			TransactionPool: &outport.TransactionPool{
				Logs: []*outport.LogData{
					{
						Log: &transaction.Log{
							Address: nil,
							Events: []*transaction.Event{
								{
									Address:    subscribedAddr,
									Identifier: []byte("deposit"),
									Topics:     createTransfer(subscribedAddr),
								},
							},
						},
					},
				},
			},
		}
		outportBlock.BlockData.HeaderType = string(core.ShardHeaderV2)

		randSeed, err := core.CalculateHash(marshaller, hasher, &outport.OutportBlock{HighestFinalBlockNonce: nonce})
		log.LogIfError(err)

		headerV2 := &block.HeaderV2{
			Header: &block.Header{
				PrevHash:     prevHash,
				Nonce:        nonce,
				Round:        nonce,
				RandSeed:     randSeed,
				PrevRandSeed: prevRandSeed,
			},
		}

		headerBytes, err := marshaller.Marshal(headerV2)
		log.LogIfError(err)

		outportBlock.BlockData.HeaderBytes = headerBytes
		headerHash, err := core.CalculateHash(marshaller, hasher, headerV2)
		log.LogIfError(err)

		outportBlock.BlockData.HeaderHash = headerHash

		outportBlockBytes, err := marshaller.Marshal(outportBlock)
		log.LogIfError(err)

		log.Info("sending block",
			"nonce", nonce,
			"hash", hex.EncodeToString(outportBlock.BlockData.HeaderHash),
			"prev hash", prevHash,
			"rand seed", randSeed,
			"prev rand seed", prevRandSeed)

		err = wsServer.Send(outportBlockBytes, outport.TopicSaveBlock)
		log.LogIfError(err)

		time.Sleep(6000 * time.Millisecond)

		nonce++
		prevHash = outportBlock.BlockData.HeaderHash //core.CalculateHash(marshaller, hasher, outportBlock)
		prevRandSeed = randSeed

		finalizedBlock := &outport.FinalizedBlock{
			HeaderHash: prevHash,
		}
		finalizeedBlockBytes, err := marshaller.Marshal(finalizedBlock)
		log.LogIfError(err)

		err = wsServer.Send(finalizeedBlockBytes, outport.TopicFinalizedBlock)
		log.LogIfError(err)

	}
}
