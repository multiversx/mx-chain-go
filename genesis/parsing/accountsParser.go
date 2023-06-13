package parsing

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	scrData "github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("genesis/parsing")

// accountsParser hold data for initial accounts decoded data from json file
type accountsParser struct {
	initialAccounts    []*data.InitialAccount
	entireSupply       *big.Int
	minterAddressBytes []byte
	pubkeyConverter    core.PubkeyConverter
	keyGenerator       crypto.KeyGenerator
	hasher             hashing.Hasher
	marshalizer        marshal.Marshalizer
}

// NewAccountsParser creates a new decoded accounts genesis structure from json config file
func NewAccountsParser(args genesis.AccountsParserArgs) (*accountsParser, error) {
	if args.EntireSupply == nil {
		return nil, genesis.ErrNilEntireSupply
	}
	if big.NewInt(0).Cmp(args.EntireSupply) >= 0 {
		return nil, genesis.ErrInvalidEntireSupply
	}
	if check.IfNil(args.PubkeyConverter) {
		return nil, genesis.ErrNilPubkeyConverter
	}
	if check.IfNil(args.KeyGenerator) {
		return nil, genesis.ErrNilKeyGenerator
	}
	if check.IfNil(args.Hasher) {
		return nil, genesis.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, genesis.ErrNilMarshalizer
	}

	minterAddressBytes, err := args.PubkeyConverter.Decode(args.MinterAddress)
	if err != nil {
		return nil, fmt.Errorf("%w while decoding minter address. error: %s", genesis.ErrInvalidAddress, err.Error())
	}

	initialAccounts := make([]*data.InitialAccount, 0)
	err = core.LoadJsonFile(&initialAccounts, args.GenesisFilePath)
	if err != nil {
		return nil, err
	}

	gp := &accountsParser{
		initialAccounts:    initialAccounts,
		entireSupply:       args.EntireSupply,
		minterAddressBytes: minterAddressBytes,
		pubkeyConverter:    args.PubkeyConverter,
		keyGenerator:       args.KeyGenerator,
		hasher:             args.Hasher,
		marshalizer:        args.Marshalizer,
	}

	err = gp.process()
	if err != nil {
		return nil, err
	}

	return gp, nil
}

func (ap *accountsParser) process() error {
	totalSupply := big.NewInt(0)
	for _, initialAccount := range ap.initialAccounts {
		err := ap.parseElement(initialAccount)
		if err != nil {
			return err
		}

		err = ap.checkInitialAccount(initialAccount)
		if err != nil {
			return err
		}

		totalSupply.Add(totalSupply, initialAccount.Supply)
	}

	err := ap.checkForDuplicates()
	if err != nil {
		return err
	}

	if totalSupply.Cmp(ap.entireSupply) != 0 {
		return fmt.Errorf("%w for entire supply provided %s, computed %s",
			genesis.ErrEntireSupplyMismatch,
			ap.entireSupply.String(),
			totalSupply.String(),
		)
	}

	return nil
}

func (ap *accountsParser) parseElement(initialAccount *data.InitialAccount) error {
	if len(initialAccount.Address) == 0 {
		return genesis.ErrEmptyAddress
	}
	addressBytes, err := ap.pubkeyConverter.Decode(initialAccount.Address)
	if err != nil {
		return fmt.Errorf("%w for `%s`", genesis.ErrInvalidAddress, initialAccount.Address)
	}

	err = ap.keyGenerator.CheckPublicKeyValid(addressBytes)
	if err != nil {
		return fmt.Errorf("%w for `%s`, error: %s",
			genesis.ErrInvalidPubKey,
			initialAccount.Address,
			err.Error(),
		)
	}

	initialAccount.SetAddressBytes(addressBytes)

	return ap.parseDelegationElement(initialAccount)
}

