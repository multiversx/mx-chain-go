package bn

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

var log = logger.NewDefaultLogger()

// Worker defines the data needed by spos to communicate between nodes which are in the validators group
type Worker struct {
	bootstraper      process.Bootstraper
	consensusState   *spos.ConsensusState
	keyGenerator     crypto.KeyGenerator
	marshalizer      marshal.Marshalizer
	privateKey       crypto.PrivateKey
	rounder          round.Rounder
	shardCoordinator sharding.ShardCoordinator

	ReceivedMessages      map[MessageType][]*spos.ConsensusData
	ReceivedMessagesCalls map[MessageType]func(*spos.ConsensusData) bool

	ExecuteMessageChannel        chan *spos.ConsensusData
	ReceivedMessagesChannel      chan *spos.ConsensusData
	ConsensusStateChangedChannel chan bool

	BroadcastTxBlockBody func([]byte)
	BroadcastHeader      func([]byte)
	SendMessage          func(consensus *spos.ConsensusData)

	mutReceivedMessages      sync.RWMutex
	mutReceivedMessagesCalls sync.RWMutex
}

// NewWorker creates a new Worker object
func NewWorker(
	bootstraper process.Bootstraper,
	consensusState *spos.ConsensusState,
	keyGenerator crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	privateKey crypto.PrivateKey,
	rounder round.Rounder,
	shardCoordinator sharding.ShardCoordinator,
) (*Worker, error) {

	err := checkNewWorkerParams(
		bootstraper,
		consensusState,
		keyGenerator,
		marshalizer,
		privateKey,
		rounder,
		shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	wrk := Worker{
		bootstraper:      bootstraper,
		consensusState:   consensusState,
		keyGenerator:     keyGenerator,
		marshalizer:      marshalizer,
		privateKey:       privateKey,
		rounder:          rounder,
		shardCoordinator: shardCoordinator,
	}

	wrk.ExecuteMessageChannel = make(chan *spos.ConsensusData)
	wrk.ReceivedMessagesChannel = make(chan *spos.ConsensusData, consensusState.ConsensusGroupSize()*consensusSubrounds)
	wrk.ReceivedMessagesCalls = make(map[MessageType]func(*spos.ConsensusData) bool)
	wrk.ConsensusStateChangedChannel = make(chan bool, 1)

	wrk.initReceivedMessages()

	go wrk.checkReceivedMessageChannel()
	go wrk.checkChannels()

	return &wrk, nil
}

func checkNewWorkerParams(
	bootstraper process.Bootstraper,
	consensusState *spos.ConsensusState,
	keyGenerator crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	privateKey crypto.PrivateKey,
	rounder round.Rounder,
	shardCoordinator sharding.ShardCoordinator,
) error {
	if bootstraper == nil {
		return spos.ErrNilBlootstraper
	}

	if consensusState == nil {
		return spos.ErrNilConsensusState
	}

	if keyGenerator == nil {
		return spos.ErrNilKeyGenerator
	}

	if marshalizer == nil {
		return spos.ErrNilMarshalizer
	}

	if privateKey == nil {
		return spos.ErrNilPrivateKey
	}

	if rounder == nil {
		return spos.ErrNilRounder
	}

	if shardCoordinator == nil {
		return spos.ErrNilShardCoordinator
	}

	return nil
}

func (wrk *Worker) initReceivedMessages() {
	wrk.mutReceivedMessages.Lock()

	wrk.ReceivedMessages = make(map[MessageType][]*spos.ConsensusData)

	wrk.ReceivedMessages[MtBlockBody] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtBlockHeader] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtCommitmentHash] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtBitmap] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtCommitment] = make([]*spos.ConsensusData, 0)
	wrk.ReceivedMessages[MtSignature] = make([]*spos.ConsensusData, 0)

	wrk.mutReceivedMessages.Unlock()
}

// AddReceivedMessageCall adds a new handler function for a received messege type
func (wrk *Worker) AddReceivedMessageCall(messageType MessageType, receivedMessageCall func(cnsDta *spos.ConsensusData) bool) {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.ReceivedMessagesCalls[messageType] = receivedMessageCall
	wrk.mutReceivedMessagesCalls.Unlock()
}

// RemoveAllReceivedMessagesCalls removes all the functions handlers
func (wrk *Worker) RemoveAllReceivedMessagesCalls() {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.ReceivedMessagesCalls = make(map[MessageType]func(*spos.ConsensusData) bool)
	wrk.mutReceivedMessagesCalls.Unlock()
}

