package node

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"

	"github.com/pkg/errors"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/davecgh/go-spew/spew"
	"time"
)

const CONSENSUS_TOPIC_NAME = "Consensus"

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func(*Node)

// Node is a structure that passes the configuration parameters and initializes
//  required services as requested
type Node struct {
	port                int
	marshalizer         marshal.Marshalizer
	ctx                 context.Context
	hasher              hashing.Hasher
	maxAllowedPeers     int
	pubSubStrategy      p2p.PubSubStrategy
	initialNodesPubkeys []string
	selfPubKey          string
	roundDuration       int64
	consensusGroupSize  int
	messenger           p2p.Messenger
	syncer              ntp.SyncTimer
	blockProcessor      process.BlockProcessor
	genesisTime         time.Time
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) *Node {
	node := &Node{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(node)
	}
	return node
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option) error {
	if n.IsRunning() {
		return errors.New("cannot apply options while node is running")
	}
	for _, opt := range opts {
		opt(n)
	}
	return nil
}

// IsRunning will return the current state of the node
func (n *Node) IsRunning() bool {
	return n.messenger != nil
}

// Address returns the first address of the running node
func (n *Node) Address() (string, error) {
	if !n.IsRunning() {
		return "", errors.New("node is not started yet")
	}
	return n.messenger.Addresses()[0], nil
}

// Start will create a new messenger and and set up the Node state as running
func (n *Node) Start() error {
	messenger, err := n.createNetMessenger()
	if err != nil {
		return err
	}
	n.messenger = messenger
	n.Bootstrap()
	return nil
}

// Stop closes the messenger and undos everything done in Start
func (n *Node) Stop() error {
	if !n.IsRunning() {
		return nil
	}
	err := n.messenger.Close()
	if err != nil {
		return err
	}

	n.messenger = nil
	return nil
}

// ConnectToInitialAddresses connect to the list of peers provided initialAddresses
func (n *Node) ConnectToInitialAddresses() error {
	if !n.IsRunning() {
		return errors.New("node is not started yet")
	}
	if n.initialNodesPubkeys == nil {
		return errors.New("no addresses to connect to")
	}

	// Don't try to connect to self
	tmp := n.removeSelfFromList(n.initialNodesPubkeys)
	n.messenger.ConnectToAddresses(n.ctx, tmp)
	return nil
}

// Bootstrap will try to connect to as many peers as possible
func (n *Node) Bootstrap() {
	n.messenger.Bootstrap(n.ctx)
}

// ConnectToAddresses will take a slice of addresses and try to connect to all of them.
func (n *Node) ConnectToAddresses(addresses []string) error {
	if !n.IsRunning() {
		return errors.New("node is not started yet")
	}
	n.messenger.ConnectToAddresses(n.ctx, addresses)
	return nil
}

// StartConsensus will start the consesus service for the current node
func (n *Node) StartConsensus() error {

	round := n.CreateRound()
	chr := n.CreateChronology(round)
	rndc := n.CreateRoundConsensus()
	rth := n.CreateRoundThreshold()
	rnds := n.CreateRoundStatus()
	cns := n.CreateConsensus(rndc, rth, rnds, chr)
	blkc := n.CreateBlockchain()
	sposWrk := n.CreateConsensusWorker(cns, blkc)
	topic := n.CreateConsensusTopic(sposWrk)

	n.messenger.AddTopic(topic)
	n.AddSubroundsToChronology(sposWrk)

	go sposWrk.Cns.Chr.StartRounds()
	go n.blockchainLog(sposWrk)

	return nil
}

func (n *Node) CreateRound() *chronology.Round {
	rnd := chronology.NewRound(
		n.genesisTime,
		n.syncer.CurrentTime(n.syncer.ClockOffset()),
		time.Millisecond*time.Duration(n.roundDuration))

	return rnd
}

func (n *Node) CreateChronology(round *chronology.Round) *chronology.Chronology {
	chr := chronology.NewChronology(
		true,
		true,
		round,
		n.genesisTime,
		n.syncer)

	return chr
}

func (n *Node) CreateRoundConsensus() *spos.RoundConsensus {
	nodes := n.initialNodesPubkeys[0:n.consensusGroupSize]
	rndc := spos.NewRoundConsensus(
		nodes,
		n.selfPubKey)

	rndc.ResetRoundState()

	return rndc
}

func (n *Node) CreateRoundThreshold() *spos.RoundThreshold {
	rth := spos.NewRoundThreshold()

	pbftThreshold := n.consensusGroupSize*2/3 + 1

	rth.SetThreshold(spos.SrBlock, 1)
	rth.SetThreshold(spos.SrCommitmentHash, pbftThreshold)
	rth.SetThreshold(spos.SrBitmap, pbftThreshold)
	rth.SetThreshold(spos.SrCommitment, pbftThreshold)
	rth.SetThreshold(spos.SrSignature, pbftThreshold)

	return rth
}

