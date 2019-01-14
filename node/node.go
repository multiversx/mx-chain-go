package node

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/pkg/errors"
)

type topicName string

const (
	// TransactionTopic is the topic used for sharing transactions
	TransactionTopic topicName = "tx"
	// ConsensusTopic is the topic used in consensus algorithm
	ConsensusTopic topicName = "cns"
	// HeadersTopic is the topic used for sharing block headers
	HeadersTopic topicName = "hdr"
	// TxBlockBodyTopic is the topic used for sharing transactions block bodies
	TxBlockBodyTopic topicName = "txBlk"
	// PeerChBodyTopic is used for sharing peer change block bodies
	PeerChBodyTopic topicName = "peerCh"
	// StateBodyTopic is used for sharing state block bodies
	StateBodyTopic topicName = "state"
)

var log = logger.NewDefaultLogger()

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func(*Node) error

// Node is a structure that passes the configuration parameters and initializes
//  required services as requested
type Node struct {
	port                     int
	marshalizer              marshal.Marshalizer
	ctx                      context.Context
	hasher                   hashing.Hasher
	maxAllowedPeers          int
	pubSubStrategy           p2p.PubSubStrategy
	initialNodesPubkeys      []string
	initialNodesBalances     map[string]big.Int
	roundDuration            uint64
	consensusGroupSize       int
	messenger                p2p.Messenger
	syncer                   ntp.SyncTimer
	blockProcessor           process.BlockProcessor
	genesisTime              time.Time
	elasticSubrounds         bool
	accounts                 state.AccountsAdapter
	addrConverter            state.AddressConverter
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter

	privateKey       crypto.PrivateKey
	publicKey        crypto.PublicKey
	singleSignKeyGen crypto.KeyGenerator
	multisig         crypto.MultiSigner

	blkc             *blockchain.BlockChain
	dataPool         data.TransientDataHolder
	shardCoordinator sharding.ShardCoordinator

	interceptors []process.Interceptor
	resolvers    []process.Resolver
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
		return errNodeNotStarted
	}
	n.messenger.ConnectToAddresses(n.ctx, addresses)
	return nil
}

// BindInterceptorsResolvers will start the interceptors and resolvers
func (n *Node) BindInterceptorsResolvers() error {
	if !n.IsRunning() {
		return errNodeNotStarted
	}

	err := n.createInterceptors()
	if err != nil {
		return err
	}

	err = n.createResolvers()
	if err != nil {
		return err
	}

	return nil
}

// StartConsensus will start the consesus service for the current node
func (n *Node) StartConsensus() error {

	genessisBlock, err := n.createGenesisBlock()
	if err != nil {
		return err
	}
	n.blkc.GenesisBlock = genessisBlock
	round := n.createRound()
	chr := n.createChronology(round)
	rndc := n.createRoundConsensus()
	rth := n.createRoundThreshold()
	rnds := n.createRoundStatus()
	cns := n.createConsensus(rndc, rth, rnds, chr)
	sposWrk, err := n.createConsensusWorker(cns)

	if err != nil {
		return err
	}

	topic := n.createConsensusTopic(sposWrk)

	err = n.messenger.AddTopic(topic)

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
		!n.elasticSubrounds,
		round,
		n.genesisTime,
		n.syncer)

	return chr
}

// createRoundConsensus method creates a RoundConsensus object
func (n *Node) createRoundConsensus() *spos.RoundConsensus {

	selfId, err := n.publicKey.ToByteArray()

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	nodes := n.initialNodesPubkeys[0:n.consensusGroupSize]
	rndc := spos.NewRoundConsensus(
		nodes,
		string(selfId))

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
		nil,
		rndc,
		rth,
		rnds,
		chr)

	return cns
}

// createConsensusWorker method creates a ConsensusWorker object
func (n *Node) createConsensusWorker(cns *spos.Consensus) (*spos.SPOSConsensusWorker, error) {
	sposWrk, err := spos.NewConsensusWorker(
		cns,
		n.blkc,
		n.hasher,
		n.marshalizer,
		n.blockProcessor,
		n.multisig,
		n.singleSignKeyGen,
		n.privateKey,
		n.publicKey,
	)

	if err != nil {
		return nil, err
	}

	sposWrk.SendMessage = n.sendMessage

	return sposWrk, nil
}

