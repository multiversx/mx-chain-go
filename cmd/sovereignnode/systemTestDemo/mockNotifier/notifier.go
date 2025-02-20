package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-bridge-go/cert"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Before merging anything into feat/chain-go-sdk, please try a "stress" system test with a local testnet and this notifier.
// Steps:
// 1. Replace github.com/multiversx/mx-chain-communication-go from cmd/sovereignnode/systemTestDemo/go.mod with the one
// from this branch: sovereign-stress-test-branch.
// 2. Keep the config in variables.sh with at least 3 validators.
//
// If you are running with a local testnet and need the necessary certificate files to mock bridge operations, you
// can find them(certificate.crt + private_key.pem) within testnet environment setup at ~MultiversX/testnet/node/config

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
		grpcEnabled,
		sovereignBridgeCertificateFile,
		sovereignBridgeCertificatePkFile,
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

	mockedGRPCServer, grpcServerConn, err := createAndStartGRPCServer(ctx)
	if err != nil {
		log.Error("cannot create grpc server", "error", err)
		return err
	}

	defer func() {
		grpcServerConn.Stop()
		err = host.Close()
		log.LogIfError(err)
	}()

	subscribedAddr, err := pubKeyConverter.Decode(subscribedAddress)
	if err != nil {
		return err
	}

	nonce := uint64(3)
	prevHash := generateRandomHash()
	prevRandSeed := generateRandomHash()
	for {
		headerV2 := createHeaderV2(nonce, prevHash, prevRandSeed)

		confirmedBridgeOps, err := mockedGRPCServer.ExtractRandomBridgeTopicsForConfirmation()
		log.LogIfError(err)

		outportBlock, err := createOutportBlock(headerV2, subscribedAddr, confirmedBridgeOps)
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

		time.Sleep(2000 * time.Millisecond)

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
			Version:                    1,
		},
		Marshaller: marshaller,
		Log:        log,
	}

	return factoryHost.CreateWebSocketHost(args)
}

func createAndStartGRPCServer(ctx *cli.Context) (MockServer, GRPCServerMock, error) {
	if !ctx.Bool(grpcEnabled.Name) {
		return NewDisabledMockServer(), NewDisabledGRPCServer(), nil
	}

	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return nil, nil, err
	}

	tlsConfig, err := cert.LoadTLSServerConfig(cert.FileCfg{
		CertFile: getAbsolutePath(ctx.GlobalString(sovereignBridgeCertificateFile.Name)),
		PkFile:   getAbsolutePath(ctx.GlobalString(sovereignBridgeCertificatePkFile.Name)),
	})
	if err != nil {
		return nil, nil, err
	}
	tlsCredentials := credentials.NewTLS(tlsConfig)
	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCredentials),
	)
	mockedServer := NewMockServer()
	sovereign.RegisterBridgeTxSenderServer(grpcServer, mockedServer)

	log.Info("starting grpc server...")

	go func() {
		for {
			if err = grpcServer.Serve(listener); err != nil {
				log.LogIfError(err)
				time.Sleep(time.Second)
			}
		}
	}()

	return mockedServer, grpcServer, nil
}

func getAbsolutePath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error("Error getting home directory: " + err.Error())
		return ""
	}
	return strings.Replace(path, "~", homeDir, 1)
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
			ChainID:      []byte("1"),
		},
	}
}

