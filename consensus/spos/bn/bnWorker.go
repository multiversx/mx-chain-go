package bn

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

var log = logger.NewDefaultLogger()

// Worker defines the data needed by spos to communicate between nodes which are in the validators group
type Worker struct {
	SPoS                   *spos.Spos
	Header                 *block.Header
	BlockBody              *block.TxBlockBody
	BlockChain             *blockchain.BlockChain
	BlockProcessor         process.BlockProcessor
	boot                   process.Bootstraper
	MessageChannels        map[MessageType]chan *spos.ConsensusData
	ReceivedMessageChannel chan *spos.ConsensusData
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	keyGen                 crypto.KeyGenerator
	privKey                crypto.PrivateKey
	pubKey                 crypto.PublicKey
	multiSigner            crypto.MultiSigner
	vgs                    consensus.ValidatorGroupSelector
	SendMessage            func(consensus *spos.ConsensusData)
	BroadcastHeader        func([]byte)
	BroadcastBlockBody     func([]byte)
	ReceivedMessages       map[MessageType][]*spos.ConsensusData
	mutReceivedMessages    sync.RWMutex
	mutCheckConsensus      sync.Mutex
}

// NewWorker creates a new Worker object
func NewWorker(
	sPoS *spos.Spos,
	blkc *blockchain.BlockChain,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockProcessor process.BlockProcessor,
	boot process.Bootstraper,
	multisig crypto.MultiSigner,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
) (*Worker, error) {

	err := checkNewWorkerParams(
		sPoS,
		blkc,
		hasher,
		marshalizer,
		blockProcessor,
		boot,
		multisig,
		keyGen,
		privKey,
		pubKey,
	)

	if err != nil {
		return nil, err
	}

	pbftThreshold := sPoS.ConsensusGroupSize()*2/3 + 1

	sPoS.SetThreshold(SrBlock, 1)
	sPoS.SetThreshold(SrCommitmentHash, pbftThreshold)
	sPoS.SetThreshold(SrBitmap, pbftThreshold)
	sPoS.SetThreshold(SrCommitment, pbftThreshold)
	sPoS.SetThreshold(SrSignature, pbftThreshold)

	wrk := Worker{
		SPoS:           sPoS,
		BlockChain:     blkc,
		hasher:         hasher,
		marshalizer:    marshalizer,
		BlockProcessor: blockProcessor,
		boot:           boot,
		multiSigner:    multisig,
		keyGen:         keyGen,
		privKey:        privKey,
		pubKey:         pubKey,
	}

	bnf, err := NewbnFactory(&wrk)

	if err != nil {
		return nil, err
	}

	bnf.GenerateSubrounds()

	wrk.MessageChannels = make(map[MessageType]chan *spos.ConsensusData)

	wrk.MessageChannels[MtBlockBody] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtBlockHeader] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtCommitmentHash] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtBitmap] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtCommitment] = make(chan *spos.ConsensusData)
	wrk.MessageChannels[MtSignature] = make(chan *spos.ConsensusData)

	wrk.ReceivedMessageChannel = make(chan *spos.ConsensusData, sPoS.ConsensusGroupSize()*consensusSubrounds)

	wrk.initReceivedMessages()

	go wrk.checkReceivedMessageChannel()
	go wrk.checkChannels()

	return &wrk, nil
}

func checkNewWorkerParams(
	sPoS *spos.Spos,
	blkc *blockchain.BlockChain,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blockProcessor process.BlockProcessor,
	boot process.Bootstraper,
	multisig crypto.MultiSigner,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
) error {
	if sPoS == nil {
		return spos.ErrNilConsensus
	}

	if blkc == nil {
		return spos.ErrNilBlockChain
	}

	if hasher == nil {
		return spos.ErrNilHasher
	}

	if marshalizer == nil {
		return spos.ErrNilMarshalizer
	}

	if blockProcessor == nil {
		return spos.ErrNilBlockProcessor
	}

	if boot == nil {
		return spos.ErrNilBlootstrap
	}

	if multisig == nil {
		return spos.ErrNilMultiSigner
	}

	if keyGen == nil {
		return spos.ErrNilKeyGenerator
	}

	if privKey == nil {
		return spos.ErrNilPrivateKey
	}

	if pubKey == nil {
		return spos.ErrNilPublicKey
	}

	return nil
}