func (n *Node) CreateRoundStatus() *spos.RoundStatus {
	rnds := spos.NewRoundStatus()

	rnds.ResetRoundStatus()

	return rnds
}

func (n *Node) CreateConsensus(rndc *spos.RoundConsensus, rth *spos.RoundThreshold, rnds *spos.RoundStatus, chr *chronology.Chronology,
) *spos.Consensus {
	cns := spos.NewConsensus(
		true,
		nil,
		rndc,
		rth,
		rnds,
		chr)

	return cns
}

func (n *Node) CreateBlockchain() *blockchain.BlockChain {
	//config := blockchainConfig()
	//
	//blkc, err := blockchain.NewBlockChain(config)
	//
	//if err != nil {
	//	return nil
	//}

	blkc := &blockchain.BlockChain{}
	return blkc
}

//func blockchainConfig() *blockchain.Config {
//	cacher := storage.CacheConfig{Type: storage.LRUCache, Size: 100}
//	bloom := storage.BloomConfig{Size: 2048, HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv}}
//	persisterBlockStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "BlockStorage"}
//	persisterBlockHeaderStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "BlockHeaderStorage"}
//	persisterTxStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxStorage"}
//	return &blockchain.Config{
//		BlockStorage:       storage.UnitConfig{CacheConf: cacher, DBConf: persisterBlockStorage, BloomConf: bloom},
//		BlockHeaderStorage: storage.UnitConfig{CacheConf: cacher, DBConf: persisterBlockHeaderStorage, BloomConf: bloom},
//		TxStorage:          storage.UnitConfig{CacheConf: cacher, DBConf: persisterTxStorage, BloomConf: bloom},
//		TxPoolStorage:      cacher,
//		BlockCache:         cacher,
//	}
//}

func (n *Node) CreateConsensusWorker(cns *spos.Consensus, blkc *blockchain.BlockChain) *spos.SPOSConsensusWorker {
	sposWrk := spos.NewConsensusWorker(
		true,
		cns,
		blkc,
		n.hasher,
		n.marshalizer,
		n.blockProcessor)

	sposWrk.OnSendMessage = n.sendMessage

	return sposWrk
}

func (n *Node) CreateConsensusTopic(sposWrk *spos.SPOSConsensusWorker) *p2p.Topic {
	t := p2p.NewTopic(CONSENSUS_TOPIC_NAME, &spos.ConsensusData{}, n.marshalizer)
	t.AddDataReceived(sposWrk.ReceivedMessage)
	return t
}

func (n *Node) AddSubroundsToChronology(sposWrk *spos.SPOSConsensusWorker) {
	roundDuration := sposWrk.Cns.Chr.Round().TimeDuration()

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrStartRound),
		chronology.SubroundId(spos.SrBlock), int64(roundDuration*5/100),
		sposWrk.Cns.GetSubroundName(spos.SrStartRound),
		sposWrk.DoStartRoundJob,
		nil,
		sposWrk.Cns.CheckStartRoundConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrBlock),
		chronology.SubroundId(spos.SrCommitmentHash),
		int64(roundDuration*25/100),
		sposWrk.Cns.GetSubroundName(spos.SrBlock),
		sposWrk.DoBlockJob,
		sposWrk.ExtendBlock,
		sposWrk.Cns.CheckBlockConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrCommitmentHash),
		chronology.SubroundId(spos.SrBitmap),
		int64(roundDuration*40/100),
		sposWrk.Cns.GetSubroundName(spos.SrCommitmentHash),
		sposWrk.DoCommitmentHashJob,
		sposWrk.ExtendCommitmentHash,
		sposWrk.Cns.CheckCommitmentHashConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrBitmap),
		chronology.SubroundId(spos.SrCommitment),
		int64(roundDuration*55/100),
		sposWrk.Cns.GetSubroundName(spos.SrBitmap),
		sposWrk.DoBitmapJob,
		sposWrk.ExtendBitmap,
		sposWrk.Cns.CheckBitmapConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrCommitment),
		chronology.SubroundId(spos.SrSignature),
		int64(roundDuration*70/100),
		sposWrk.Cns.GetSubroundName(spos.SrCommitment),
		sposWrk.DoCommitmentJob,
		sposWrk.ExtendCommitment,
		sposWrk.Cns.CheckCommitmentConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrSignature),
		chronology.SubroundId(spos.SrEndRound),
		int64(roundDuration*85/100),
		sposWrk.Cns.GetSubroundName(spos.SrSignature),
		sposWrk.DoSignatureJob,
		sposWrk.ExtendSignature,
		sposWrk.Cns.CheckSignatureConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrEndRound),
		chronology.SubroundId(spos.SrEndRound+1),
		int64(roundDuration*95/100),
		sposWrk.Cns.GetSubroundName(spos.SrEndRound),
		sposWrk.DoEndRoundJob,
		sposWrk.ExtendEndRound,
		sposWrk.Cns.CheckEndRoundConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrEndRound+1),
		-1,
		int64(roundDuration*100/100),
		"<ADVANCE>",
		nil,
		nil,
		nil))
}

