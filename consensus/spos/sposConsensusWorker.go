package spos

import (
	"bytes"
	"encoding/base64"
	"fmt"

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
const maxTransactionsInBlock = 1000

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

// ConsensusData defines the data needed by spos to comunicate between nodes over network in all subrounds
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

// SPOSConsensusWorker defines the data needed by spos to comunicate between nodes which are in the validators group
type SPOSConsensusWorker struct {
	Cns             *Consensus
	Header          *block.Header
	BlockBody       *block.TxBlockBody
	BlockChain      *blockchain.BlockChain
	Rounds          int // only for statistic
	RoundsWithBlock int // only for statistic
	BlockProcessor  process.BlockProcessor
	MessageChannels map[MessageType]chan *ConsensusData
	hasher          hashing.Hasher
	marshalizer     marshal.Marshalizer
	keyGen          crypto.KeyGenerator
	privKey         crypto.PrivateKey
	pubKey          crypto.PublicKey
	multiSigner     crypto.MultiSigner
	// this is a pointer to a function which actually send the message from a node to the network
	SendMessage        func(consensus *ConsensusData)
	BroadcastHeader    func([]byte)
	BroadcastBlockBody func([]byte)
}

// NewConsensusWorker creates a new SPOSConsensusWorker object
func NewConsensusWorker(
	cns *Consensus,
	blkc *blockchain.BlockChain,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockProcessor process.BlockProcessor,
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

	sposWorker.MessageChannels[MtBlockBody] = make(chan *ConsensusData, nodes)
	sposWorker.MessageChannels[MtBlockHeader] = make(chan *ConsensusData, nodes)
	sposWorker.MessageChannels[MtCommitmentHash] = make(chan *ConsensusData, nodes)
	sposWorker.MessageChannels[MtBitmap] = make(chan *ConsensusData, nodes)
	sposWorker.MessageChannels[MtCommitment] = make(chan *ConsensusData, nodes)
	sposWorker.MessageChannels[MtSignature] = make(chan *ConsensusData, nodes)

	go sposWorker.CheckChannels()

	return &sposWorker, nil
}

func checkNewConsensusWorkerParams(
	cns *Consensus,
	blkc *blockchain.BlockChain,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockProcessor process.BlockProcessor,
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

// DoStartRoundJob method is the function which actually do the job of the StartRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct,
// from spos package)
func (sposWorker *SPOSConsensusWorker) DoStartRoundJob() bool {
	sposWorker.BlockBody = nil
	sposWorker.Header = nil
	sposWorker.Cns.Data = nil
	sposWorker.Cns.ResetRoundStatus()
	sposWorker.Cns.ResetRoundState()

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
		sposWorker.Cns.getFormattedTime(), getPrettyByteArray([]byte(leader)), msg))

	// TODO: Unccomment ShouldSync check
	//if sposWorker.ShouldSync() { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
	//	log.Info(fmt.Sprintf("%sCanceled round %d in subround %s, not synchronized",
	//		sposWorker.Cns.getFormattedTime(), sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBlock)))
	//	sposWorker.Cns.Chr.SetSelfSubround(-1)
	//	return false
	//}

	pubKeys := sposWorker.Cns.ConsensusGroup()

	selfIndex, err := sposWorker.Cns.IndexSelfConsensusGroup()

	if err != nil {
		log.Error(err.Error())
		sposWorker.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	err = sposWorker.multiSigner.Reset(pubKeys, uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())
		sposWorker.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	return true
}

// DoEndRoundJob method is the function which actually do the job of the EndRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct,
// from spos package)
func (sposWorker *SPOSConsensusWorker) DoEndRoundJob() bool {
	if !sposWorker.Cns.CheckEndRoundConsensus() {
		return false
	}

	bitmap := sposWorker.genBitmap(SrBitmap)

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
		sposWorker.BlockProcessor.RevertAccountState()
		return false
	}

	err = sposWorker.BlockProcessor.RemoveBlockTxsFromPool(sposWorker.BlockBody)

	if err != nil {
		log.Error(err.Error())
	}

	// broadcast block body
	err = sposWorker.broadcastTxBlockBody()
	if err != nil {
		log.Error(fmt.Sprintf("%s\n", err.Error()))
	}

	// broadcast header
	err = sposWorker.broadcastHeader()
	if err != nil {
		log.Error(fmt.Sprintf("%s\n", err.Error()))
	}

	if sposWorker.Cns.IsNodeLeaderInCurrentRound(sposWorker.Cns.SelfPubKey()) {
		log.Info(fmt.Sprintf("\n%s++++++++++++++++++++ ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN ++++++++++++++++++++\n\n",
			sposWorker.Cns.getFormattedTime(), sposWorker.Header.Nonce))
	} else {
		log.Info(fmt.Sprintf("\n%sxxxxxxxxxxxxxxxxxxxx ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN xxxxxxxxxxxxxxxxxxxx\n\n",
			sposWorker.Cns.getFormattedTime(), sposWorker.Header.Nonce))
	}

	sposWorker.Rounds++          // only for statistic
	sposWorker.RoundsWithBlock++ // only for statistic

	return true
}