// createConsensusTopic creates a consensus topic for node
func (n *Node) createConsensusTopic(sposWrk *spos.SPOSConsensusWorker) *p2p.Topic {
	t := p2p.NewTopic(string(ConsensusTopic), &spos.ConsensusData{}, n.marshalizer)
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

	topic := n.messenger.GetTopic(string(TransactionTopic))

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

func (n *Node) createGenesisBlock() (*block.Header, error) {
	blockBody := n.blockProcessor.CreateGenesisBlockBody(n.initialNodesBalances, 0)
	marshalizedBody, err := n.marshalizer.Marshal(blockBody)
	if err != nil {
		return nil, err
	}
	blockBodyHash := n.hasher.Compute(string(marshalizedBody))
	return &block.Header{
		Nonce:         0,
		ShardId:       blockBody.ShardID,
		TimeStamp:     uint64(n.genesisTime.Unix()),
		BlockBodyHash: blockBodyHash,
		BlockBodyType: block.StateBlock,
		Signature:     blockBodyHash,
	}, nil
}

func (n *Node) blockchainLog(sposWrk *spos.SPOSConsensusWorker) {
	// TODO: this method and its call should be removed after initial testing of our first version of testnet
	oldNonce := uint64(0)
	prevHeaderHash := []byte("")
	recheckPeriod := sposWrk.Cns.Chr.Round().TimeDuration() * 5 / 100

	for {
		time.Sleep(recheckPeriod)

		hdr := sposWrk.BlockChain.CurrentBlockHeader
		txBlock := sposWrk.BlockChain.CurrentTxBlockBody

		if hdr == nil || txBlock == nil {
			continue
		}

		if hdr.Nonce > oldNonce {
			newNonce, newPrevHash, blockHash, err := n.computeNewNoncePrevHash(sposWrk, hdr, txBlock, prevHeaderHash)

			if err != nil {
				log.Error(err.Error())
				continue
			}

			n.displayLogInfo(hdr, txBlock, newPrevHash, prevHeaderHash, sposWrk, blockHash)

			oldNonce = newNonce
			prevHeaderHash = newPrevHash
		}
	}
}

func (n *Node) computeNewNoncePrevHash(
	sposWrk *spos.SPOSConsensusWorker,
	hdr *block.Header,
	txBlock *block.TxBlockBody,
	prevHash []byte,
) (uint64, []byte, []byte, error) {

	if sposWrk == nil {
		return 0, nil, nil, errNilSposWorker
	}

	if sposWrk.BlockChain == nil {
		return 0, nil, nil, errNilBlockchain
	}

	headerMarsh, err := n.marshalizer.Marshal(hdr)
	if err != nil {
		return 0, nil, nil, err
	}

	txBlkMarsh, err := n.marshalizer.Marshal(txBlock)
	if err != nil {
		return 0, nil, nil, err
	}

	headerHash := n.hasher.Compute(string(headerMarsh))
	blockHash := n.hasher.Compute(string(txBlkMarsh))

	return hdr.Nonce, headerHash, blockHash, nil
}

func (n *Node) displayLogInfo(
	header *block.Header,
	txBlock *block.TxBlockBody,
	headerHash []byte,
	prevHash []byte,
	sposWrk *spos.SPOSConsensusWorker,
	blockHash []byte,
) {

	log.Info(fmt.Sprintf("Block with nonce %d and hash %s was added into the blockchain. Previous block hash was %s\n\n", header.Nonce, toB64(headerHash), toB64(prevHash)))

	dispHeader, dispLines := createDisplayableHeaderAndBlockBody(header, txBlock, blockHash)

	tblString, err := display.CreateTableString(dispHeader, dispLines)
	if err != nil {
		log.Error(err.Error())
	}
	fmt.Println(tblString)

	log.Info(fmt.Sprintf("\n********** There was %d rounds and was proposed %d blocks, which means %.2f%% hit rate **********\n",
		sposWrk.Rounds, sposWrk.RoundsWithBlock, float64(sposWrk.RoundsWithBlock)*100/float64(sposWrk.Rounds)))
}

func createDisplayableHeaderAndBlockBody(
	hdr *block.Header,
	txBody *block.TxBlockBody,
	txBlockHash []byte) ([]string, []*display.LineData) {

	header := []string{"Part", "Parameter", "Value"}

	lines := displayHeader(hdr)

	if hdr.BlockBodyType == block.TxBlock {
		lines = displayTxBlockBody(lines, txBody, txBlockHash)

		return header, lines
	}

	//TODO: implement the other block bodies

	lines = append(lines, display.NewLineData(false, []string{"Unknown", "", ""}))
	return header, lines
}

func displayHeader(hdr *block.Header) []*display.LineData {
	lines := make([]*display.LineData, 0)

	lines = append(lines, display.NewLineData(false, []string{
		"Header",
		"Nonce",
		fmt.Sprintf("%d", hdr.Nonce)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Shard",
		fmt.Sprintf("%d", hdr.ShardId)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Epoch",
		fmt.Sprintf("%d", hdr.Epoch)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Round",
		fmt.Sprintf("%d", hdr.Round)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Timestamp",
		fmt.Sprintf("%d", hdr.TimeStamp)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev hash",
		toB64(hdr.PrevHash)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Body type",
		hdr.BlockBodyType.String()}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Body hash",
		toB64(hdr.BlockBodyHash)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Pub keys bitmap",
		toHex(hdr.PubKeysBitmap)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Commitment",
		toB64(hdr.Commitment)}))
	lines = append(lines, display.NewLineData(true, []string{
		"",
		"Signature",
		toB64(hdr.Signature)}))

	return lines
}

func displayTxBlockBody(lines []*display.LineData, txBody *block.TxBlockBody, hash []byte) []*display.LineData {
	lines = append(lines, display.NewLineData(false, []string{"TxBody", "Block hash", toB64(hash)}))
	lines = append(lines, display.NewLineData(true, []string{"", "Root hash", toB64(txBody.RootHash)}))

	for i := 0; i < len(txBody.MiniBlocks); i++ {
		miniBlock := txBody.MiniBlocks[i]

		part := fmt.Sprintf("TxBody_%d", miniBlock.ShardID)

		if miniBlock.TxHashes == nil || len(miniBlock.TxHashes) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				part, "", "<NIL> or <EMPTY>"}))
		}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			lines = append(lines, display.NewLineData(false, []string{
				part,
				fmt.Sprintf("Tx hash %d", j),
				toB64(miniBlock.TxHashes[j])}))

			part = ""
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	return lines
}

func toHex(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return "0x" + hex.EncodeToString(buff)
}

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}

func (n *Node) sendMessage(cnsDta *spos.ConsensusData) {
	topic := n.messenger.GetTopic(string(ConsensusTopic))

	if topic == nil {
		log.Debug(fmt.Sprintf("could not get consensus topic"))
		return
	}

	err := topic.Broadcast(*cnsDta)

	if err != nil {
		log.Debug(fmt.Sprintf("could not broadcast message: " + err.Error()))
	}
}
