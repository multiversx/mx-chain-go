package spos

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

var log = logger.NewDefaultLogger()

const (
	// SrStartRound defines ID of subround "Start round"
	SrStartRound chronology.SubroundId = iota
	// SrBlock defines ID of subround "block"
	SrBlock
	// SrCommitmentHash defines ID of subround "commitment hash"
	SrCommitmentHash
	// SrBitmap defines ID of subround "bitmap"
	SrBitmap
	// SrCommitment defines ID of subround "commitment"
	SrCommitment
	// SrSignature defines ID of subround "signature"
	SrSignature
	// SrEndRound defines ID of subround "End round"
	SrEndRound
	// SrAdvance defines ID of subround "Advance"
	SrAdvance
)

//TODO: current shards (this should be injected, and this const should be removed later)
const shardId = 0

//TODO: maximum transactions in one block (this should be injected, and this const should be removed later)
const maxTransactionsInBlock = 15000

// consensusSubrounds specifies how many subrounds of consensus are in this implementation
const consensusSubrounds = 6

// maxBlockProcessingTimePercent specifies which is the max allocated time percent,
// for processing block, from the total time of one round
const maxBlockProcessingTimePercent = float64(0.85)

// MessageType specifies what type of message was received
type MessageType int

const (
	// MtUnknown defines ID of a message that has unknown Data inside
	MtUnknown MessageType = iota
	// MtBlockBody defines ID of a message that has a block body inside
	MtBlockBody
	// MtBlockHeader defines ID of a message that has a block header inside
	MtBlockHeader
	// MtCommitmentHash defines ID of a message that has a commitment hash inside
	MtCommitmentHash
	// MtBitmap defines ID of a message that has a bitmap inside
	MtBitmap
	// MtCommitment defines ID of a message that has a commitment inside
	MtCommitment
	// MtSignature defines ID of a message that has a Signature inside
	MtSignature
)

// ConsensusData defines the data needed by spos to communicate between nodes over network in all subrounds
type ConsensusData struct {
	BlockHeaderHash []byte
	SubRoundData    []byte
	PubKey          []byte
	Signature       []byte
	MsgType         MessageType
	TimeStamp       uint64
	RoundIndex      int32
}

// NewConsensusData creates a new ConsensusData object
func NewConsensusData(
	blHeaderHash []byte,
	subRoundData []byte,
	pubKey []byte,
	sig []byte,
	msg MessageType,
	tms uint64,
	roundIndex int32,
) *ConsensusData {

	return &ConsensusData{
		BlockHeaderHash: blHeaderHash,
		SubRoundData:    subRoundData,
		PubKey:          pubKey,
		Signature:       sig,
		MsgType:         msg,
		TimeStamp:       tms,
		RoundIndex:      roundIndex,
	}
}

// Create method creates a new ConsensusData object
func (cd *ConsensusData) Create() p2p.Creator {
	return &ConsensusData{}
}

// ID gets an unique id of the ConsensusData object
func (cd *ConsensusData) ID() string {
	id := fmt.Sprintf("%d-%s-%d", cd.RoundIndex, cd.Signature, cd.MsgType)
	return id
}

// SPOSConsensusWorker defines the data needed by spos to communicate between nodes which are in the validators group
type SPOSConsensusWorker struct {
	Cns                    *Consensus
	Header                 *block.Header
	BlockBody              *block.TxBlockBody
	BlockChain             *blockchain.BlockChain
	BlockProcessor         process.BlockProcessor
	boot                   process.Bootstraper
	MessageChannels        map[MessageType]chan *ConsensusData
	ReceivedMessageChannel chan *ConsensusData
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	keyGen                 crypto.KeyGenerator
	privKey                crypto.PrivateKey
	pubKey                 crypto.PublicKey
	singleSigner           crypto.SingleSigner
	multiSigner            crypto.MultiSigner
	SendMessage            func(consensus *ConsensusData)
	BroadcastHeader        func([]byte)
	BroadcastBlockBody     func([]byte)
	ReceivedMessages       map[MessageType][]*ConsensusData
	mutReceivedMessages    sync.RWMutex
}

// NewConsensusWorker creates a new SPOSConsensusWorker object
func NewConsensusWorker(
	cns *Consensus,
	blkc *blockchain.BlockChain,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockProcessor process.BlockProcessor,
	boot process.Bootstraper,
	singleSigner crypto.SingleSigner,
	multisig crypto.MultiSigner,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
) (*SPOSConsensusWorker, error) {

	err := checkNewConsensusWorkerParams(
		cns,
		blkc,
		hasher,
		marshalizer,
		blockProcessor,
		boot,
		singleSigner,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	if err != nil {
		return nil, err
	}

	sposWorker := SPOSConsensusWorker{
		Cns:            cns,
		BlockChain:     blkc,
		hasher:         hasher,
		marshalizer:    marshalizer,
		BlockProcessor: blockProcessor,
		boot:           boot,
		singleSigner:   singleSigner,
		multiSigner:    multisig,
		keyGen:         keyGen,
		privKey:        privKey,
		pubKey:         pubKey,
	}

	sposWorker.MessageChannels = make(map[MessageType]chan *ConsensusData)

	nodes := 0

	if cns != nil &&
		cns.RoundConsensus != nil {
		nodes = len(cns.RoundConsensus.ConsensusGroup())
	}

	sposWorker.MessageChannels[MtBlockBody] = make(chan *ConsensusData)
	sposWorker.MessageChannels[MtBlockHeader] = make(chan *ConsensusData)
	sposWorker.MessageChannels[MtCommitmentHash] = make(chan *ConsensusData)
	sposWorker.MessageChannels[MtBitmap] = make(chan *ConsensusData)
	sposWorker.MessageChannels[MtCommitment] = make(chan *ConsensusData)
	sposWorker.MessageChannels[MtSignature] = make(chan *ConsensusData)

	sposWorker.ReceivedMessageChannel = make(chan *ConsensusData, nodes*consensusSubrounds)

	sposWorker.initReceivedMessages()

	go sposWorker.checkReceivedMessageChannel()
	go sposWorker.checkChannels()

	return &sposWorker, nil
}

func checkNewConsensusWorkerParams(
	cns *Consensus,
	blkc *blockchain.BlockChain,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockProcessor process.BlockProcessor,
	boot process.Bootstraper,
	singleSigner crypto.SingleSigner,
	multisig crypto.MultiSigner,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
) error {
	if cns == nil {
		return ErrNilConsensus
	}

	if blkc == nil {
		return ErrNilBlockChain
	}

	if hasher == nil {
		return ErrNilHasher
	}

	if marshalizer == nil {
		return ErrNilMarshalizer
	}

	if blockProcessor == nil {
		return ErrNilBlockProcessor
	}

	if boot == nil {
		return ErrNilBlootstrap
	}

	if singleSigner == nil {
		return ErrNilSingleSigner
	}

	if multisig == nil {
		return ErrNilMultiSigner
	}

	if keyGen == nil {
		return ErrNilKeyGenerator
	}

	if privKey == nil {
		return ErrNilPrivateKey
	}

	if pubKey == nil {
		return ErrNilPublicKey
	}

	return nil
}

func (sposWorker *SPOSConsensusWorker) checkSignaturesValidity(bitmap []byte) error {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroup := sposWorker.Cns.ConsensusGroup()
	consensusGroupSize := len(consensusGroup)
	size := consensusGroupSize

	if consensusGroupSize > nbBitsBitmap {
		size = nbBitsBitmap
	}

	for i := 0; i < size; i++ {
		indexRequired := (bitmap[i/8] & (1 << uint16(i%8))) > 0

		if !indexRequired {
			continue
		}

		pubKey := consensusGroup[i]
		isSigJobDone, err := sposWorker.Cns.GetJobDone(pubKey, SrSignature)

		if err != nil {
			return err
		}

		if !isSigJobDone {
			return ErrNilSignature
		}

		signature, err := sposWorker.multiSigner.SignatureShare(uint16(i))

		if err != nil {
			return err
		}

		// verify partial signature
		err = sposWorker.multiSigner.VerifySignatureShare(uint16(i), signature, bitmap)

		if err != nil {
			return err
		}
	}

	return nil
}

// DoStartRoundJob method is the function which actually does the job of the StartRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct,
// from spos package)
func (sposWorker *SPOSConsensusWorker) DoStartRoundJob() bool {
	sposWorker.BlockBody = nil
	sposWorker.Header = nil
	sposWorker.Cns.Data = nil
	sposWorker.Cns.ResetRoundStatus()
	sposWorker.Cns.ResetRoundState()
	sposWorker.cleanReceivedMessages()

	leader, err := sposWorker.Cns.GetLeader()

	if err != nil {
		log.Error(err.Error())
		leader = "Unknown"
	}

	msg := ""
	if leader == sposWorker.Cns.SelfPubKey() {
		msg = " (MY TURN)"
	}

	log.Info(fmt.Sprintf("%sStep 0: Preparing for this round with leader %s%s\n",
		sposWorker.Cns.getFormattedTime(), hex.EncodeToString([]byte(leader)), msg))

	pubKeys := sposWorker.Cns.ConsensusGroup()

	selfIndex, err := sposWorker.Cns.IndexSelfConsensusGroup()

	if err != nil {
		log.Info(fmt.Sprintf("%sCanceled round %d in subround %s, NOT IN THE CONSENSUS GROUP\n",
			sposWorker.Cns.getFormattedTime(), sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrStartRound)))

		sposWorker.Cns.Chr.SetSelfSubround(-1)

		return false
	}

	multiSig, err := sposWorker.multiSigner.Create(pubKeys, uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())

		sposWorker.Cns.Chr.SetSelfSubround(-1)

		return false
	}

	sposWorker.multiSigner = multiSig
	sposWorker.Cns.SetStatus(SrStartRound, SsFinished)

	return true
}