func createOutportBlock(headerV2 *block.HeaderV2, subscribedAddr []byte, confirmedBridgeOps []*ConfirmedBridgeOp) (*outport.OutportBlock, error) {
	blockData, err := createBlockData(headerV2)
	if err != nil {
		return nil, err
	}
	incomingLogs, err := createLogs(subscribedAddr, headerV2.GetNonce())
	if err != nil {
		return nil, err
	}

	logs := make([]*outport.LogData, 0)
	logs = append(logs, incomingLogs...)

	bridgeConfirmationLogs := createOutGoingBridgeOpsConfirmationLogs(confirmedBridgeOps, subscribedAddr)
	if len(bridgeConfirmationLogs) != 0 {
		logs = append(logs, bridgeConfirmationLogs...)
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

	eventData := createEventData(ct, subscribedAddr)

	return []*outport.LogData{
		{
			Log: &transaction.Log{
				Address: nil,
				Events: []*transaction.Event{
					{
						Address:    subscribedAddr,
						Identifier: []byte("deposit"),
						Topics:     topics,
						Data:       eventData,
					},
				},
			},
		},
	}, nil
}

func createOutGoingBridgeOpsConfirmationLogs(confirmedBridgeOps []*ConfirmedBridgeOp, subscribedAddr []byte) []*outport.LogData {
	ret := make([]*outport.LogData, 0, len(confirmedBridgeOps))
	for _, confirmedBridgeOp := range confirmedBridgeOps {
		ret = append(ret, &outport.LogData{
			Log: &transaction.Log{
				Events: []*transaction.Event{
					{
						Address:    subscribedAddr,
						Identifier: []byte("execute"),
						Topics:     [][]byte{[]byte("executedBridgeOp"), confirmedBridgeOp.HashOfHashes, confirmedBridgeOp.BridgeOpHash},
					},
				},
			},
		})
	}

	return ret
}

func createTransferTopics(addr []byte, ct int64) ([][]byte, error) {
	nftTransferNonce := big.NewInt(ct%2 + 1)
	nftMetaData := createESDTTokenData(core.NonFungibleV2, big.NewInt(1).Bytes(), "hash", "name", "attributes", addr, big.NewInt(1000).Bytes(), "uri1", "uri2")
	transferNFT := [][]byte{
		[]byte("ASH-a642d1"),     // id
		nftTransferNonce.Bytes(), // nonce != 0
		nftMetaData,              // meta data
	}

	tokenMetaData := createESDTTokenData(core.Fungible, big.NewInt(50+ct).Bytes(), "", "", "", addr, big.NewInt(0).Bytes())
	transferESDT := [][]byte{
		[]byte("WEGLD-bd4d79"), // id
		big.NewInt(0).Bytes(),  // nonce = 0
		tokenMetaData,          // meta data
	}

	topics := make([][]byte, 0)
	topics = append(topics, []byte("deposit"))
	topics = append(topics, addr)
	topics = append(topics, transferNFT...)
	topics = append(topics, transferESDT...)
	return topics, nil
}

func createESDTTokenData(
	esdtType core.ESDTType,
	amount []byte,
	hash string,
	name string,
	attributes string,
	creator []byte,
	royalties []byte,
	uris ...string,
) []byte {
	esdtTokenData := make([]byte, 0)
	esdtTokenData = append(esdtTokenData, uint8(esdtType))                                        // esdt type
	esdtTokenData = append(esdtTokenData, numberToBytes(uint64(len(amount)), lenItemSize)...)     // length of amount
	esdtTokenData = append(esdtTokenData, amount...)                                              // amount
	esdtTokenData = append(esdtTokenData, []byte{0x00}...)                                        // not frozen
	esdtTokenData = append(esdtTokenData, numberToBytes(uint64(len(hash)), lenItemSize)...)       // length of hash
	esdtTokenData = append(esdtTokenData, hash...)                                                // hash
	esdtTokenData = append(esdtTokenData, numberToBytes(uint64(len(name)), lenItemSize)...)       // length of name
	esdtTokenData = append(esdtTokenData, name...)                                                // name
	esdtTokenData = append(esdtTokenData, numberToBytes(uint64(len(attributes)), lenItemSize)...) // length of attributes
	esdtTokenData = append(esdtTokenData, attributes...)                                          // attributes
	esdtTokenData = append(esdtTokenData, creator...)                                             // creator
	esdtTokenData = append(esdtTokenData, numberToBytes(uint64(len(royalties)), lenItemSize)...)  // length of royalties
	esdtTokenData = append(esdtTokenData, royalties...)                                           // royalties
	esdtTokenData = append(esdtTokenData, numberToBytes(uint64(len(uris)), lenItemSize)...)       // number of uris
	for _, uri := range uris {
		esdtTokenData = append(esdtTokenData, numberToBytes(uint64(len(uri)), lenItemSize)...) // length of uri
		esdtTokenData = append(esdtTokenData, []byte(uri)...)                                  // uri
	}

	return esdtTokenData
}

func createEventData(nonce uint64, addr []byte) []byte {
	gasLimit := uint64(10000000)
	function := []byte("func")
	args := [][]byte{[]byte("arg1")}

	eventData := make([]byte, 0)
	eventData = append(eventData, numberToBytes(nonce, u64Size)...)                     // event nonce
	eventData = append(eventData, addr...)                                              // original sender
	eventData = append(eventData, []byte{0x01}...)                                      // has transfer data
	eventData = append(eventData, numberToBytes(gasLimit, u64Size)...)                  // gas limit bytes
	eventData = append(eventData, numberToBytes(uint64(len(function)), lenItemSize)...) // length of function
	eventData = append(eventData, function...)                                          // function
	eventData = append(eventData, numberToBytes(uint64(len(args)), lenItemSize)...)     // number of arguments
	for _, arg := range args {
		eventData = append(eventData, numberToBytes(uint64(len(arg)), lenItemSize)...) // length of current argument
		eventData = append(eventData, arg...)                                          // current argument
	}

	return eventData
}

func numberToBytes(number uint64, size int) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, number)
	return result[8-size:]
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
}