func (n *Node) blockchainLog(sposWrk *spos.SPOSConsensusWorker) {
	oldNounce := uint64(0)

	for {
		time.Sleep(100 * time.Millisecond)

		currentBlock := sposWrk.Blkc.CurrentBlockHeader
		if currentBlock == nil {
			continue
		}

		if currentBlock.Nonce > oldNounce {
			oldNounce = currentBlock.Nonce
			spew.Dump(currentBlock)
			fmt.Printf("\n********** There was %d rounds and was proposed %d blocks, which means %.2f%% hit rate **********\n",
				sposWrk.Rounds, sposWrk.RoundsWithBlock, float64(sposWrk.RoundsWithBlock)*100/float64(sposWrk.Rounds))
		}
	}
}

func (n *Node) sendMessage(cnsDta *spos.ConsensusData) {
	n.messenger.GetTopic(CONSENSUS_TOPIC_NAME).Broadcast(*cnsDta)
}

//Gets the balance for a specific address
func (n *Node) GetBalance(address string) (*big.Int, error) {
	return big.NewInt(0), nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(sender string, receiver string, amount big.Int, code string) (*transaction.Transaction, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

//GetTransaction gets the transaction
func (n *Node) GetTransaction(hash string) (*transaction.Transaction, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

func (n *Node) createNetMessenger() (p2p.Messenger, error) {
	if n.port == 0 {
		return nil, errors.New("Cannot start node on port 0")
	}
	if n.marshalizer == nil {
		return nil, errors.New("Canot start node without providing a marshalizer")
	}
	if n.hasher == nil {
		return nil, errors.New("Canot start node without providing a hasher")
	}
	if n.maxAllowedPeers == 0 {
		return nil, errors.New("Canot start node without providing maxAllowedPeers")
	}

	cp, err := p2p.NewConnectParamsFromPort(n.port)
	if err != nil {
		return nil, err
	}

	nm, err := p2p.NewNetMessenger(n.ctx, n.marshalizer, n.hasher, cp, n.maxAllowedPeers, n.pubSubStrategy)
	if err != nil {
		return nil, err
	}
	return nm, nil
}

func (n *Node) removeSelfFromList(peers []string) []string {
	tmp := peers[:0]
	addr, _ := n.Address()
	for _, p := range peers {
		if addr != p {
			tmp = append(tmp, p)
		}
	}
	return tmp
}

// WithPort sets up the port option for the Node
func WithPort(port int) Option {
	return func(n *Node) {
		n.port = port
	}
}

// WithMarshalizer sets up the marshalizer option for the Node
func WithMarshalizer(marshalizer marshal.Marshalizer) Option {
	return func(n *Node) {
		n.marshalizer = marshalizer
	}
}

// WithContext sets up the context option for the Node
func WithContext(ctx context.Context) Option {
	return func(n *Node) {
		n.ctx = ctx
	}
}

// WithHasher sets up the hasher option for the Node
func WithHasher(hasher hashing.Hasher) Option {
	return func(n *Node) {
		n.hasher = hasher
	}
}

// WithMaxAllowedPeers sets up the maxAllowedPeers option for the Node
func WithMaxAllowedPeers(maxAllowedPeers int) Option {
	return func(n *Node) {
		n.maxAllowedPeers = maxAllowedPeers
	}
}

// WithMaxAllowedPeers sets up the strategy option for the Node
func WithPubSubStrategy(strategy p2p.PubSubStrategy) Option {
	return func(n *Node) {
		n.pubSubStrategy = strategy
	}
}

func WithInitialNodesPubKeys(pubKeys []string) Option {
	return func(n *Node) {
		n.initialNodesPubkeys = pubKeys
	}
}

func WithSelfPubKey(pubKey string) Option {
	return func(n *Node) {
		n.selfPubKey = pubKey
	}
}

func WithRoundDuration(roundDuration int64) Option {
	return func(n *Node) {
		n.roundDuration = roundDuration
	}
}

func WithConsensusGroupSize(consensusGroupSize int) Option {
	return func(n *Node) {
		n.consensusGroupSize = consensusGroupSize
	}
}

func WithSyncer(syncer ntp.SyncTimer) Option {
	return func(n *Node) {
		n.syncer = syncer
	}
}

func WithBlockProcessor(blockProcessor process.BlockProcessor) Option {
	return func(n *Node) {
		n.blockProcessor = blockProcessor
	}
}

func WithGenesisTime(genesisTime time.Time) Option {
	return func(n *Node) {
		n.genesisTime = genesisTime
	}
}