func (ap *accountsParser) parseDelegationElement(initialAccount *data.InitialAccount) error {
	delegationData := initialAccount.Delegation

	if big.NewInt(0).Cmp(delegationData.Value) == 0 {
		return nil
	}

	if len(delegationData.Address) == 0 {
		return fmt.Errorf("%w for address '%s'",
			genesis.ErrEmptyDelegationAddress, initialAccount.Address)
	}
	addressBytes, err := ap.pubkeyConverter.Decode(delegationData.Address)
	if err != nil {
		return fmt.Errorf("%w for `%s`, address %s, error %s",
			genesis.ErrInvalidDelegationAddress,
			delegationData.Address,
			initialAccount.Address,
			err.Error(),
		)
	}

	delegationData.SetAddressBytes(addressBytes)

	return nil
}

func (ap *accountsParser) checkInitialAccount(initialAccount *data.InitialAccount) error {
	isSmartContract := core.IsSmartContractAddress(initialAccount.AddressBytes())
	if isSmartContract {
		return fmt.Errorf("%w for address %s",
			genesis.ErrAddressIsSmartContract,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Supply) >= 0 {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidSupply,
			initialAccount.Supply,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Balance) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidBalance,
			initialAccount.Balance,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.StakingValue) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidStakingBalance,
			initialAccount.Balance,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Delegation.Value) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidDelegationValue,
			initialAccount.Delegation.Value,
			initialAccount.Address,
		)
	}

	sum := big.NewInt(0)
	sum.Add(sum, initialAccount.Balance)
	sum.Add(sum, initialAccount.StakingValue)
	sum.Add(sum, initialAccount.Delegation.Value)

	isSupplyCorrect := big.NewInt(0).Cmp(initialAccount.Supply) < 0 && initialAccount.Supply.Cmp(sum) == 0
	if !isSupplyCorrect {
		return fmt.Errorf("%w for address %s, provided %s, computed %s",
			genesis.ErrSupplyMismatch,
			initialAccount.Address,
			initialAccount.Supply.String(),
			sum.String(),
		)
	}

	return nil
}

func (ap *accountsParser) checkForDuplicates() error {
	for idx1 := 0; idx1 < len(ap.initialAccounts); idx1++ {
		ia1 := ap.initialAccounts[idx1]
		for idx2 := idx1 + 1; idx2 < len(ap.initialAccounts); idx2++ {
			ia2 := ap.initialAccounts[idx2]
			if ia1.Address == ia2.Address {
				return fmt.Errorf("%w found for '%s'",
					genesis.ErrDuplicateAddress,
					ia1.Address,
				)
			}
		}
	}

	return nil
}

// InitialAccounts return the initial accounts contained by this parser
func (ap *accountsParser) InitialAccounts() []genesis.InitialAccountHandler {
	accounts := make([]genesis.InitialAccountHandler, len(ap.initialAccounts))

	for idx, ia := range ap.initialAccounts {
		accounts[idx] = ia
	}

	return accounts
}

// GenesisMintingAddress returns the encoded genesis minting address
func (ap *accountsParser) GenesisMintingAddress() string {
	return ap.pubkeyConverter.SilentEncode(ap.minterAddressBytes, log)
}

// InitialAccountsSplitOnAddressesShards gets the initial accounts of the nodes split on the addresses' shards
func (ap *accountsParser) InitialAccountsSplitOnAddressesShards(
	shardCoordinator sharding.Coordinator,
) (map[uint32][]genesis.InitialAccountHandler, error) {

	if check.IfNil(shardCoordinator) {
		return nil, genesis.ErrNilShardCoordinator
	}

	var addresses = make(map[uint32][]genesis.InitialAccountHandler)
	for _, ia := range ap.initialAccounts {
		shardID := shardCoordinator.ComputeId(ia.AddressBytes())

		addresses[shardID] = append(addresses[shardID], ia)
	}

	return addresses, nil
}

// GetTotalStakedForDelegationAddress returns the total staked value for a provided delegation address
func (ap *accountsParser) GetTotalStakedForDelegationAddress(delegationAddress string) *big.Int {
	sum := big.NewInt(0)

	for _, in := range ap.initialAccounts {
		if in.Delegation.Address == delegationAddress {
			sum.Add(sum, in.Delegation.Value)
		}
	}

	return sum
}