// DoBlockJob method actually send the proposed block in the Block subround, when this node is leader
// (it is used as a handler function of the doSubroundJob pointer function declared in Subround struct,
// from spos package)
func (sposWorker *SPOSConsensusWorker) DoBlockJob() bool {
	isBlockJobDone, err := sposWorker.Cns.GetJobDone(sposWorker.Cns.SelfPubKey(), SrBlock)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrBlock) == SsFinished || // is subround Block already finished?
		isBlockJobDone || // has block already been sent?
		!sposWorker.Cns.IsNodeLeaderInCurrentRound(sposWorker.Cns.SelfPubKey()) { // is another node leader in this round?
		return false
	}

	if !sposWorker.SendBlockBody() ||
		!sposWorker.SendBlockHeader() {
		return false
	}

	err = sposWorker.Cns.SetJobDone(sposWorker.Cns.SelfPubKey(), SrBlock, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sposWorker.multiSigner.SetMessage(sposWorker.Cns.Data)

	return true
}

// SendBlockBody method send the proposed block body in the Block subround
func (sposWorker *SPOSConsensusWorker) SendBlockBody() bool {

	currentSubRound := sposWorker.GetSubround()

	haveTime := func() bool {
		if sposWorker.GetSubround() > currentSubRound {
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
		sposWorker.GetTime(),
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
	return sposWorker.Cns.Chr.GetSubroundFromDateTime(sposWorker.Cns.Chr.SyncTime().CurrentTime(sposWorker.Cns.Chr.ClockOffset()))
}

// SendBlockHeader method send the proposed block header in the Block subround
func (sposWorker *SPOSConsensusWorker) SendBlockHeader() bool {
	hdr := &block.Header{}

	if sposWorker.BlockChain.CurrentBlockHeader == nil {
		hdr.Nonce = 1
		hdr.Round = uint32(sposWorker.Cns.Chr.Round().Index())
		hdr.TimeStamp = sposWorker.GetTime()
	} else {
		hdr.Nonce = sposWorker.BlockChain.CurrentBlockHeader.Nonce + 1
		hdr.Round = uint32(sposWorker.Cns.Chr.Round().Index())
		hdr.TimeStamp = sposWorker.GetTime()

		prevHeader, err := sposWorker.marshalizer.Marshal(sposWorker.BlockChain.CurrentBlockHeader)

		if err != nil {
			log.Error(err.Error())
			return false
		}

		prevHeaderHash := sposWorker.hasher.Compute(string(prevHeader))
		hdr.PrevHash = prevHeaderHash
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
		sposWorker.GetTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Sending block header with nonce %d and hash %s\n",
		sposWorker.Cns.getFormattedTime(), hdr.Nonce, getPrettyByteArray(hdrHash)))

	sposWorker.Header = hdr
	sposWorker.Cns.Data = hdrHash

	return true
}

func (sposWorker *SPOSConsensusWorker) genCommitmentHash() ([]byte, error) {
	commitmentSecret, commitment, err := sposWorker.multiSigner.CreateCommitment()

	if err != nil {
		return nil, err
	}

	selfIndex, err := sposWorker.Cns.IndexSelfConsensusGroup()

	if err != nil {
		return nil, err
	}

	err = sposWorker.multiSigner.AddCommitment(uint16(selfIndex), commitment)

	if err != nil {
		return nil, err
	}

	err = sposWorker.multiSigner.SetCommitmentSecret(commitmentSecret)

	if err != nil {
		return nil, err
	}

	commitmentHash := sposWorker.hasher.Compute(string(commitment))

	err = sposWorker.multiSigner.AddCommitmentHash(uint16(selfIndex), commitmentHash)

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

	isCommHashJobDone, err := sposWorker.Cns.GetJobDone(sposWorker.Cns.SelfPubKey(), SrCommitmentHash)

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
		sposWorker.GetTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 2: Sending commitment hash\n", sposWorker.Cns.getFormattedTime()))

	err = sposWorker.Cns.SetJobDone(sposWorker.Cns.SelfPubKey(), SrCommitmentHash, true)

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
		isJobDone, err := sposWorker.Cns.GetJobDone(sposWorker.Cns.ConsensusGroup()[i], subround)

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

	isBitmapJobDone, err := sposWorker.Cns.GetJobDone(sposWorker.Cns.SelfPubKey(), SrBitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrBitmap) == SsFinished || // is subround Bitmap already finished?
		isBitmapJobDone || // has been bitmap already sent?
		!sposWorker.Cns.IsNodeLeaderInCurrentRound(sposWorker.Cns.SelfPubKey()) || // is another node leader in this round?
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
		sposWorker.GetTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 3: Sending bitmap\n", sposWorker.Cns.getFormattedTime()))

	for i := 0; i < len(sposWorker.Cns.ConsensusGroup()); i++ {
		isJobCommHashJobDone, err := sposWorker.Cns.GetJobDone(sposWorker.Cns.ConsensusGroup()[i], SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobCommHashJobDone {
			err = sposWorker.Cns.SetJobDone(sposWorker.Cns.ConsensusGroup()[i], SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

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

	isCommJobDone, err := sposWorker.Cns.GetJobDone(sposWorker.Cns.SelfPubKey(), SrCommitment)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrCommitment) == SsFinished || // is subround Commitment already finished?
		isCommJobDone || // has been commitment already sent?
		!sposWorker.Cns.IsValidatorInBitmap(sposWorker.Cns.SelfPubKey()) || // isn't node in the leader's bitmap?
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
		sposWorker.GetTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 4: Sending commitment\n", sposWorker.Cns.getFormattedTime()))

	err = sposWorker.Cns.SetJobDone(sposWorker.Cns.SelfPubKey(), SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
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

	isSignJobDone, err := sposWorker.Cns.GetJobDone(sposWorker.Cns.SelfPubKey(), SrSignature)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sposWorker.Cns.Status(SrSignature) == SsFinished || // is subround Signature already finished?
		isSignJobDone || // has been signature already sent?
		!sposWorker.Cns.IsValidatorInBitmap(sposWorker.Cns.SelfPubKey()) || // isn't node in the leader's bitmap?
		sposWorker.Cns.Data == nil { // is consensus data not set?
		return false
	}

	bitmap := sposWorker.genBitmap(SrBitmap)

	// first compute commitment aggregation
	_, err = sposWorker.multiSigner.AggregateCommitments(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sigPart, err := sposWorker.multiSigner.SignPartial(bitmap)

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
		sposWorker.GetTime(),
		sposWorker.Cns.Chr.Round().Index())

	if !sposWorker.SendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 5: Sending signature\n", sposWorker.Cns.getFormattedTime()))

	err = sposWorker.Cns.SetJobDone(sposWorker.Cns.SelfPubKey(), SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func (sposWorker *SPOSConsensusWorker) genConsensusDataSignature(cnsDta *ConsensusData) ([]byte, error) {

	cnsDtaStr, err := sposWorker.marshalizer.Marshal(cnsDta)

	if err != nil {
		return nil, err
	}

	signature, err := sposWorker.privKey.Sign(cnsDtaStr)

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

func (sposWorker *SPOSConsensusWorker) broadcastTxBlockBody() error {
	if sposWorker.BlockBody != nil {
		return ErrNilTxBlockBody
	}

	message, err := sposWorker.marshalizer.Marshal(sposWorker.BlockBody)

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

func (sposWorker *SPOSConsensusWorker) broadcastHeader() error {
	if sposWorker.Header == nil {
		return ErrNilBlockHeader
	}

	message, err := sposWorker.marshalizer.Marshal(sposWorker.Header)

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

// ExtendBlock method put this subround in the extended mode and print some messages
func (sposWorker *SPOSConsensusWorker) ExtendBlock() {
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
	log.Info(fmt.Sprintf("%sStep 5: Extended the (SIGNATURE) subround. Got only %d from %d sigantures which are not enough\n",
		sposWorker.Cns.getFormattedTime(), sposWorker.Cns.ComputeSize(SrSignature), len(sposWorker.Cns.ConsensusGroup())))
}

// ExtendEndRound method just print some messages as no extend will be permited, because a new round
// will be start
func (sposWorker *SPOSConsensusWorker) ExtendEndRound() {
	log.Info(fmt.Sprintf("\n%s++++++++++++++++++++ THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN ++++++++++++++++++++\n\n",
		sposWorker.Cns.getFormattedTime()))
	sposWorker.Rounds++ // only for statistic
}

// ReceivedMessage method redirects the received message to the channel which should handle it
func (sposWorker *SPOSConsensusWorker) ReceivedMessage(name string, data interface{}, msgInfo *p2p.MessageInfo) {
	if sposWorker.Cns.Chr.IsCancelled() {
		return
	}

	cnsData, ok := data.(*ConsensusData)

	if !ok {
		return
	}

	senderOK := sposWorker.Cns.IsNodeInConsensusGroup(string(cnsData.PubKey))

	if !senderOK {
		return
	}

	sigVerifErr := sposWorker.checkSignature(cnsData)
	if sigVerifErr != nil {
		return
	}

	if ch, ok := sposWorker.MessageChannels[cnsData.MsgType]; ok {
		ch <- cnsData
	}
}

// CheckChannels method is used to listen to the channels through which node receives and consumes,
// during the round, different messages from the nodes which are in the validators group
func (sposWorker *SPOSConsensusWorker) CheckChannels() {
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

	err = pubKey.Verify(dataNoSigString, signature)

	return err
}

// ReceivedBlockBody method is called when a block body is received through the block body channel.
func (sposWorker *SPOSConsensusWorker) ReceivedBlockBody(cnsDta *ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if node == sposWorker.Cns.SelfPubKey() || // is block body received from myself?
		!sposWorker.Cns.IsNodeLeaderInCurrentRound(node) || // is another node leader in this round?
		sposWorker.BlockBody != nil { // is block body already received?
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Received block body\n", sposWorker.Cns.getFormattedTime()))

	sposWorker.BlockBody = sposWorker.DecodeBlockBody(cnsDta.SubRoundData)

	if sposWorker.BlockBody != nil &&
		sposWorker.Header != nil {
		err := sposWorker.BlockProcessor.ProcessBlock(sposWorker.BlockChain, sposWorker.Header, sposWorker.BlockBody)

		if err != nil {
			log.Error(err.Error())
			return false
		}

		sposWorker.multiSigner.SetMessage(sposWorker.Cns.Data)
		err = sposWorker.Cns.RoundConsensus.SetJobDone(node, SrBlock, true)

		if err != nil {
			log.Error(err.Error())
			return false
		}
	}

	return true
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
		isBlockJobDone { // is block header already received?
		return false
	}

	hdr := sposWorker.DecodeBlockHeader(cnsDta.SubRoundData)

	if !sposWorker.CheckIfBlockIsValid(hdr) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s\n",
			sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBlock)))
		sposWorker.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Received block header with nonce %d and hash %s\n",
		sposWorker.Cns.getFormattedTime(), hdr.Nonce, getPrettyByteArray(cnsDta.BlockHeaderHash)))

	sposWorker.Header = hdr
	sposWorker.Cns.Data = cnsDta.BlockHeaderHash

	if sposWorker.BlockBody != nil &&
		sposWorker.Header != nil {
		err := sposWorker.BlockProcessor.ProcessBlock(sposWorker.BlockChain, sposWorker.Header, sposWorker.BlockBody)

		if err != nil {
			log.Error(err.Error())
			return false
		}

		sposWorker.multiSigner.SetMessage(sposWorker.Cns.Data)
		err = sposWorker.Cns.RoundConsensus.SetJobDone(node, SrBlock, true)

		if err != nil {
			log.Error(err.Error())
			return false
		}
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
		!sposWorker.Cns.IsNodeInConsensusGroup(node) || // isn't node in the jobDone group?
		isCommHashJobDone || // is commitment hash already received?
		sposWorker.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.BlockHeaderHash, sposWorker.Cns.Data) { // is this the consesnus data of this round?
		return false
	}

	// if this node is leader in this round and already he received 2/3 + 1 of commitment hashes
	// he will ignore any others received later
	if sposWorker.Cns.IsNodeLeaderInCurrentRound(sposWorker.Cns.SelfPubKey()) {
		if sposWorker.Cns.IsCommitmentHashReceived(sposWorker.Cns.Threshold(SrCommitmentHash)) {
			return false
		}
	}

	index, err := sposWorker.Cns.ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sposWorker.multiSigner.AddCommitmentHash(uint16(index), cnsDta.SubRoundData)

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
		isBitmapJobDone || // is bitmap already received?
		sposWorker.Cns.Data == nil || // is consensus data not set?

		!bytes.Equal(cnsDta.BlockHeaderHash, sposWorker.Cns.Data) { // is this the consesnus data of this round?
		return false
	}

	signersBitmap := cnsDta.SubRoundData

	// count signers
	nbSigners := countBitmapFlags(signersBitmap)

	if int(nbSigners) < sposWorker.Cns.Threshold(SrBitmap) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s\n",
			sposWorker.Cns.Chr.Round().Index(), sposWorker.Cns.GetSubroundName(SrBitmap)))
		sposWorker.Cns.Chr.SetSelfSubround(-1)
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
		isCommJobDone || // is commitment already received?
		sposWorker.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.BlockHeaderHash, sposWorker.Cns.Data) { // is this the consesnus data of this round?
		return false
	}

	index, err := sposWorker.Cns.ConsensusGroupIndex(node)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	computedCommitmentHash := sposWorker.hasher.Compute(string(cnsDta.SubRoundData))
	rcvCommitmentHash, err := sposWorker.multiSigner.CommitmentHash(uint16(index))

	if !bytes.Equal(computedCommitmentHash, rcvCommitmentHash) {
		log.Info(fmt.Sprintf("Commitment %s does not match, expected %s\n", computedCommitmentHash, rcvCommitmentHash))
		return false
	}

	err = sposWorker.multiSigner.AddCommitment(uint16(index), cnsDta.SubRoundData)

	if err != nil {
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
		isSignJobDone || // is signature already received?
		sposWorker.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.BlockHeaderHash, sposWorker.Cns.Data) { // is this the consesnus data of this round?
		return false
	}

	index, err := sposWorker.Cns.ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	bitmap := sposWorker.genBitmap(SrBitmap)

	// verify partial signature
	err = sposWorker.multiSigner.VerifyPartial(uint16(index), cnsDta.SubRoundData, bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sposWorker.multiSigner.AddSignPartial(uint16(index), cnsDta.SubRoundData)

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
	// TODO: This logic is temporary and it should be refactored after the bootstrap mechanism will be implemented

	if sposWorker.BlockChain.CurrentBlockHeader == nil {
		if receivedHeader.Nonce == 1 { // first block after genesis
			if bytes.Equal(receivedHeader.PrevHash, []byte("")) {
				return true
			}

			log.Info(fmt.Sprintf("Hash not match: local block hash is empty and node received block with previous hash %s\n",
				getPrettyByteArray(receivedHeader.PrevHash)))
			return false
		}

		// to resolve the situation when a node comes later in the network and it has the
		// bootstrap mechanism not implemented yet (he will accept the block received)
		log.Info(fmt.Sprintf("Nonce not match: local block nonce is 0 and node received block with nonce %d\n",
			receivedHeader.Nonce))
		log.Info(fmt.Sprintf("\n++++++++++++++++++++ ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET ++++++++++++++++++++\n\n",
			receivedHeader.Nonce))
		return true
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
			getPrettyByteArray(prevHeaderHash), getPrettyByteArray(receivedHeader.PrevHash)))
		return false
	}

	// to resolve the situation when a node misses some Blocks and it has the bootstrap mechanism
	// not implemented yet (he will accept the block received)
	log.Info(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d\n",
		sposWorker.BlockChain.CurrentBlockHeader.Nonce, receivedHeader.Nonce))
	log.Info(fmt.Sprintf("\n++++++++++++++++++++ ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET ++++++++++++++++++++\n\n",
		receivedHeader.Nonce))
	return true
}

// ShouldSync method returns the synch state of the node. If it returns 'true', this means that the node
// is not synchronized yet and it has to continue the bootstrapping mechanism, otherwise the node is already
// synched and it can participate to the consensus, if it is in the jobDone group of this round
func (sposWorker *SPOSConsensusWorker) ShouldSync() bool {
	if sposWorker.Cns == nil ||
		sposWorker.Cns.Chr == nil ||
		sposWorker.Cns.Chr.Round() == nil {
		return true
	}

	rnd := sposWorker.Cns.Chr.Round()

	if sposWorker.BlockChain == nil ||
		sposWorker.BlockChain.CurrentBlockHeader == nil {
		return rnd.Index() > 0
	}

	return sposWorker.BlockChain.CurrentBlockHeader.Round+1 < uint32(rnd.Index())
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
		return "Undifined message type"
	}
}

// GetTime method returns a string containing the current time
func (sposWorker *SPOSConsensusWorker) GetTime() uint64 {
	return uint64(sposWorker.Cns.Chr.SyncTime().CurrentTime(sposWorker.Cns.Chr.ClockOffset()).Unix())
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

	if cns.Status(SrCommitmentHash) == SsFinished {
		return true
	}

	threshold := cns.Threshold(SrCommitmentHash)

	if !cns.IsNodeLeaderInCurrentRound(cns.selfPubKey) {
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
		return "Undifined subround"
	}
}

// PrintBlockCM method prints the <BLOCK> consensus messages
func (cns *Consensus) PrintBlockCM() {
	if !cns.IsNodeLeaderInCurrentRound(cns.selfPubKey) {
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
	if !cns.IsNodeLeaderInCurrentRound(cns.selfPubKey) {
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

func getPrettyByteArray(array []byte) string {
	base64pk := make([]byte, base64.StdEncoding.EncodedLen(len(array)))
	base64.StdEncoding.Encode(base64pk, array)
	return string(base64pk)
}
