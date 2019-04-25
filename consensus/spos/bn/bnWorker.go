package bn

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

var log = logger.DefaultLogger()

// worker defines the data needed by spos to communicate between nodes which are in the validators group
type worker struct {
	blockProcessor   process.BlockProcessor
	bootstraper      process.Bootstrapper
	consensusState   *spos.ConsensusState
	keyGenerator     crypto.KeyGenerator
	marshalizer      marshal.Marshalizer
	privateKey       crypto.PrivateKey
	rounder          consensus.Rounder
	shardCoordinator sharding.Coordinator
	singleSigner     crypto.SingleSigner

	receivedMessages      map[spos.MessageType][]*consensus.Message
	receivedMessagesCalls map[spos.MessageType]func(*consensus.Message) bool

	executeMessageChannel         chan *consensus.Message
	consensusStateChangedChannels chan bool

	BroadcastBlock func(data.BodyHandler, data.HeaderHandler) error
	SendMessage    func(consensus *consensus.Message)

	mutReceivedMessages      sync.RWMutex
	mutReceivedMessagesCalls sync.RWMutex
}

// NewWorker creates a new worker object
func NewWorker(
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstrapper,
	consensusState *spos.ConsensusState,
	keyGenerator crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	privateKey crypto.PrivateKey,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) (*worker, error) {
	err := checkNewWorkerParams(
		blockProcessor,
		bootstraper,
		consensusState,
		keyGenerator,
		marshalizer,
		privateKey,
		rounder,
		shardCoordinator,
		singleSigner,
	)
	if err != nil {
		return nil, err
	}

	wrk := worker{
		blockProcessor:   blockProcessor,
		bootstraper:      bootstraper,
		consensusState:   consensusState,
		keyGenerator:     keyGenerator,
		marshalizer:      marshalizer,
		privateKey:       privateKey,
		rounder:          rounder,
		shardCoordinator: shardCoordinator,
		singleSigner:     singleSigner,
	}

	wrk.executeMessageChannel = make(chan *consensus.Message)
	wrk.receivedMessagesCalls = make(map[spos.MessageType]func(*consensus.Message) bool)
	wrk.consensusStateChangedChannels = make(chan bool, 1)
	wrk.bootstraper.AddSyncStateListener(wrk.receivedSyncState)
	wrk.initReceivedMessages()
	go wrk.checkChannels()

	return &wrk, nil
}

func checkNewWorkerParams(
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstrapper,
	consensusState *spos.ConsensusState,
	keyGenerator crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	privateKey crypto.PrivateKey,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) error {
	if blockProcessor == nil {
		return spos.ErrNilBlockProcessor
	}

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

	if singleSigner == nil {
		return spos.ErrNilSingleSigner
	}

	return nil
}

func (wrk *worker) receivedSyncState(isNodeSynchronized bool) {
	if isNodeSynchronized {
		if len(wrk.consensusStateChangedChannels) == 0 {
			wrk.consensusStateChangedChannels <- true
		}
	}
}

func (wrk *worker) initReceivedMessages() {
	wrk.mutReceivedMessages.Lock()

	wrk.receivedMessages = make(map[spos.MessageType][]*consensus.Message)
	wrk.receivedMessages[MtBlockBody] = make([]*consensus.Message, 0)
	wrk.receivedMessages[MtBlockHeader] = make([]*consensus.Message, 0)
	wrk.receivedMessages[MtCommitmentHash] = make([]*consensus.Message, 0)
	wrk.receivedMessages[MtBitmap] = make([]*consensus.Message, 0)
	wrk.receivedMessages[MtCommitment] = make([]*consensus.Message, 0)
	wrk.receivedMessages[MtSignature] = make([]*consensus.Message, 0)

	wrk.mutReceivedMessages.Unlock()
}

// AddReceivedMessageCall adds a new handler function for a received messege type
func (wrk *worker) AddReceivedMessageCall(messageType spos.MessageType, receivedMessageCall func(cnsDta *consensus.Message) bool) {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls[messageType] = receivedMessageCall
	wrk.mutReceivedMessagesCalls.Unlock()
}

// RemoveAllReceivedMessagesCalls removes all the functions handlers
func (wrk *worker) RemoveAllReceivedMessagesCalls() {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls = make(map[spos.MessageType]func(*consensus.Message) bool)
	wrk.mutReceivedMessagesCalls.Unlock()
}

func (wrk *worker) getCleanedList(cnsDataList []*consensus.Message) []*consensus.Message {
	cleanedCnsDataList := make([]*consensus.Message, 0)

	for i := 0; i < len(cnsDataList); i++ {
		if cnsDataList[i] == nil {
			continue
		}

		if wrk.rounder.Index() > cnsDataList[i].RoundIndex {
			continue
		}

		cleanedCnsDataList = append(cleanedCnsDataList, cnsDataList[i])
	}

	return cleanedCnsDataList
}

