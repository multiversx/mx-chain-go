package node

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

type topicName string

const (
	transactionTopic topicName = "tx"
	consensusTopic   topicName = "cns"
)

var log = logger.NewDefaultLogger()

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func(*Node) error

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
	publicKey           crypto.PublicKey
	roundDuration       int64
	consensusGroupSize  int
	messenger           p2p.Messenger
	syncer              ntp.SyncTimer
	blockProcessor      process.BlockProcessor
	genesisTime         time.Time
	elasticSubrounds    bool
	accounts            state.AccountsAdapter
	addrConverter       state.AddressConverter
	privateKey          crypto.PrivateKey
	blkc                *blockchain.BlockChain
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) (*Node, error) {
	node := &Node{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		err := opt(node)
		if err != nil {
			return nil, errors.New("error applying option: " + err.Error())
		}
	}
	return node, nil
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option) error {
	if n.IsRunning() {
		return errors.New("cannot apply options while node is running")
	}
	for _, opt := range opts {
		err := opt(n)
		if err != nil {
			return errors.New("error applying option: " + err.Error())
		}
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
	n.P2PBootstrap()
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

// P2PBootstrap will try to connect to many peers as possible
func (n *Node) P2PBootstrap() {
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

	round := n.createRound()
	chr := n.createChronology(round)
	rndc := n.createRoundConsensus()
	rth := n.createRoundThreshold()
	rnds := n.createRoundStatus()
	cns := n.createConsensus(rndc, rth, rnds, chr)
	sposWrk := n.createConsensusWorker(cns)
	topic := n.createConsensusTopic(sposWrk)

	err := n.messenger.AddTopic(topic)

	if err != nil {
		log.Debug(fmt.Sprintf(err.Error()))
	}

	n.addSubroundsToChronology(sposWrk)

	go sposWrk.Cns.Chr.StartRounds()
	go n.blockchainLog(sposWrk)

	return nil
}

// createRound method creates a round object
func (n *Node) createRound() *chronology.Round {
	rnd := chronology.NewRound(
		n.genesisTime,
		n.syncer.CurrentTime(n.syncer.ClockOffset()),
		time.Millisecond*time.Duration(n.roundDuration))

	return rnd
}

// createChronology method creates a chronology object
func (n *Node) createChronology(round *chronology.Round) *chronology.Chronology {
	chr := chronology.NewChronology(
		true,
		!n.elasticSubrounds,
		round,
		n.genesisTime,
		n.syncer)

	return chr
}

func (n *Node) getPrettyPublicKey() string {
	pk, _ := n.publicKey.ToByteArray()
	base64pk := make([]byte, base64.StdEncoding.EncodedLen(len(pk)))
	base64.StdEncoding.Encode(base64pk, pk)
	return string(base64pk)
}

// createRoundConsensus method creates a RoundConsensus object
func (n *Node) createRoundConsensus() *spos.RoundConsensus {

	nodes := n.initialNodesPubkeys[0:n.consensusGroupSize]
	rndc := spos.NewRoundConsensus(
		nodes,
		n.getPrettyPublicKey())

	rndc.ResetRoundState()

	return rndc
}

// createRoundThreshold method creates a RoundThreshold object
func (n *Node) createRoundThreshold() *spos.RoundThreshold {
	rth := spos.NewRoundThreshold()

	pbftThreshold := n.consensusGroupSize*2/3 + 1

	rth.SetThreshold(spos.SrBlock, 1)
	rth.SetThreshold(spos.SrCommitmentHash, pbftThreshold)
	rth.SetThreshold(spos.SrBitmap, pbftThreshold)
	rth.SetThreshold(spos.SrCommitment, pbftThreshold)
	rth.SetThreshold(spos.SrSignature, pbftThreshold)

	return rth
}

// createRoundStatus method creates a RoundStatus object
func (n *Node) createRoundStatus() *spos.RoundStatus {
	rnds := spos.NewRoundStatus()

	rnds.ResetRoundStatus()

	return rnds
}

// createConsensus method creates a Consensus object
func (n *Node) createConsensus(rndc *spos.RoundConsensus, rth *spos.RoundThreshold, rnds *spos.RoundStatus, chr *chronology.Chronology,
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

// createConsensusWorker method creates a ConsensusWorker object
func (n *Node) createConsensusWorker(cns *spos.Consensus) *spos.SPOSConsensusWorker {
	sposWrk := spos.NewConsensusWorker(
		true,
		cns,
		n.blkc,
		n.hasher,
		n.marshalizer,
		n.blockProcessor)

	sposWrk.SendMessage = n.sendMessage

	return sposWrk
}

// createConsensusTopic creates a consensus topic for node
func (n *Node) createConsensusTopic(sposWrk *spos.SPOSConsensusWorker) *p2p.Topic {
	t := p2p.NewTopic(string(consensusTopic), &spos.ConsensusData{}, n.marshalizer)
	t.AddDataReceived(sposWrk.ReceivedMessage)
	return t
}

// addSubroundsToChronology adds subrounds to chronology
func (n *Node) addSubroundsToChronology(sposWrk *spos.SPOSConsensusWorker) {
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
		chronology.SubroundId(spos.SrAdvance),
		int64(roundDuration*95/100),
		sposWrk.Cns.GetSubroundName(spos.SrEndRound),
		sposWrk.DoEndRoundJob,
		sposWrk.ExtendEndRound,
		sposWrk.Cns.CheckEndRoundConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrAdvance),
		-1,
		int64(roundDuration*100/100),
		sposWrk.Cns.GetSubroundName(spos.SrAdvance),
		nil,
		nil,
		nil))
}