// GetInitialAccountsForDelegated returns the initial accounts that are set to the provided delegated address
func (ap *accountsParser) GetInitialAccountsForDelegated(addressBytes []byte) []genesis.InitialAccountHandler {
	list := make([]genesis.InitialAccountHandler, 0)
	for _, ia := range ap.initialAccounts {
		if bytes.Equal(ia.Delegation.AddressBytes(), addressBytes) {
			list = append(list, ia)
		}
	}

	return list
}

func (ap *accountsParser) createIndexerPools(shardIDs []uint32) map[uint32]*outportcore.TransactionPool {
	txsPoolPerShard := make(map[uint32]*outportcore.TransactionPool)

	for _, id := range shardIDs {
		txsPoolPerShard[id] = &outportcore.TransactionPool{
			Transactions:         make(map[string]*outportcore.TxInfo),
			SmartContractResults: make(map[string]*outportcore.SCRInfo),
		}
	}

	return txsPoolPerShard
}

func (ap *accountsParser) createMintTransactions() []coreData.TransactionHandler {
	txs := make([]coreData.TransactionHandler, 0, len(ap.initialAccounts))

	for nonce, ia := range ap.initialAccounts {
		tx := ap.createMintTransaction(ia, uint64(nonce))
		txs = append(txs, tx)
	}

	return txs
}

func (ap *accountsParser) createMintTransaction(ia genesis.InitialAccountHandler, nonce uint64) *transactionData.Transaction {
	tx := &transactionData.Transaction{
		Nonce:     nonce,
		SndAddr:   ap.minterAddressBytes,
		Value:     ia.GetSupply(),
		RcvAddr:   ia.AddressBytes(),
		GasPrice:  0,
		GasLimit:  0,
		Signature: []byte(common.GenesisTxSignatureString),
	}

	return tx
}

// TODO: extend sharding Coordinator with a similar function, GetShardIDs
func getShardIDs(shardCoordinator sharding.Coordinator) []uint32 {
	shardIDs := make([]uint32, shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		shardIDs[i] = i
	}
	shardIDs[shardCoordinator.NumberOfShards()] = core.MetachainShardId

	return shardIDs
}

func createMiniBlocks(shardIDs []uint32, blockType block.Type) []*block.MiniBlock {
	miniBlocks := make([]*block.MiniBlock, 0)

	for _, i := range shardIDs {
		for _, j := range shardIDs {
			miniBlock := &block.MiniBlock{
				TxHashes:        nil,
				ReceiverShardID: i,
				SenderShardID:   j,
				Type:            blockType,
			}

			miniBlocks = append(miniBlocks, miniBlock)
		}
	}

	return miniBlocks
}

func (ap *accountsParser) getAllTxs(
	indexingData map[uint32]*genesis.IndexingData,
) []coreData.TransactionHandler {
	allTxs := make([]coreData.TransactionHandler, 0)

	for _, txs := range indexingData {
		allTxs = append(allTxs, txs.DelegationTxs...)
		allTxs = append(allTxs, txs.StakingTxs...)
		allTxs = append(allTxs, txs.DeploySystemScTxs...)
		allTxs = append(allTxs, txs.DeployInitialScTxs...)
	}

	return allTxs
}

func (ap *accountsParser) setScrsTxsPool(
	shardCoordinator sharding.Coordinator,
	indexingData map[uint32]*genesis.IndexingData,
	txsPoolPerShard map[uint32]*outportcore.TransactionPool,
) {
	for _, id := range indexingData {
		for txHash, tx := range id.ScrsTxs {
			senderShardID := shardCoordinator.ComputeId(tx.GetSndAddr())
			receiverShardID := shardCoordinator.ComputeId(tx.GetRcvAddr())

			scrTx, ok := tx.(*scrData.SmartContractResult)
			if !ok {
				continue
			}
			scrTx.GasLimit = uint64(0)

			txsPoolPerShard[senderShardID].SmartContractResults[txHash] = &outportcore.SCRInfo{
				SmartContractResult: scrTx,
				FeeInfo:             &outportcore.FeeInfo{Fee: big.NewInt(0)},
			}
			txsPoolPerShard[receiverShardID].SmartContractResults[txHash] = &outportcore.SCRInfo{
				SmartContractResult: scrTx,
				FeeInfo:             &outportcore.FeeInfo{Fee: big.NewInt(0)},
			}
		}
	}
}