// ProcessReceivedMessage method redirects the received message to the channel which should handle it
func (wrk *worker) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if message == nil {
		return spos.ErrNilMessage
	}

	if message.Data() == nil {
		return spos.ErrNilDataToProcess
	}

	cnsDta := &consensus.Message{}
	err := wrk.marshalizer.Unmarshal(cnsDta, message.Data())
	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("received %s from %s\n", getStringValue(spos.MessageType(cnsDta.MsgType)), hex.EncodeToString(cnsDta.PubKey)))

	senderOK := wrk.consensusState.IsNodeInEligibleList(string(cnsDta.PubKey))
	if !senderOK {
		return spos.ErrSenderNotOk
	}

	if wrk.consensusState.RoundIndex > cnsDta.RoundIndex {
		return spos.ErrMessageForPastRound
	}

	sigVerifErr := wrk.checkSignature(cnsDta)
	if sigVerifErr != nil {
		return spos.ErrInvalidSignature
	}

	errNotCritical := wrk.checkSelfState(cnsDta)
	if errNotCritical != nil {
		//in this case should return nil but do not process the message
		//nil error will mean that the interceptor will validate this message and broadcast it to the connected peers
		return nil
	}

	go wrk.executeReceivedMessages(cnsDta)
	return nil
}

func (wrk *worker) checkSelfState(cnsDta *consensus.Message) error {
	if wrk.consensusState.SelfPubKey() == string(cnsDta.PubKey) {
		return spos.ErrMessageFromItself
	}

	if wrk.consensusState.RoundCanceled && wrk.consensusState.RoundIndex == cnsDta.RoundIndex {
		return spos.ErrRoundCanceled
	}

	return nil
}

func (wrk *worker) checkSignature(cnsDta *consensus.Message) error {
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

	err = wrk.singleSigner.Verify(pubKey, dataNoSigString, signature)
	return err
}

func (wrk *worker) executeReceivedMessages(cnsDta *consensus.Message) {
	wrk.mutReceivedMessages.Lock()

	msgType := spos.MessageType(cnsDta.MsgType)
	cnsDataList := wrk.receivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.receivedMessages[msgType] = cnsDataList

	for i := MtBlockBody; i <= MtSignature; i++ {
		cnsDataList = wrk.receivedMessages[i]
		if len(cnsDataList) == 0 {
			continue
		}

		wrk.executeMessage(cnsDataList)
		cleanedCnsDtaList := wrk.getCleanedList(cnsDataList)
		wrk.receivedMessages[i] = cleanedCnsDtaList
	}

	wrk.mutReceivedMessages.Unlock()
}

func (wrk *worker) executeMessage(cnsDtaList []*consensus.Message) {
	for i, cnsDta := range cnsDtaList {
		if cnsDta == nil {
			continue
		}

		if wrk.consensusState.RoundIndex != cnsDta.RoundIndex {
			continue
		}

		msgType := spos.MessageType(cnsDta.MsgType)

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
		wrk.executeMessageChannel <- cnsDta
	}
}

// checkChannels method is used to listen to the channels through which node receives and consumes,
// during the round, different messages from the nodes which are in the validators group
func (wrk *worker) checkChannels() {
	for {
		select {
		case rcvDta := <-wrk.executeMessageChannel:
			msgType := spos.MessageType(rcvDta.MsgType)
			if callReceivedMessage, exist := wrk.receivedMessagesCalls[msgType]; exist {
				if callReceivedMessage(rcvDta) {
					if len(wrk.consensusStateChangedChannels) == 0 {
						wrk.consensusStateChangedChannels <- true
					}
				}
			}
		}
	}
}

// sendConsensusMessage sends the consensus message
func (wrk *worker) sendConsensusMessage(cnsDta *consensus.Message) bool {
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

func (wrk *worker) genConsensusDataSignature(cnsDta *consensus.Message) ([]byte, error) {
	cnsDtaStr, err := wrk.marshalizer.Marshal(cnsDta)
	if err != nil {
		return nil, err
	}

	signature, err := wrk.singleSigner.Sign(wrk.privateKey, cnsDtaStr)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (wrk *worker) extend(subroundId int) {
	log.Info(fmt.Sprintf("extend function is called from subround: %s\n", getSubroundName(subroundId)))

	if wrk.bootstraper.ShouldSync() {
		return
	}

	for wrk.consensusState.ProcessingBlock() {
		time.Sleep(time.Millisecond)
	}

	wrk.blockProcessor.RevertAccountState()
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