func (wrk *Worker) getCleanedList(cnsDataList []*spos.ConsensusData) []*spos.ConsensusData {
	cleanedCnsDataList := make([]*spos.ConsensusData, 0)

	for i := 0; i < len(cnsDataList); i++ {
		if cnsDataList[i] == nil {
			continue
		}

		if wrk.consensusState.IsMessageForPastRound(wrk.rounder.Index(), cnsDataList[i].RoundIndex) {
			continue
		}

		cleanedCnsDataList = append(cleanedCnsDataList, cnsDataList[i])
	}

	return cleanedCnsDataList
}

// ReceivedMessage method redirects the received message to the channel which should handle it
func (wrk *Worker) ReceivedMessage(name string, data interface{}, msgInfo *p2p.MessageInfo) {
	if wrk.consensusState.RoundCanceled {
		return
	}

	cnsDta, ok := data.(*spos.ConsensusData)

	if !ok {
		return
	}

	senderOK := wrk.consensusState.IsNodeInConsensusGroup(string(cnsDta.PubKey))

	if !senderOK {
		return
	}

	if wrk.consensusState.IsMessageForPastRound(wrk.rounder.Index(), cnsDta.RoundIndex) {
		return
	}

	if wrk.consensusState.SelfPubKey() == string(cnsDta.PubKey) {
		return
	}

	sigVerifErr := wrk.checkSignature(cnsDta)
	if sigVerifErr != nil {
		return
	}

	wrk.ReceivedMessagesChannel <- cnsDta
}

func (wrk *Worker) checkSignature(cnsDta *spos.ConsensusData) error {
	if cnsDta == nil {
		return spos.ErrNilConsensusData
	}

	if cnsDta.PubKey == nil {
		return spos.ErrNilPublicKey
	}

	if cnsDta.Signature == nil {
		return spos.ErrNilSignature
	}

	pubKey, err := wrk.keyGenerator.PublicKeyFromByteArray(cnsDta.PubKey)

	if err != nil {
		return err
	}

	dataNoSig := *cnsDta
	signature := cnsDta.Signature

	dataNoSig.Signature = nil
	dataNoSigString, err := wrk.marshalizer.Marshal(dataNoSig)

	if err != nil {
		return err
	}

	err = pubKey.Verify(dataNoSigString, signature)

	return err
}

func (wrk *Worker) checkReceivedMessageChannel() {
	for {
		select {
		case cnsDta := <-wrk.ReceivedMessagesChannel:
			wrk.executeReceivedMessages(cnsDta)
		}
	}
}

func (wrk *Worker) executeReceivedMessages(cnsDta *spos.ConsensusData) {
	wrk.mutReceivedMessages.Lock()

	msgType := MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	for i := MtBlockBody; i <= MtSignature; i++ {
		cnsDataList = wrk.ReceivedMessages[i]

		if len(cnsDataList) == 0 {
			continue
		}

		wrk.executeMessage(cnsDataList)
		cleanedCnsDtaList := wrk.getCleanedList(cnsDataList)
		wrk.ReceivedMessages[i] = cleanedCnsDtaList
	}

	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) executeMessage(cnsDtaList []*spos.ConsensusData) {
	for i, cnsDta := range cnsDtaList {
		if cnsDta == nil {
			continue
		}

		if wrk.consensusState.IsMessageForOtherRound(wrk.rounder.Index(), cnsDta.RoundIndex) {
			continue
		}

		msgType := MessageType(cnsDta.MsgType)

		switch msgType {
		case MtBlockBody:
			if wrk.consensusState.Status(SrStartRound) != spos.SsFinished {
				continue
			}
		case MtBlockHeader:
			if wrk.consensusState.Status(SrStartRound) != spos.SsFinished {
				continue
			}
		case MtCommitmentHash:
			if wrk.consensusState.Status(SrBlock) != spos.SsFinished {
				continue
			}
		case MtBitmap:
			if wrk.consensusState.Status(SrBlock) != spos.SsFinished {
				continue
			}
		case MtCommitment:
			if wrk.consensusState.Status(SrBitmap) != spos.SsFinished {
				continue
			}
		case MtSignature:
			if wrk.consensusState.Status(SrBitmap) != spos.SsFinished {
				continue
			}
		}

		cnsDtaList[i] = nil

		wrk.ExecuteMessageChannel <- cnsDta
	}
}