// GetBalance gets the balance for a specific address
func (n *Node) GetBalance(address string) (*big.Int, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}
	accAddress, err := n.addrConverter.CreateAddressFromHex(address)
	if err != nil {
		return nil, errors.New("invalid address: " + err.Error())
	}
	account, err := n.accounts.GetExistingAccount(accAddress)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}

	if account == nil {
		return big.NewInt(0), nil
	}

	return &account.BaseAccount().Balance, nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(sender string, receiver string, amount big.Int, code string) (*transaction.Transaction, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	if n.privateKey == nil {
		return nil, errors.New("initialize PrivateKey first")
	}

	senderAddress, err := n.addrConverter.CreateAddressFromHex(sender)
	if err != nil {
		return nil, errors.New("could not create sender address from provided param")
	}
	senderAccount, err := n.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}
	newNonce := uint64(0)
	if senderAccount != nil {
		newNonce = senderAccount.BaseAccount().Nonce
	}

	tx := transaction.Transaction{
		Nonce:   newNonce,
		Value:   amount,
		RcvAddr: []byte(receiver),
		SndAddr: []byte(sender),
	}

	txToByteArray, err := n.marshalizer.Marshal(tx)
	if err != nil {
		return nil, errors.New("could not create byte array representation of the transaction")
	}

	sig, err := n.privateKey.Sign(txToByteArray)
	if err != nil {
		return nil, errors.New("could not sign the transaction")
	}
	tx.Signature = sig

	return &tx, nil
}

// SendTransaction will send a new transaction on the topic channel
func (n *Node) SendTransaction(
	nonce uint64,
	sender string,
	receiver string,
	value big.Int,
	transactionData string,
	signature string) (*transaction.Transaction, error) {

	tx := transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		RcvAddr:   []byte(receiver),
		SndAddr:   []byte(sender),
		Data:      []byte(transactionData),
		Signature: []byte(signature),
	}

	topic := n.messenger.GetTopic(string(transactionTopic))

	if topic == nil {
		return nil, errors.New("could not get transaction topic")
	}

	err := topic.Broadcast(tx)
	if err != nil {
		return nil, errors.New("could not broadcast transaction: " + err.Error())
	}
	return &tx, nil
}