// DoEndRoundJob method is the function which actually does the job of the EndRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct,
// from spos package)
func (sposWorker *SPOSConsensusWorker) DoEndRoundJob() bool {
	if !sposWorker.Cns.CheckEndRoundConsensus() {
		return false
	}

	bitmap := sposWorker.genBitmap(SrBitmap)

	err := sposWorker.checkSignaturesValidity(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// Aggregate sig and add it to the block
	sig, err := sposWorker.multiSigner.AggregateSigs(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sposWorker.Header.Signature = sig

	// Commit the block (commits also the account state)
	err = sposWorker.BlockProcessor.CommitBlock(sposWorker.BlockChain, sposWorker.Header, sposWorker.BlockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sposWorker.Cns.SetStatus(SrEndRound, SsFinished)

	err = sposWorker.BlockProcessor.RemoveBlockTxsFromPool(sposWorker.BlockBody)

	if err != nil {
		log.Error(err.Error())
	}

	// broadcast block body
	err = sposWorker.broadcastTxBlockBody(sposWorker.BlockBody)

	if err != nil {
		log.Error(err.Error())
	}

	// broadcast header
	err = sposWorker.broadcastHeader(sposWorker.Header)

	if err != nil {
		log.Error(err.Error())
	}

	log.Info(fmt.Sprintf("%sStep 6: Commiting and broadcasting TxBlockBody and Header\n", sposWorker.Cns.getFormattedTime()))

	if sposWorker.Cns.IsSelfLeaderInCurrentRound() {
		log.Info(fmt.Sprintf("\n%s++++++++++++++++++++ ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN ++++++++++++++++++++\n\n",
			sposWorker.Cns.getFormattedTime(), sposWorker.Header.Nonce))
	} else {
		log.Info(fmt.Sprintf("\n%sxxxxxxxxxxxxxxxxxxxx ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN xxxxxxxxxxxxxxxxxxxx\n\n",
			sposWorker.Cns.getFormattedTime(), sposWorker.Header.Nonce))
	}

	return true
}

// DoBlockJob method actually send the proposed block in the Block subround, when this node is leader
// (it is used as a handler function of the doSubroundJob pointer function declared in Subround struct,
// from spos package)
func (sposWorker *SPOSConsensusWorker) DoBlockJob() bool {
	if sposWorker.boot.ShouldSync() { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
		return false
	}

	isBlockJobDone, err := sposWorker.Cns.GetSelfJobDone(SrBlock)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrBlock) == SsFinished || // is subround Block already finished?
		isBlockJobDone || // has block already been sent?
		!sposWorker.Cns.IsSelfLeaderInCurrentRound() { // is another node leader in this round?
		return false
	}

	if !sposWorker.SendBlockBody() ||
		!sposWorker.SendBlockHeader() {
		return false
	}

	err = sposWorker.Cns.SetSelfJobDone(SrBlock, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sposWorker.multiSigner.SetMessage(sposWorker.Cns.Data)

	return true
}

// SendBlockBody method send the proposed block body in the Block subround
func (sposWorker *SPOSConsensusWorker) SendBlockBody() bool {
	roundIndex := sposWorker.Cns.Chr.Round().Index()
	haveTime := func() bool {
		if roundIndex < sposWorker.Cns.Chr.Round().Index() ||
			sposWorker.GetSubround() > chronology.SubroundId(SrBlock) {
			return false
		}

		return true
	}

	blk, err := sposWorker.BlockProcessor.CreateTxBlockBody(
		shardId,
		maxTransactionsInBlock,
		sposWorker.Cns.Chr.Round().Index(),
		haveTime,
	)

	blkStr, err := sposWorker.marshalizer.Marshal(blk)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := NewConsensusData(
		nil,
		blkStr,
		[]byte(sposWorker.Cns.selfPubKey),
		nil,
		MtBlockBody,
		sposWorker.GetRoundTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Sending block body\n", sposWorker.Cns.getFormattedTime()))

	sposWorker.BlockBody = blk

	return true
}

// GetSubround method returns current subround taking in consideration the current time
func (sposWorker *SPOSConsensusWorker) GetSubround() chronology.SubroundId {
	chr := sposWorker.Cns.Chr
	currentTime := chr.SyncTime().CurrentTime(chr.ClockOffset())

	return chr.GetSubroundFromDateTime(currentTime)
}

// SendBlockHeader method send the proposed block header in the Block subround
func (sposWorker *SPOSConsensusWorker) SendBlockHeader() bool {
	hdr := &block.Header{}

	hdr.Round = uint32(sposWorker.Cns.Chr.Round().Index())
	hdr.TimeStamp = sposWorker.GetRoundTime()

	if sposWorker.BlockChain.CurrentBlockHeader == nil {
		hdr.Nonce = 1
		hdr.PrevHash = sposWorker.BlockChain.GenesisHeaderHash
	} else {
		hdr.Nonce = sposWorker.BlockChain.CurrentBlockHeader.Nonce + 1
		hdr.PrevHash = sposWorker.BlockChain.CurrentBlockHeaderHash
	}

	blkStr, err := sposWorker.marshalizer.Marshal(sposWorker.BlockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdr.BlockBodyHash = sposWorker.hasher.Compute(string(blkStr))

	hdrStr, err := sposWorker.marshalizer.Marshal(hdr)

	hdrHash := sposWorker.hasher.Compute(string(hdrStr))

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := NewConsensusData(
		hdrHash,
		hdrStr,
		[]byte(sposWorker.Cns.SelfPubKey()),
		nil,
		MtBlockHeader,
		sposWorker.GetRoundTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Sending block header with nonce %d and hash %s\n",
		sposWorker.Cns.getFormattedTime(), hdr.Nonce, toB64(hdrHash)))

	sposWorker.Cns.Data = hdrHash
	sposWorker.Header = hdr

	return true
}

func (sposWorker *SPOSConsensusWorker) genCommitmentHash() ([]byte, error) {
	_, commitment := sposWorker.multiSigner.CreateCommitment()

	selfIndex, err := sposWorker.Cns.IndexSelfConsensusGroup()

	if err != nil {
		return nil, err
	}

	commitmentHash := sposWorker.hasher.Compute(string(commitment))

	err = sposWorker.multiSigner.StoreCommitmentHash(uint16(selfIndex), commitmentHash)

	if err != nil {
		return nil, err
	}

	return commitmentHash, nil
}

// DoCommitmentHashJob method is the function which is actually used to send the commitment hash for the received
// block from the leader in the CommitmentHash subround (it is used as the handler function of the doSubroundJob
// pointer variable function in Subround struct, from spos package)
func (sposWorker *SPOSConsensusWorker) DoCommitmentHashJob() bool {
	if sposWorker.Cns.Status(SrBlock) != SsFinished { // is subround Block not finished?
		if !sposWorker.DoBlockJob() {
			return false
		}

		if !sposWorker.Cns.CheckBlockConsensus() {
			return false
		}
	}

	isCommHashJobDone, err := sposWorker.Cns.GetSelfJobDone(SrCommitmentHash)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrCommitmentHash) == SsFinished || // is subround CommitmentHash already finished?
		isCommHashJobDone || // is commitment hash already sent?
		sposWorker.Cns.Data == nil { // is consensus data not set?
		return false
	}

	commitmentHash, err := sposWorker.genCommitmentHash()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := NewConsensusData(
		sposWorker.Cns.Data,
		commitmentHash,
		[]byte(sposWorker.Cns.SelfPubKey()),
		nil,
		MtCommitmentHash,
		sposWorker.GetRoundTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 2: Sending commitment hash\n", sposWorker.Cns.getFormattedTime()))

	err = sposWorker.Cns.SetSelfJobDone(SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func (sposWorker *SPOSConsensusWorker) genBitmap(subround chronology.SubroundId) []byte {
	// generate bitmap according to set commitment hashes
	sizeConsensus := len(sposWorker.Cns.ConsensusGroup())

	bitmap := make([]byte, sizeConsensus/8+1)

	for i := 0; i < sizeConsensus; i++ {
		pubKey := sposWorker.Cns.ConsensusGroup()[i]
		isJobDone, err := sposWorker.Cns.GetJobDone(pubKey, subround)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobDone {
			bitmap[i/8] |= 1 << (uint16(i) % 8)
		}
	}

	return bitmap
}

// DoBitmapJob method is the function which is actually used to send the bitmap with the commitment hashes
// received, in the Bitmap subround, when this node is leader (it is used as the handler function of the
// doSubroundJob pointer variable function in Subround struct, from spos package)
func (sposWorker *SPOSConsensusWorker) DoBitmapJob() bool {
	if sposWorker.Cns.Status(SrCommitmentHash) != SsFinished { // is subround CommitmentHash not finished?
		if !sposWorker.DoCommitmentHashJob() {
			return false
		}

		if !sposWorker.Cns.CheckCommitmentHashConsensus() {
			return false
		}
	}

	isBitmapJobDone, err := sposWorker.Cns.GetSelfJobDone(SrBitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrBitmap) == SsFinished || // is subround Bitmap already finished?
		isBitmapJobDone || // has been bitmap already sent?
		!sposWorker.Cns.IsSelfLeaderInCurrentRound() || // is another node leader in this round?
		sposWorker.Cns.Data == nil { // is consensus data not set?

		return false
	}

	bitmap := sposWorker.genBitmap(SrCommitmentHash)

	dta := NewConsensusData(
		sposWorker.Cns.Data,
		bitmap,
		[]byte(sposWorker.Cns.SelfPubKey()),
		nil,
		MtBitmap,
		sposWorker.GetRoundTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 3: Sending bitmap\n", sposWorker.Cns.getFormattedTime()))

	for i := 0; i < len(sposWorker.Cns.ConsensusGroup()); i++ {
		pubKey := sposWorker.Cns.ConsensusGroup()[i]
		isJobCommHashJobDone, err := sposWorker.Cns.GetJobDone(pubKey, SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobCommHashJobDone {
			err = sposWorker.Cns.SetJobDone(pubKey, SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	sposWorker.Header.PubKeysBitmap = bitmap

	return true
}

// DoCommitmentJob method is the function which is actually used to send the commitment for the received block,
// in the Commitment subround (it is used as the handler function of the doSubroundJob pointer variable function
// in Subround struct, from spos package)
func (sposWorker *SPOSConsensusWorker) DoCommitmentJob() bool {
	if sposWorker.Cns.Status(SrBitmap) != SsFinished { // is subround Bitmap not finished?
		if !sposWorker.DoBitmapJob() {
			return false
		}

		if !sposWorker.Cns.CheckBitmapConsensus() {
			return false
		}
	}

	isCommJobDone, err := sposWorker.Cns.GetSelfJobDone(SrCommitment)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrCommitment) == SsFinished || // is subround Commitment already finished?
		isCommJobDone || // has been commitment already sent?
		!sposWorker.Cns.IsSelfInBitmap() || // isn't node in the leader's bitmap?
		sposWorker.Cns.Data == nil { // is consensus data not set?

		return false
	}

	selfIndex, err := sposWorker.Cns.IndexSelfConsensusGroup()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// commitment
	commitment, err := sposWorker.multiSigner.Commitment(uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := NewConsensusData(
		sposWorker.Cns.Data,
		commitment,
		[]byte(sposWorker.Cns.SelfPubKey()),
		nil,
		MtCommitment,
		sposWorker.GetRoundTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 4: Sending commitment\n", sposWorker.Cns.getFormattedTime()))

	err = sposWorker.Cns.SetSelfJobDone(SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func (sposWorker *SPOSConsensusWorker) checkCommitmentsValidity(bitmap []byte) error {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroup := sposWorker.Cns.ConsensusGroup()
	consensusGroupSize := len(consensusGroup)
	size := consensusGroupSize

	if consensusGroupSize > nbBitsBitmap {
		size = nbBitsBitmap
	}

	for i := 0; i < size; i++ {
		indexRequired := (bitmap[i/8] & (1 << uint16(i%8))) > 0

		if !indexRequired {
			continue
		}

		pubKey := consensusGroup[i]
		isCommJobDone, err := sposWorker.Cns.GetJobDone(pubKey, SrCommitment)

		if err != nil {
			return err
		}

		if !isCommJobDone {
			return ErrNilCommitment
		}

		commitment, err := sposWorker.multiSigner.Commitment(uint16(i))

		if err != nil {
			return err
		}

		computedCommitmentHash := sposWorker.hasher.Compute(string(commitment))
		receivedCommitmentHash, err := sposWorker.multiSigner.CommitmentHash(uint16(i))

		if err != nil {
			return err
		}

		if !bytes.Equal(computedCommitmentHash, receivedCommitmentHash) {
			log.Info(fmt.Sprintf("Commitment %s does not match, expected %s\n",
				toB64(computedCommitmentHash),
				toB64(receivedCommitmentHash)))

			return ErrCommitmentHashDoesNotMatch
		}
	}

	return nil
}

// DoSignatureJob method is the function which is actually used to send the Signature for the received block,
// in the Signature subround (it is used as the handler function of the doSubroundJob pointer variable function
// in Subround struct, from spos package)
func (sposWorker *SPOSConsensusWorker) DoSignatureJob() bool {
	if sposWorker.Cns.Status(SrCommitment) != SsFinished { // is subround Commitment not finished?
		if !sposWorker.DoCommitmentJob() {
			return false
		}

		if !sposWorker.Cns.CheckCommitmentConsensus() {
			return false
		}
	}

	isSignJobDone, err := sposWorker.Cns.GetSelfJobDone(SrSignature)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrSignature) == SsFinished || // is subround Signature already finished?
		isSignJobDone || // has been signature already sent?
		!sposWorker.Cns.IsSelfInBitmap() || // isn't node in the leader's bitmap?
		sposWorker.Cns.Data == nil { // is consensus data not set?

		return false
	}

	bitmap := sposWorker.genBitmap(SrBitmap)

	err = sposWorker.checkCommitmentsValidity(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// first compute commitment aggregation
	aggComm, err := sposWorker.multiSigner.AggregateCommitments(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sigPart, err := sposWorker.multiSigner.CreateSignatureShare(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := NewConsensusData(
		sposWorker.Cns.Data,
		sigPart,
		[]byte(sposWorker.Cns.SelfPubKey()),
		nil,
		MtSignature,
		sposWorker.GetRoundTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 5: Sending signature\n", sposWorker.Cns.getFormattedTime()))

	selfIndex, err := sposWorker.Cns.IndexSelfConsensusGroup()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sposWorker.multiSigner.StoreSignatureShare(uint16(selfIndex), sigPart)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sposWorker.Cns.SetSelfJobDone(SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sposWorker.Header.Commitment = aggComm

	return true
}

// DoAdvanceJob method is the function which actually does the job of the Advance subround (it is used as the handler
// function of the doSubroundJob pointer variable function in Subround struct, from spos package)
func (sposWorker *SPOSConsensusWorker) DoAdvanceJob() bool {
	if sposWorker.Cns.Status(SrEndRound) == SsFinished {
		return false
	}

	sposWorker.BlockProcessor.RevertAccountState()

	log.Info(fmt.Sprintf("%sStep 7: Creating and broadcasting an empty block\n", sposWorker.Cns.getFormattedTime()))

	sposWorker.createEmptyBlock()

	return true
}

func (sposWorker *SPOSConsensusWorker) genConsensusDataSignature(cnsDta *ConsensusData) ([]byte, error) {

	cnsDtaStr, err := sposWorker.marshalizer.Marshal(cnsDta)

	if err != nil {
		return nil, err
	}

	signature, err := sposWorker.singleSigner.Sign(sposWorker.privKey, cnsDtaStr)

	if err != nil {
		return nil, err
	}

	return signature, nil
}

// SendConsensusMessage sends the consensus message
func (sposWorker *SPOSConsensusWorker) SendConsensusMessage(cnsDta *ConsensusData) bool {
	signature, err := sposWorker.genConsensusDataSignature(cnsDta)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	signedCnsData := *cnsDta
	signedCnsData.Signature = signature

	if sposWorker.SendMessage == nil {
		log.Error("SendMessage call back function is not set\n")
		return false
	}

	go sposWorker.SendMessage(&signedCnsData)

	return true
}

func (sposWorker *SPOSConsensusWorker) broadcastTxBlockBody(blockBody *block.TxBlockBody) error {
	if blockBody == nil {
		return ErrNilTxBlockBody
	}

	message, err := sposWorker.marshalizer.Marshal(blockBody)

	if err != nil {
		return err
	}

	// send message
	if sposWorker.BroadcastBlockBody == nil {
		return ErrNilOnBroadcastTxBlockBody
	}

	go sposWorker.BroadcastBlockBody(message)

	return nil
}

func (sposWorker *SPOSConsensusWorker) broadcastHeader(header *block.Header) error {
	if header == nil {
		return ErrNilBlockHeader
	}

	message, err := sposWorker.marshalizer.Marshal(header)

	if err != nil {
		return err
	}

	// send message
	if sposWorker.BroadcastHeader == nil {
		return ErrNilOnBroadcastHeader
	}

	go sposWorker.BroadcastHeader(message)

	return nil
}

// ExtendStartRound method just call the DoStartRoundJob method to be sure that the init will be done
func (sposWorker *SPOSConsensusWorker) ExtendStartRound() {
	sposWorker.Cns.SetStatus(SrStartRound, SsExtended)

	log.Info(fmt.Sprintf("%sStep 0: Extended the (START_ROUND) subround\n", sposWorker.Cns.getFormattedTime()))

	sposWorker.DoStartRoundJob()
}

// ExtendBlock method put this subround in the extended mode and print some messages
func (sposWorker *SPOSConsensusWorker) ExtendBlock() {
	if sposWorker.boot.ShouldSync() {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, NOT SYNCRONIZED YET\n",
			sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBlock)))

		sposWorker.Cns.Chr.SetSelfSubround(-1)

		return
	}

	sposWorker.Cns.SetStatus(SrBlock, SsExtended)

	log.Info(fmt.Sprintf("%sStep 1: Extended the (BLOCK) subround\n", sposWorker.Cns.getFormattedTime()))
}

// ExtendCommitmentHash method put this subround in the extended mode and print some messages
func (sposWorker *SPOSConsensusWorker) ExtendCommitmentHash() {
	sposWorker.Cns.SetStatus(SrCommitmentHash, SsExtended)

	if sposWorker.Cns.ComputeSize(SrCommitmentHash) < sposWorker.Cns.Threshold(SrCommitmentHash) {
		log.Info(fmt.Sprintf("%sStep 2: Extended the (COMMITMENT_HASH) subround. Got only %d from %d commitment hashes which are not enough\n",
			sposWorker.Cns.getFormattedTime(), sposWorker.Cns.ComputeSize(SrCommitmentHash), len(sposWorker.Cns.ConsensusGroup())))
	} else {
		log.Info(fmt.Sprintf("%sStep 2: Extended the (COMMITMENT_HASH) subround\n", sposWorker.Cns.getFormattedTime()))
	}
}

// ExtendBitmap method put this subround in the extended mode and print some messages
func (sposWorker *SPOSConsensusWorker) ExtendBitmap() {
	sposWorker.Cns.SetStatus(SrBitmap, SsExtended)

	log.Info(fmt.Sprintf("%sStep 3: Extended the (BITMAP) subround\n", sposWorker.Cns.getFormattedTime()))
}

// ExtendCommitment method put this subround in the extended mode and print some messages
func (sposWorker *SPOSConsensusWorker) ExtendCommitment() {
	sposWorker.Cns.SetStatus(SrCommitment, SsExtended)

	log.Info(fmt.Sprintf("%sStep 4: Extended the (COMMITMENT) subround. Got only %d from %d commitments which are not enough\n",
		sposWorker.Cns.getFormattedTime(), sposWorker.Cns.ComputeSize(SrCommitment), len(sposWorker.Cns.ConsensusGroup())))
}

// ExtendSignature method put this subround in the extended mode and print some messages
func (sposWorker *SPOSConsensusWorker) ExtendSignature() {
	sposWorker.Cns.SetStatus(SrSignature, SsExtended)

	log.Info(fmt.Sprintf("%sStep 5: Extended the (SIGNATURE) subround. Got only %d from %d signatures which are not enough\n",
		sposWorker.Cns.getFormattedTime(), sposWorker.Cns.ComputeSize(SrSignature), len(sposWorker.Cns.ConsensusGroup())))
}

// ExtendEndRound method just print some messages as no extend will be permited, because a new round will be start
func (sposWorker *SPOSConsensusWorker) ExtendEndRound() {
	sposWorker.Cns.SetStatus(SrEndRound, SsExtended)

	log.Info(fmt.Sprintf("%sStep 6: Extended the (END_ROUND) subround\n", sposWorker.Cns.getFormattedTime()))
}

// createEmptyBlock creates, commits and broadcasts an empty block at the end of the round if no block was proposed or
// syncronized in this round
func (sposWorker *SPOSConsensusWorker) createEmptyBlock() bool {
	blk := sposWorker.BlockProcessor.CreateEmptyBlockBody(
		shardId,
		sposWorker.Cns.Chr.Round().Index())

	hdr := &block.Header{}
	hdr.Round = uint32(sposWorker.Cns.Chr.Round().Index())
	hdr.TimeStamp = sposWorker.GetRoundTime()

	var prevHeaderHash []byte

	if sposWorker.BlockChain.CurrentBlockHeader == nil {
		hdr.Nonce = 1
		prevHeaderHash = sposWorker.BlockChain.GenesisHeaderHash
	} else {
		hdr.Nonce = sposWorker.BlockChain.CurrentBlockHeader.Nonce + 1
		prevHeaderHash = sposWorker.BlockChain.CurrentBlockHeaderHash
	}

	hdr.PrevHash = prevHeaderHash
	blkStr, err := sposWorker.marshalizer.Marshal(blk)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	hdr.BlockBodyHash = sposWorker.hasher.Compute(string(blkStr))

	cnsGroup := sposWorker.Cns.ConsensusGroup()
	cnsGroupSize := len(cnsGroup)

	hdr.PubKeysBitmap = make([]byte, cnsGroupSize/8+1)

	// TODO: decide the signature for the empty block
	headerStr, err := sposWorker.marshalizer.Marshal(hdr)
	hdrHash := sposWorker.hasher.Compute(string(headerStr))
	hdr.Signature = hdrHash
	hdr.Commitment = hdrHash

	// Commit the block (commits also the account state)
	err = sposWorker.BlockProcessor.CommitBlock(sposWorker.BlockChain, hdr, blk)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	// broadcast block body
	err = sposWorker.broadcastTxBlockBody(blk)

	if err != nil {
		log.Info(err.Error())
	}

	// broadcast header
	err = sposWorker.broadcastHeader(hdr)

	if err != nil {
		log.Info(err.Error())
	}

	log.Info(fmt.Sprintf("\n%s******************** ADDED EMPTY BLOCK WITH NONCE  %d  IN BLOCKCHAIN ********************\n\n",
		sposWorker.Cns.getFormattedTime(), hdr.Nonce))

	return true
}

// ReceivedMessage method redirects the received message to the channel which should handle it
func (sposWorker *SPOSConsensusWorker) ReceivedMessage(name string, data interface{}, msgInfo *p2p.MessageInfo) {
	if sposWorker.Cns.Chr.IsCancelled() {
		return
	}

	cnsDta, ok := data.(*ConsensusData)

	if !ok {
		return
	}

	senderOK := sposWorker.Cns.IsNodeInConsensusGroup(string(cnsDta.PubKey))

	if !senderOK {
		return
	}

	if sposWorker.shouldDropConsensusMessage(cnsDta) {
		return
	}

	if sposWorker.Cns.SelfPubKey() == string(cnsDta.PubKey) {
		return
	}

	sigVerifErr := sposWorker.checkSignature(cnsDta)
	if sigVerifErr != nil {
		return
	}

	sposWorker.ReceivedMessageChannel <- cnsDta
}

func (sposWorker *SPOSConsensusWorker) shouldDropConsensusMessage(cnsDta *ConsensusData) bool {
	if cnsDta.RoundIndex < sposWorker.Cns.Chr.Round().Index() {
		return true
	}

	if cnsDta.RoundIndex == sposWorker.Cns.Chr.Round().Index() &&
		sposWorker.GetSubround() > chronology.SubroundId(SrEndRound) {
		return true
	}

	return false
}

// CheckChannels method is used to listen to the channels through which node receives and consumes,
// during the round, different messages from the nodes which are in the validators group
func (sposWorker *SPOSConsensusWorker) checkChannels() {
	for {
		select {
		case rcvDta := <-sposWorker.MessageChannels[MtBlockBody]:
			if sposWorker.ReceivedBlockBody(rcvDta) {
				sposWorker.Cns.CheckBlockConsensus()
			}
		case rcvDta := <-sposWorker.MessageChannels[MtBlockHeader]:
			if sposWorker.ReceivedBlockHeader(rcvDta) {
				sposWorker.Cns.CheckBlockConsensus()
			}
		case rcvDta := <-sposWorker.MessageChannels[MtCommitmentHash]:
			if sposWorker.ReceivedCommitmentHash(rcvDta) {
				sposWorker.Cns.CheckCommitmentHashConsensus()
			}
		case rcvDta := <-sposWorker.MessageChannels[MtBitmap]:
			if sposWorker.ReceivedBitmap(rcvDta) {
				sposWorker.Cns.CheckBitmapConsensus()
			}
		case rcvDta := <-sposWorker.MessageChannels[MtCommitment]:
			if sposWorker.ReceivedCommitment(rcvDta) {
				sposWorker.Cns.CheckCommitmentConsensus()
			}
		case rcvDta := <-sposWorker.MessageChannels[MtSignature]:
			if sposWorker.ReceivedSignature(rcvDta) {
				sposWorker.Cns.CheckSignatureConsensus()
			}
		}
	}
}

func (sposWorker *SPOSConsensusWorker) checkSignature(cnsData *ConsensusData) error {
	if cnsData == nil {
		return ErrNilConsensusData
	}

	if cnsData.PubKey == nil {
		return ErrNilPublicKey
	}

	if cnsData.Signature == nil {
		return ErrNilSignature
	}

	pubKey, err := sposWorker.keyGen.PublicKeyFromByteArray(cnsData.PubKey)

	if err != nil {
		return err
	}

	dataNoSig := *cnsData
	signature := cnsData.Signature

	dataNoSig.Signature = nil
	dataNoSigString, err := sposWorker.marshalizer.Marshal(dataNoSig)

	if err != nil {
		return err
	}

	err = sposWorker.singleSigner.Verify(pubKey, dataNoSigString, signature)

	return err
}

// ReceivedBlockBody method is called when a block body is received through the block body channel.
func (sposWorker *SPOSConsensusWorker) ReceivedBlockBody(cnsDta *ConsensusData) bool {
	node := string(cnsDta.PubKey)

	isBlockJobDone, err := sposWorker.Cns.RoundConsensus.GetJobDone(node, SrBlock)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if node == sposWorker.Cns.SelfPubKey() || // is block header received from myself?
		sposWorker.Cns.Status(SrBlock) == SsFinished || // is subround Block already finished?
		!sposWorker.Cns.IsNodeLeaderInCurrentRound(node) || // is another node leader in this round?
		isBlockJobDone || // is block job of this node already done?
		sposWorker.BlockBody != nil || // is block body already received?
		cnsDta.RoundIndex != sposWorker.Cns.Chr.Round().Index() || // is this the consesnus data of this round?
		sposWorker.GetSubround() > chronology.SubroundId(SrEndRound) { // is message received too late in this round?

		return false
	}

	sposWorker.BlockBody = sposWorker.DecodeBlockBody(cnsDta.SubRoundData)

	if sposWorker.BlockBody == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Received block body\n", sposWorker.Cns.getFormattedTime()))

	blockProcessedWithSuccess := sposWorker.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// DecodeBlockBody method decodes block body which is marshalized in the received message
func (sposWorker *SPOSConsensusWorker) DecodeBlockBody(dta []byte) *block.TxBlockBody {
	if dta == nil {
		return nil
	}

	var blk block.TxBlockBody

	err := sposWorker.marshalizer.Unmarshal(&blk, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &blk
}

// ReceivedBlockHeader method is called when a block header is received through the block header channel.
// If the block header is valid, than the validatorRoundStates map coresponding to the node which sent it,
// is set on true for the subround Block
func (sposWorker *SPOSConsensusWorker) ReceivedBlockHeader(cnsDta *ConsensusData) bool {
	node := string(cnsDta.PubKey)

	isBlockJobDone, err := sposWorker.Cns.RoundConsensus.GetJobDone(node, SrBlock)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if node == sposWorker.Cns.SelfPubKey() || // is block header received from myself?
		sposWorker.Cns.Status(SrBlock) == SsFinished || // is subround Block already finished?
		!sposWorker.Cns.IsNodeLeaderInCurrentRound(node) || // is another node leader in this round?
		isBlockJobDone || // is block job of this node already done?
		sposWorker.Header != nil || // is block header already received?
		sposWorker.Cns.Data != nil || // is consensus data already set?
		cnsDta.RoundIndex != sposWorker.Cns.Chr.Round().Index() || // is this the consesnus data of this round?
		sposWorker.GetSubround() > chronology.SubroundId(SrEndRound) { // is message received too late in this round?

		return false
	}

	sposWorker.Cns.Data = cnsDta.BlockHeaderHash
	sposWorker.Header = sposWorker.DecodeBlockHeader(cnsDta.SubRoundData)

	if sposWorker.Header == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Received block header with nonce %d and hash %s\n",
		sposWorker.Cns.getFormattedTime(), sposWorker.Header.Nonce, toB64(cnsDta.BlockHeaderHash)))

	if !sposWorker.CheckIfBlockIsValid(sposWorker.Header) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, INVALID BLOCK\n",
			sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBlock)))

		return false
	}

	blockProcessedWithSuccess := sposWorker.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

func (sposWorker *SPOSConsensusWorker) processReceivedBlock(cnsDta *ConsensusData) bool {

	if sposWorker.BlockBody == nil ||
		sposWorker.Header == nil {
		return false
	}

	node := string(cnsDta.PubKey)

	err := sposWorker.BlockProcessor.ProcessBlock(sposWorker.BlockChain, sposWorker.Header, sposWorker.BlockBody, sposWorker.haveTime)

	if err != nil {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, %s\n",
			sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBlock), err.Error()))

		return false
	}

	subround := sposWorker.GetSubround()

	if cnsDta.RoundIndex != sposWorker.Cns.Chr.Round().Index() {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, CURRENT ROUND IS %d\n",
			cnsDta.RoundIndex, sposWorker.Cns.GetSubroundName(SrBlock), sposWorker.Cns.Chr.Round().Index()))

		sposWorker.BlockProcessor.RevertAccountState()

		return false
	}

	if subround > chronology.SubroundId(SrEndRound) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, CURRENT SUBROUND IS %s\n",
			cnsDta.RoundIndex, sposWorker.Cns.GetSubroundName(SrBlock), sposWorker.Cns.GetSubroundName(subround)))

		sposWorker.BlockProcessor.RevertAccountState()

		return false
	}

	sposWorker.multiSigner.SetMessage(sposWorker.Cns.Data)
	err = sposWorker.Cns.RoundConsensus.SetJobDone(node, SrBlock, true)

	if err != nil {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, %s\n",
			sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBlock), err.Error()))

		return false
	}

	return true
}

// DecodeBlockHeader method decodes block header which is marshalized in the received message
func (sposWorker *SPOSConsensusWorker) DecodeBlockHeader(dta []byte) *block.Header {
	if dta == nil {
		return nil
	}

	var hdr block.Header

	err := sposWorker.marshalizer.Unmarshal(&hdr, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &hdr
}

// ReceivedCommitmentHash method is called when a commitment hash is received through the commitment hash
// channel. If the commitment hash is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround ComitmentHash
func (sposWorker *SPOSConsensusWorker) ReceivedCommitmentHash(cnsDta *ConsensusData) bool {
	node := string(cnsDta.PubKey)

	isCommHashJobDone, err := sposWorker.Cns.RoundConsensus.GetJobDone(node, SrCommitmentHash)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if node == sposWorker.Cns.SelfPubKey() || // is commitment hash received from myself?
		sposWorker.Cns.Status(SrCommitmentHash) == SsFinished || // is subround CommitmentHash already finished?
		!sposWorker.Cns.IsNodeInConsensusGroup(node) || // isn't node in the consensus group?
		isCommHashJobDone || // is commitment hash job of this node already done?
		sposWorker.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.BlockHeaderHash, sposWorker.Cns.Data) || // is this the consesnus data of this round?
		sposWorker.GetSubround() > chronology.SubroundId(SrEndRound) { // is message received too late in this round?

		return false
	}

	// if this node is leader in this round and already he received 2/3 + 1 of commitment hashes
	// he will ignore any others received later
	if sposWorker.Cns.IsSelfLeaderInCurrentRound() {
		threshold := sposWorker.Cns.Threshold(SrCommitmentHash)
		if sposWorker.Cns.IsCommitmentHashReceived(threshold) {
			return false
		}
	}

	index, err := sposWorker.Cns.ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sposWorker.multiSigner.StoreCommitmentHash(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sposWorker.Cns.RoundConsensus.SetJobDone(node, SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func countBitmapFlags(bitmap []byte) uint16 {
	nbBytes := len(bitmap)
	flags := 0
	for i := 0; i < nbBytes; i++ {
		for j := 0; j < 8; j++ {
			if bitmap[i]&(1<<uint8(j)) != 0 {
				flags++
			}
		}
	}
	return uint16(flags)
}

// ReceivedBitmap method is called when a bitmap is received through the bitmap channel.
// If the bitmap is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Bitmap
func (sposWorker *SPOSConsensusWorker) ReceivedBitmap(cnsDta *ConsensusData) bool {
	node := string(cnsDta.PubKey)

	isBitmapJobDone, err := sposWorker.Cns.RoundConsensus.GetJobDone(node, SrBitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if node == sposWorker.Cns.SelfPubKey() || // is bitmap received from myself?
		sposWorker.Cns.Status(SrBitmap) == SsFinished || // is subround Bitmap already finished?
		!sposWorker.Cns.IsNodeLeaderInCurrentRound(node) || // is another node leader in this round?
		isBitmapJobDone || // is bitmap job of this node already done?
		sposWorker.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.BlockHeaderHash, sposWorker.Cns.Data) || // is this the consesnus data of this round?
		sposWorker.GetSubround() > chronology.SubroundId(SrEndRound) { // is message received too late in this round?

		return false
	}

	signersBitmap := cnsDta.SubRoundData

	// count signers
	nbSigners := countBitmapFlags(signersBitmap)

	if int(nbSigners) < sposWorker.Cns.Threshold(SrBitmap) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, TOO FEW SIGNERS IN BITMAP\n",
			sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBitmap)))

		return false
	}

	publicKeys := sposWorker.Cns.ConsensusGroup()

	for i := 0; i < len(publicKeys); i++ {
		byteNb := i / 8
		bitNb := i % 8
		isNodeSigner := (signersBitmap[byteNb] & (1 << uint8(bitNb))) != 0

		if isNodeSigner {
			err = sposWorker.Cns.RoundConsensus.SetJobDone(publicKeys[i], SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	if !sposWorker.Cns.IsValidatorInBitmap(sposWorker.Cns.selfPubKey) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, NOT INCLUDED IN THE BITMAP\n",
			sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBitmap)))

		sposWorker.Cns.Chr.SetSelfSubround(-1)

		sposWorker.BlockProcessor.RevertAccountState()

		return false
	}

	sposWorker.Header.PubKeysBitmap = signersBitmap

	return true
}