func (ap *accountsParser) setTxsPoolAndMiniBlocks(
	shardCoordinator sharding.Coordinator,
	allTxs []coreData.TransactionHandler,
	txsPoolPerShard map[uint32]*outportcore.TransactionPool,
	miniBlocks []*block.MiniBlock,
) error {

	for _, txHandler := range allTxs {
		receiverShardID := shardCoordinator.ComputeId(txHandler.GetRcvAddr())

		var senderShardID uint32
		if bytes.Equal(txHandler.GetSndAddr(), ap.minterAddressBytes) {
			senderShardID = core.MetachainShardId
		} else {
			senderShardID = shardCoordinator.ComputeId(txHandler.GetSndAddr())
		}

		txHash, err := core.CalculateHash(ap.marshalizer, ap.hasher, txHandler)
		if err != nil {
			return err
		}

		tx, ok := txHandler.(*transactionData.Transaction)
		if !ok {
			continue
		}
		tx.Signature = []byte(common.GenesisTxSignatureString)
		tx.GasLimit = uint64(0)

		txsPoolPerShard[senderShardID].Transactions[string(txHash)] = &outportcore.TxInfo{
			Transaction: tx,
			FeeInfo:     &outportcore.FeeInfo{Fee: big.NewInt(0)},
		}

		txsPoolPerShard[receiverShardID].Transactions[string(txHash)] = &outportcore.TxInfo{
			Transaction: tx,
			FeeInfo:     &outportcore.FeeInfo{Fee: big.NewInt(0)},
		}

		for _, miniBlock := range miniBlocks {
			if senderShardID == miniBlock.GetSenderShardID() &&
				receiverShardID == miniBlock.GetReceiverShardID() {
				miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
			}
		}
	}

	return nil
}

func getNonEmptyMiniBlocks(miniBlocks []*block.MiniBlock) []*block.MiniBlock {
	mbs := make([]*block.MiniBlock, 0)

	for _, mb := range miniBlocks {
		if len(mb.GetTxHashes()) > 0 {
			mbs = append(mbs, mb)
		}
	}

	return mbs
}

// GenerateInitialTransactions will generate initial transactions pool and the miniblocks for the generated transactions
func (ap *accountsParser) GenerateInitialTransactions(
	shardCoordinator sharding.Coordinator,
	indexingData map[uint32]*genesis.IndexingData,
) ([]*block.MiniBlock, map[uint32]*outportcore.TransactionPool, error) {
	if check.IfNil(shardCoordinator) {
		return nil, nil, genesis.ErrNilShardCoordinator
	}

	shardIDs := getShardIDs(shardCoordinator)
	txsPoolPerShard := ap.createIndexerPools(shardIDs)

	mintTxs := ap.createMintTransactions()

	allTxs := ap.getAllTxs(indexingData)
	allTxs = append(allTxs, mintTxs...)
	miniBlocks := createMiniBlocks(shardIDs, block.TxBlock)

	err := ap.setTxsPoolAndMiniBlocks(shardCoordinator, allTxs, txsPoolPerShard, miniBlocks)
	if err != nil {
		return nil, nil, err
	}

	ap.setScrsTxsPool(shardCoordinator, indexingData, txsPoolPerShard)

	miniBlocks = getNonEmptyMiniBlocks(miniBlocks)

	return miniBlocks, txsPoolPerShard, nil
}

// IsInterfaceNil returns true if the underlying object is nil
func (ap *accountsParser) IsInterfaceNil() bool {
	return ap == nil
}