//GetTransaction gets the transaction
func (n *Node) GetTransaction(hash string) (*transaction.Transaction, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func (n *Node) createNetMessenger() (p2p.Messenger, error) {
	if n.port == 0 {
		return nil, errors.New("Cannot start node on port 0")
	}

	if n.maxAllowedPeers == 0 {
		return nil, errors.New("Cannot start node without providing maxAllowedPeers")
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
	return func(n *Node) error {
		n.port = port
		return nil
	}
}

// WithMarshalizer sets up the marshalizer option for the Node
func WithMarshalizer(marshalizer marshal.Marshalizer) Option {
	return func(n *Node) error {
		if marshalizer == nil {
			return errors.New("trying to set nil marshalizer")
		}
		n.marshalizer = marshalizer
		return nil
	}
}

// WithContext sets up the context option for the Node
func WithContext(ctx context.Context) Option {
	return func(n *Node) error {
		if ctx == nil {
			return errors.New("trying to set nil context")
		}
		n.ctx = ctx
		return nil
	}
}

// WithHasher sets up the hasher option for the Node
func WithHasher(hasher hashing.Hasher) Option {
	return func(n *Node) error {
		if hasher == nil {
			return errors.New("trying to set nil hasher")
		}
		n.hasher = hasher
		return nil
	}
}

// WithMaxAllowedPeers sets up the maxAllowedPeers option for the Node
func WithMaxAllowedPeers(maxAllowedPeers int) Option {
	return func(n *Node) error {
		n.maxAllowedPeers = maxAllowedPeers
		return nil
	}
}

// WithPubSubStrategy sets up the strategy option for the Node
func WithPubSubStrategy(strategy p2p.PubSubStrategy) Option {
	return func(n *Node) error {
		n.pubSubStrategy = strategy
		return nil
	}
}

// WithAccountsAdapter sets up the accounts adapter option for the Node
func WithAccountsAdapter(accounts state.AccountsAdapter) Option {
	return func(n *Node) error {
		if accounts == nil {
			return errors.New("trying to set nil accounts adapter")
		}
		n.accounts = accounts
		return nil
	}
}

// WithAddressConverter sets up the address converter adapter option for the Node
func WithAddressConverter(addrConverter state.AddressConverter) Option {
	return func(n *Node) error {
		if addrConverter == nil {
			return errors.New("trying to set nil address converter")
		}
		n.addrConverter = addrConverter
		return nil
	}
}

// WithBlockChain sets up the blockchain option for the Node
func WithBlockChain(blkc *blockchain.BlockChain) Option {
	return func(n *Node) error {
		if blkc == nil {
			return errors.New("trying to set nil blockchain")
		}
		n.blkc = blkc
		return nil
	}
}

// WithPrivateKey sets up the private key option for the Node
func WithPrivateKey(sk crypto.PrivateKey) Option {
	return func(n *Node) error {
		if sk == nil {
			return errors.New("trying to set nil private key")
		}
		n.privateKey = sk
		return nil
	}
}

// WithInitialNodesPubKeys sets up the initial nodes public key option for the Node
func WithInitialNodesPubKeys(pubKeys []string) Option {
	return func(n *Node) error {
		n.initialNodesPubkeys = pubKeys
		return nil
	}
}

// WithPublicKey sets up the public key option for the Node
func WithPublicKey(pk crypto.PublicKey) Option {
	return func(n *Node) error {
		n.publicKey = pk
		return nil
	}
}

// WithRoundDuration sets up the round duration option for the Node
func WithRoundDuration(roundDuration int64) Option {
	return func(n *Node) error {
		n.roundDuration = roundDuration
		return nil
	}
}

// WithConsensusGroupSize sets up the consensus group size option for the Node
func WithConsensusGroupSize(consensusGroupSize int) Option {
	return func(n *Node) error {
		n.consensusGroupSize = consensusGroupSize
		return nil
	}
}

// WithSyncer sets up the syncer option for the Node
func WithSyncer(syncer ntp.SyncTimer) Option {
	return func(n *Node) error {
		if syncer == nil {
			return errors.New("trying to set nil sync timer")
		}

		n.syncer = syncer
		return nil
	}
}

// WithBlockProcessor sets up the block processor option for the Node
func WithBlockProcessor(blockProcessor process.BlockProcessor) Option {
	return func(n *Node) error {
		if blockProcessor == nil {
			return errors.New("trying to set nil block processor")
		}
		n.blockProcessor = blockProcessor
		return nil
	}
}

// WithGenesisTime sets up the genesis time option for the Node
func WithGenesisTime(genesisTime time.Time) Option {
	return func(n *Node) error {
		n.genesisTime = genesisTime
		return nil
	}
}

// WithElasticSubrounds sets up the elastic subround option for the Node
func WithElasticSubrounds(elasticSubrounds bool) Option {
	return func(n *Node) error {
		n.elasticSubrounds = elasticSubrounds
		return nil
	}
}

func (n *Node) blockchainLog(sposWrk *spos.SPOSConsensusWorker) {
	// TODO: this method and its call should be removed after initial testing of aur first version of testnet
	oldNonce := uint64(0)

	for {
		time.Sleep(100 * time.Millisecond)

		currentBlock := sposWrk.Blkc.CurrentBlockHeader
		if currentBlock == nil {
			continue
		}

		if currentBlock.Nonce > oldNonce {
			oldNonce = currentBlock.Nonce
			spew.Dump(currentBlock)
			fmt.Printf("\n********** There was %d rounds and was proposed %d blocks, which means %.2f%% hit rate **********\n",
				sposWrk.Rounds, sposWrk.RoundsWithBlock, float64(sposWrk.RoundsWithBlock)*100/float64(sposWrk.Rounds))
		}
	}
}

func (n *Node) sendMessage(cnsDta *spos.ConsensusData) {
	topic := n.messenger.GetTopic(string(consensusTopic))

	if topic == nil {
		log.Debug(fmt.Sprintf("could not get consensus topic"))
		return
	}

	err := topic.Broadcast(*cnsDta)

	if err != nil {
		log.Debug(fmt.Sprintf("could not broadcast message: " + err.Error()))
	}
}