// ReceivedCommitment method is called when a commitment is received through the commitment channel.
// If the commitment is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Comitment
func (sposWorker *SPOSConsensusWorker) ReceivedCommitment(cnsDta *ConsensusData) bool {
	node := string(cnsDta.PubKey)

	isCommJobDone, err := sposWorker.Cns.RoundConsensus.GetJobDone(node, SrCommitment)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if node == sposWorker.Cns.SelfPubKey() || // is commitment received from myself?
		sposWorker.Cns.Status(SrCommitment) == SsFinished || // is subround Commitment already finished?
		!sposWorker.Cns.IsValidatorInBitmap(node) || // isn't node in the bitmap group?
		isCommJobDone || // is commitment job of this node already done?
		sposWorker.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.BlockHeaderHash, sposWorker.Cns.Data) || // is this the consesnus data of this round?
		sposWorker.GetSubround() > chronology.SubroundId(SrEndRound) { // is message received too late in this round?

		return false
	}

	index, err := sposWorker.Cns.ConsensusGroupIndex(node)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = sposWorker.multiSigner.StoreCommitment(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = sposWorker.Cns.RoundConsensus.SetJobDone(node, SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// ReceivedSignature method is called when a Signature is received through the Signature channel.
// If the Signature is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Signature
func (sposWorker *SPOSConsensusWorker) ReceivedSignature(cnsDta *ConsensusData) bool {
	node := string(cnsDta.PubKey)

	isSignJobDone, err := sposWorker.Cns.RoundConsensus.GetJobDone(node, SrSignature)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if node == sposWorker.Cns.SelfPubKey() || // is signature received from myself?
		sposWorker.Cns.Status(SrSignature) == SsFinished || // is subround Signature already finished?
		!sposWorker.Cns.IsValidatorInBitmap(node) || // isn't node in the bitmap group?
		isSignJobDone || // is signature job of this node already done?
		sposWorker.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.BlockHeaderHash, sposWorker.Cns.Data) || // is this the consesnus data of this round?
		sposWorker.GetSubround() > chronology.SubroundId(SrEndRound) { // is message received too late in this round?

		return false
	}

	index, err := sposWorker.Cns.ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sposWorker.multiSigner.StoreSignatureShare(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sposWorker.Cns.RoundConsensus.SetJobDone(node, SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// CheckIfBlockIsValid method checks if the received block is valid
func (sposWorker *SPOSConsensusWorker) CheckIfBlockIsValid(receivedHeader *block.Header) bool {
	if sposWorker.BlockChain.CurrentBlockHeader == nil {
		if receivedHeader.Nonce == 1 { // first block after genesis
			if bytes.Equal(receivedHeader.PrevHash, sposWorker.BlockChain.GenesisHeaderHash) {
				return true
			}

			log.Info(fmt.Sprintf("Hash not match: local block hash is empty and node received block with previous hash %s\n",
				toB64(receivedHeader.PrevHash)))

			return false
		}

		log.Info(fmt.Sprintf("Nonce not match: local block nonce is 0 and node received block with nonce %d\n",
			receivedHeader.Nonce))

		return false
	}

	if receivedHeader.Nonce < sposWorker.BlockChain.CurrentBlockHeader.Nonce+1 {
		log.Info(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d\n",
			sposWorker.BlockChain.CurrentBlockHeader.Nonce, receivedHeader.Nonce))

		return false
	}

	if receivedHeader.Nonce == sposWorker.BlockChain.CurrentBlockHeader.Nonce+1 {
		prevHeaderHash := sposWorker.getHeaderHash(sposWorker.BlockChain.CurrentBlockHeader)

		if bytes.Equal(receivedHeader.PrevHash, prevHeaderHash) {
			return true
		}

		log.Info(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s\n",
			toB64(prevHeaderHash), toB64(receivedHeader.PrevHash)))

		return false
	}

	log.Info(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d\n",
		sposWorker.BlockChain.CurrentBlockHeader.Nonce, receivedHeader.Nonce))

	return false
}

// GetMessageTypeName method returns the name of the message from a given message ID
func (sposWorker *SPOSConsensusWorker) GetMessageTypeName(messageType MessageType) string {
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

// GetRoundTime method returns time stamp of the current round
func (sposWorker *SPOSConsensusWorker) GetRoundTime() uint64 {
	chr := sposWorker.Cns.Chr
	currentRoundIndex := chr.Round().Index()

	return chr.RoundTimeStamp(currentRoundIndex)
}

// CheckEndRoundConsensus method checks if the consensus is achieved in each subround from first subround to the given
// subround. If the consensus is achieved in one subround, the subround status is marked as finished
func (cns *Consensus) CheckEndRoundConsensus() bool {
	for i := SrBlock; i <= SrSignature; i++ {
		currentSubRound := cns.Chr.SubroundHandlers()[i]
		if !currentSubRound.Check() {
			return false
		}
	}

	return true
}

// CheckStartRoundConsensus method checks if the consensus is achieved in the start subround.
func (cns *Consensus) CheckStartRoundConsensus() bool {
	return !cns.Chr.IsCancelled()
}

// CheckBlockConsensus method checks if the consensus in the <BLOCK> subround is achieved
func (cns *Consensus) CheckBlockConsensus() bool {
	cns.mut.Lock()
	defer cns.mut.Unlock()

	if cns.Chr.IsCancelled() {
		return false
	}

	if cns.Status(SrBlock) == SsFinished {
		return true
	}

	if cns.IsBlockReceived(cns.Threshold(SrBlock)) {
		cns.PrintBlockCM() // only for printing block consensus messages
		cns.SetStatus(SrBlock, SsFinished)

		return true
	}

	return false
}

// CheckCommitmentHashConsensus method checks if the consensus in the <COMMITMENT_HASH> subround is achieved
func (cns *Consensus) CheckCommitmentHashConsensus() bool {
	cns.mut.Lock()
	defer cns.mut.Unlock()

	if cns.Chr.IsCancelled() {
		return false
	}

	if cns.Status(SrCommitmentHash) == SsFinished {
		return true
	}

	threshold := cns.Threshold(SrCommitmentHash)

	if !cns.IsSelfLeaderInCurrentRound() {
		threshold = len(cns.consensusGroup)
	}

	if cns.IsCommitmentHashReceived(threshold) {
		cns.PrintCommitmentHashCM() // only for printing commitment hash consensus messages
		cns.SetStatus(SrCommitmentHash, SsFinished)

		return true
	}

	if cns.CommitmentHashesCollected(cns.Threshold(SrBitmap)) {
		cns.PrintCommitmentHashCM() // only for printing commitment hash consensus messages
		cns.SetStatus(SrCommitmentHash, SsFinished)

		return true
	}

	return false
}

// CheckBitmapConsensus method checks if the consensus in the <BITMAP> subround is achieved
func (cns *Consensus) CheckBitmapConsensus() bool {
	cns.mut.Lock()
	defer cns.mut.Unlock()

	if cns.Chr.IsCancelled() {
		return false
	}

	if cns.Status(SrBitmap) == SsFinished {
		return true
	}

	if cns.CommitmentHashesCollected(cns.Threshold(SrBitmap)) {
		cns.PrintBitmapCM() // only for printing bitmap consensus messages
		cns.SetStatus(SrBitmap, SsFinished)

		return true
	}

	return false
}

// CheckCommitmentConsensus method checks if the consensus in the <COMMITMENT> subround is achieved
func (cns *Consensus) CheckCommitmentConsensus() bool {
	cns.mut.Lock()
	defer cns.mut.Unlock()

	if cns.Chr.IsCancelled() {
		return false
	}

	if cns.Status(SrCommitment) == SsFinished {
		return true
	}

	if cns.CommitmentsCollected(cns.Threshold(SrCommitment)) {
		cns.PrintCommitmentCM() // only for printing commitment consensus messages
		cns.SetStatus(SrCommitment, SsFinished)

		return true
	}

	return false
}

// CheckSignatureConsensus method checks if the consensus in the <SIGNATURE> subround is achieved
func (cns *Consensus) CheckSignatureConsensus() bool {
	cns.mut.Lock()
	defer cns.mut.Unlock()

	if cns.Chr.IsCancelled() {
		return false
	}

	if cns.Status(SrSignature) == SsFinished {
		return true
	}

	if cns.SignaturesCollected(cns.Threshold(SrSignature)) {
		cns.PrintSignatureCM() // only for printing signature consensus messages
		cns.SetStatus(SrSignature, SsFinished)

		return true
	}

	return false
}

// CheckAdvanceConsensus method checks if the consensus is achieved in the advance subround.
func (cns *Consensus) CheckAdvanceConsensus() bool {
	return true
}

// GetSubroundName returns the name of each subround from a given subround ID
func (cns *Consensus) GetSubroundName(subroundId chronology.SubroundId) string {
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
	case SrAdvance:
		return "(ADVANCE)"
	default:
		return "Undefined subround"
	}
}

// PrintBlockCM method prints the <BLOCK> consensus messages
func (cns *Consensus) PrintBlockCM() {
	if !cns.IsSelfLeaderInCurrentRound() {
		log.Info(fmt.Sprintf("%sStep 1: Synchronized block\n", cns.getFormattedTime()))
	}

	log.Info(fmt.Sprintf("%sStep 1: Subround (BLOCK) has been finished\n", cns.getFormattedTime()))
}

// PrintCommitmentHashCM method prints the <COMMITMENT_HASH> consensus messages
func (cns *Consensus) PrintCommitmentHashCM() {
	n := cns.ComputeSize(SrCommitmentHash)

	if n == len(cns.consensusGroup) {
		log.Info(fmt.Sprintf("%sStep 2: Received all (%d from %d) commitment hashes\n",
			cns.getFormattedTime(), n, len(cns.consensusGroup)))
	} else {
		log.Info(fmt.Sprintf("%sStep 2: Received %d from %d commitment hashes, which are enough\n",
			cns.getFormattedTime(), n, len(cns.consensusGroup)))
	}

	log.Info(fmt.Sprintf("%sStep 2: Subround (COMMITMENT_HASH) has been finished\n", cns.getFormattedTime()))
}

// PrintBitmapCM method prints the <BITMAP> consensus messages
func (cns *Consensus) PrintBitmapCM() {
	if !cns.IsSelfLeaderInCurrentRound() {
		msg := fmt.Sprintf("%sStep 3: Received bitmap from leader, matching with my own, and it got %d from %d commitment hashes, which are enough",
			cns.getFormattedTime(), cns.ComputeSize(SrBitmap), len(cns.consensusGroup))

		if cns.IsValidatorInBitmap(cns.selfPubKey) {
			msg = fmt.Sprintf("%s, AND I WAS selected in this bitmap\n", msg)
		} else {
			msg = fmt.Sprintf("%s, BUT I WAS NOT selected in this bitmap\n", msg)
		}

		log.Info(msg)
	}

	log.Info(fmt.Sprintf("%sStep 3: Subround (BITMAP) has been finished\n", cns.getFormattedTime()))
}

// PrintCommitmentCM method prints the <COMMITMENT> consensus messages
func (cns *Consensus) PrintCommitmentCM() {
	log.Info(fmt.Sprintf("%sStep 4: Received %d from %d commitments, which are matching with bitmap and are enough\n",
		cns.getFormattedTime(), cns.ComputeSize(SrCommitment), len(cns.consensusGroup)))

	log.Info(fmt.Sprintf("%sStep 4: Subround (COMMITMENT) has been finished\n", cns.getFormattedTime()))
}

// PrintSignatureCM method prints the <SIGNATURE> consensus messages
func (cns *Consensus) PrintSignatureCM() {
	log.Info(fmt.Sprintf("%sStep 5: Received %d from %d signatures, which are matching with bitmap and are enough\n",
		cns.getFormattedTime(), cns.ComputeSize(SrSignature), len(cns.consensusGroup)))

	log.Info(fmt.Sprintf("%sStep 5: Subround (SIGNATURE) has been finished\n", cns.getFormattedTime()))
}

func (sposWorker *SPOSConsensusWorker) getHeaderHash(hdr *block.Header) []byte {
	headerMarsh, err := sposWorker.marshalizer.Marshal(hdr)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return sposWorker.hasher.Compute(string(headerMarsh))
}

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}

	return base64.StdEncoding.EncodeToString(buff)
}

func (sposWorker *SPOSConsensusWorker) haveTime() time.Duration {
	chr := sposWorker.Cns.Chr

	roundStartTime := chr.Round().TimeStamp()
	currentTime := chr.SyncTime().CurrentTime(chr.ClockOffset())
	elapsedTime := currentTime.Sub(roundStartTime)
	haveTime := float64(chr.Round().TimeDuration())*maxBlockProcessingTimePercent - float64(elapsedTime)

	return time.Duration(haveTime)
}