// checkChannels method is used to listen to the channels through which node receives and consumes,
// during the round, different messages from the nodes which are in the validators group
func (wrk *Worker) checkChannels() {
	for {
		select {
		case rcvDta := <-wrk.ExecuteMessageChannel:

			msgType := MessageType(rcvDta.MsgType)

			if callReceivedMessage, exist := wrk.ReceivedMessagesCalls[msgType]; exist {
				if callReceivedMessage(rcvDta) {
					if len(wrk.ConsensusStateChangedChannel) == 0 {
						wrk.ConsensusStateChangedChannel <- true
					}
				}
			}
		}
	}
}

// sendConsensusMessage sends the consensus message
func (wrk *Worker) sendConsensusMessage(cnsDta *spos.ConsensusData) bool {
	signature, err := wrk.genConsensusDataSignature(cnsDta)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	signedCnsData := *cnsDta
	signedCnsData.Signature = signature

	if wrk.SendMessage == nil {
		log.Error("sendMessage call back function is not set\n")
		return false
	}

	go wrk.SendMessage(&signedCnsData)

	return true
}

func (wrk *Worker) genConsensusDataSignature(cnsDta *spos.ConsensusData) ([]byte, error) {

	cnsDtaStr, err := wrk.marshalizer.Marshal(cnsDta)

	if err != nil {
		return nil, err
	}

	signature, err := wrk.privateKey.Sign(cnsDtaStr)

	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (wrk *Worker) broadcastTxBlockBody(blockBody *block.TxBlockBody) error {
	if blockBody == nil {
		return spos.ErrNilTxBlockBody
	}

	message, err := wrk.marshalizer.Marshal(blockBody)

	if err != nil {
		return err
	}

	// job message
	if wrk.BroadcastTxBlockBody == nil {
		return spos.ErrNilOnBroadcastTxBlockBody
	}

	go wrk.BroadcastTxBlockBody(message)

	return nil
}

func (wrk *Worker) broadcastHeader(header *block.Header) error {
	if header == nil {
		return spos.ErrNilBlockHeader
	}

	message, err := wrk.marshalizer.Marshal(header)

	if err != nil {
		return err
	}

	// job message
	if wrk.BroadcastHeader == nil {
		return spos.ErrNilOnBroadcastHeader
	}

	go wrk.BroadcastHeader(message)

	return nil
}

func (wrk *Worker) extend(subroundId int) {
	log.Info(fmt.Sprintf("extend function is called from subround: %s\n", getSubroundName(subroundId)))

	if wrk.consensusState.RoundCanceled {
		return
	}

	if wrk.bootstraper.ShouldSync() {
		return
	}

	blk, hdr := wrk.bootstraper.CreateEmptyBlock(wrk.shardCoordinator.ShardForCurrentNode())

	if blk == nil || hdr == nil {
		return
	}

	// broadcast block body
	err := wrk.broadcastTxBlockBody(blk)

	if err != nil {
		log.Info(err.Error())
	}

	// broadcast header
	err = wrk.broadcastHeader(hdr)

	if err != nil {
		log.Info(err.Error())
	}

	return
}

// getMessageTypeName method returns the name of the message from a given message ID
func getMessageTypeName(messageType MessageType) string {
	switch messageType {
	case MtBlockBody:
		return "(BLOCK_BODY)"
	case MtBlockHeader:
		return "(BLOCK_HEADER)"
	case MtCommitmentHash:
		return "(COMMITMENT_HASH)"
	case MtBitmap:
		return "(BITMAP)"
	case MtCommitment:
		return "(COMMITMENT)"
	case MtSignature:
		return "(SIGNATURE)"
	case MtUnknown:
		return "(UNKNOWN)"
	default:
		return "Undefined message type"
	}
}

// getSubroundName returns the name of each subround from a given subround ID
func getSubroundName(subroundId int) string {
	switch subroundId {
	case SrStartRound:
		return "(START_ROUND)"
	case SrBlock:
		return "(BLOCK)"
	case SrCommitmentHash:
		return "(COMMITMENT_HASH)"
	case SrBitmap:
		return "(BITMAP)"
	case SrCommitment:
		return "(COMMITMENT)"
	case SrSignature:
		return "(SIGNATURE)"
	case SrEndRound:
		return "(END_ROUND)"
	default:
		return "Undefined subround"
	}
}
