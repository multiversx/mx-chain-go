package hooks

import (
	"encoding/binary"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// VMAccountsDB is a wrapper over AccountsAdapter that satisfy vmcommon.BlockchainHook interface
type VMAccountsDB struct {
	accounts         state.AccountsAdapter
	addrConv         state.AddressConverter
	storageService   dataRetriever.StorageService
	blockChain       data.ChainHandler
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	uint64Converter  typeConverters.Uint64ByteSliceConverter

	mutCurrentHdr sync.RWMutex
	currentHdr    data.HeaderHandler

	mutTempAccounts sync.Mutex
	tempAccounts    map[string]state.AccountHandler
}

// NewVMAccountsDB creates a new VMAccountsDB instance
func NewVMAccountsDB(
	args ArgBlockChainHook,
) (*VMAccountsDB, error) {
	err := checkForNil(args)
	if err != nil {
		return nil, err
	}

	vmAccountsDB := &VMAccountsDB{
		accounts:         args.Accounts,
		addrConv:         args.AddrConv,
		storageService:   args.StorageService,
		blockChain:       args.BlockChain,
		shardCoordinator: args.ShardCoordinator,
		marshalizer:      args.Marshalizer,
		uint64Converter:  args.Uint64Converter,
	}

	vmAccountsDB.tempAccounts = make(map[string]state.AccountHandler, 0)

	vmAccountsDB.mutCurrentHdr.Lock()
	if vmAccountsDB.shardCoordinator.SelfId() == sharding.MetachainShardId {
		vmAccountsDB.currentHdr = &block.MetaBlock{}
	} else {
		vmAccountsDB.currentHdr = &block.Header{}
	}
	vmAccountsDB.mutCurrentHdr.Unlock()

	return vmAccountsDB, nil
}

func checkForNil(args ArgBlockChainHook) error {
	if args.Accounts == nil || args.Accounts.IsInterfaceNil() {
		return process.ErrNilAccountsAdapter
	}
	if args.AddrConv == nil || args.AddrConv.IsInterfaceNil() {
		return process.ErrNilAddressConverter
	}
	if args.StorageService == nil || args.StorageService.IsInterfaceNil() {
		return process.ErrNilStorage
	}
	if args.BlockChain == nil || args.BlockChain.IsInterfaceNil() {
		return process.ErrNilBlockChain
	}
	if args.ShardCoordinator == nil || args.ShardCoordinator.IsInterfaceNil() {
		return process.ErrNilShardCoordinator
	}
	if args.Marshalizer == nil || args.Marshalizer.IsInterfaceNil() {
		return process.ErrNilMarshalizer
	}
	if args.Uint64Converter == nil || args.Uint64Converter.IsInterfaceNil() {
		return process.ErrNilUint64Converter
	}
	return nil
}

// AccountExists checks if an account exists in provided AccountAdapter
func (vadb *VMAccountsDB) AccountExists(address []byte) (bool, error) {
	_, err := vadb.getAccountFromAddressBytes(address)
	if err != nil {
		if err == state.ErrAccNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetBalance returns the balance of a shard account
func (vadb *VMAccountsDB) GetBalance(address []byte) (*big.Int, error) {
	exists, err := vadb.AccountExists(address)
	if err != nil {
		return nil, err
	}
	if !exists {
		return big.NewInt(0), nil
	}

	shardAccount, err := vadb.getShardAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	return shardAccount.Balance, nil
}

// GetNonce returns the nonce of a shard account
func (vadb *VMAccountsDB) GetNonce(address []byte) (uint64, error) {
	exists, err := vadb.AccountExists(address)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	shardAccount, err := vadb.getShardAccountFromAddressBytes(address)
	if err != nil {
		return 0, err
	}

	return shardAccount.Nonce, nil
}

// GetStorageData returns the storage value of a variable held in account's data trie
func (vadb *VMAccountsDB) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	exists, err := vadb.AccountExists(accountAddress)
	if err != nil {
		return nil, err
	}
	if !exists {
		return make([]byte, 0), nil
	}

	account, err := vadb.getAccountFromAddressBytes(accountAddress)
	if err != nil {
		return nil, err
	}

	return account.DataTrieTracker().RetrieveValue(index)
}

// IsCodeEmpty returns if the code is empty
func (vadb *VMAccountsDB) IsCodeEmpty(address []byte) (bool, error) {
	exists, err := vadb.AccountExists(address)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}

	account, err := vadb.getAccountFromAddressBytes(address)
	if err != nil {
		return false, err
	}

	isCodeEmpty := len(account.GetCode()) == 0
	return isCodeEmpty, nil
}

// GetCode retrieves the account's code
func (vadb *VMAccountsDB) GetCode(address []byte) ([]byte, error) {
	account, err := vadb.getAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	code := account.GetCode()
	if len(code) == 0 {
		return nil, ErrEmptyCode
	}

	return code, nil
}

// GetBlockhash returns the header hash for a requested nonce delta
func (vadb *VMAccountsDB) GetBlockhash(offset *big.Int) ([]byte, error) {
	if offset.Cmp(big.NewInt(0)) > 0 {
		return nil, process.ErrInvalidNonceRequest
	}

	hdr := vadb.blockChain.GetCurrentBlockHeader()

	requestedNonce := big.NewInt(0).SetUint64(hdr.GetNonce())
	requestedNonce.Sub(requestedNonce, offset)

	if requestedNonce.Cmp(big.NewInt(0)) < 0 {
		return nil, process.ErrInvalidNonceRequest
	}

	if offset.Cmp(big.NewInt(0)) == 0 {
		return vadb.blockChain.GetCurrentBlockHeaderHash(), nil
	}

	if vadb.shardCoordinator.SelfId() == sharding.MetachainShardId {
		_, hash, err := process.GetMetaHeaderFromStorageWithNonce(
			requestedNonce.Uint64(),
			vadb.storageService,
			vadb.uint64Converter,
			vadb.marshalizer,
		)
		if err != nil {
			return nil, err
		}

		return hash, nil
	}

	_, hash, err := process.GetShardHeaderFromStorageWithNonce(
		requestedNonce.Uint64(),
		vadb.shardCoordinator.SelfId(),
		vadb.storageService,
		vadb.uint64Converter,
		vadb.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// LastNonce returns the nonce from from the last committed block
func (vadb *VMAccountsDB) LastNonce() uint64 {
	if vadb.blockChain.GetCurrentBlockHeader() != nil {
		return vadb.blockChain.GetCurrentBlockHeader().GetNonce()
	}
	return 0
}

// LastRound returns the round from the last committed block
func (vadb *VMAccountsDB) LastRound() uint64 {
	if vadb.blockChain.GetCurrentBlockHeader() != nil {
		return vadb.blockChain.GetCurrentBlockHeader().GetRound()
	}
	return 0
}

// LastTimeStamp returns the timeStamp from the last committed block
func (vadb *VMAccountsDB) LastTimeStamp() uint64 {
	if vadb.blockChain.GetCurrentBlockHeader() != nil {
		return vadb.blockChain.GetCurrentBlockHeader().GetTimeStamp()
	}
	return 0
}

// LastRandomSeed returns the random seed from the last committed block
func (vadb *VMAccountsDB) LastRandomSeed() []byte {
	if vadb.blockChain.GetCurrentBlockHeader() != nil {
		return vadb.blockChain.GetCurrentBlockHeader().GetRandSeed()
	}
	return []byte{}
}

// LastEpoch returns the epoch from the last committed block
func (vadb *VMAccountsDB) LastEpoch() uint32 {
	if vadb.blockChain.GetCurrentBlockHeader() != nil {
		return vadb.blockChain.GetCurrentBlockHeader().GetEpoch()
	}
	return 0
}

// GetStateRootHash returns the state root hash from the last committed block
func (vadb *VMAccountsDB) GetStateRootHash() []byte {
	if vadb.blockChain.GetCurrentBlockHeader() != nil {
		return vadb.blockChain.GetCurrentBlockHeader().GetRootHash()
	}
	return []byte{}
}

// CurrentNonce returns the nonce from the current block
func (vadb *VMAccountsDB) CurrentNonce() uint64 {
	vadb.mutCurrentHdr.RLock()
	defer vadb.mutCurrentHdr.RUnlock()
	return vadb.currentHdr.GetNonce()
}

// CurrentRound returns the round from the current block
func (vadb *VMAccountsDB) CurrentRound() uint64 {
	vadb.mutCurrentHdr.RLock()
	defer vadb.mutCurrentHdr.RUnlock()
	return vadb.currentHdr.GetRound()
}

// CurrentTimeStamp return the timestamp from the current block
func (vadb *VMAccountsDB) CurrentTimeStamp() uint64 {
	vadb.mutCurrentHdr.RLock()
	defer vadb.mutCurrentHdr.RUnlock()
	return vadb.currentHdr.GetTimeStamp()
}

// CurrentRandomSeed returns the random seed from the current header
func (vadb *VMAccountsDB) CurrentRandomSeed() []byte {
	vadb.mutCurrentHdr.RLock()
	defer vadb.mutCurrentHdr.RUnlock()
	return vadb.currentHdr.GetRandSeed()
}

// CurrentEpoch returns the current epoch
func (vadb *VMAccountsDB) CurrentEpoch() uint32 {
	vadb.mutCurrentHdr.RLock()
	defer vadb.mutCurrentHdr.RUnlock()
	return vadb.currentHdr.GetEpoch()
}

// NewAddress is a hook which creates a new smart contract address from the creators address and nonce
// The address is created by applied keccak256 on the appended value off creator address and nonce
// Prefix mask is applied for first 8 bytes 0, and for bytes 9-10 - VM type
// Suffix mask is applied - last 2 bytes are for the shard ID - mask is applied as suffix mask
func (vadb *VMAccountsDB) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	addressLength := vadb.addrConv.AddressLen()
	if len(creatorAddress) != addressLength {
		return nil, ErrAddressLengthNotCorrect
	}

	if len(vmType) != VMTypeLen {
		return nil, ErrVMTypeLengthIsNotCorrect
	}

	_, err := vadb.getShardAccountFromAddressBytes(creatorAddress)
	if err != nil {
		return nil, err
	}

	base := hashFromAddressAndNonce(creatorAddress, creatorNonce)
	prefixMask := createPrefixMask(vmType)
	suffixMask := createSuffixMask(creatorAddress)

	copy(base[:NumInitCharactersForScAddress], prefixMask)
	copy(base[len(base)-ShardIdentiferLen:], suffixMask)

	return base, nil
}

func hashFromAddressAndNonce(creatorAddress []byte, creatorNonce uint64) []byte {
	buffNonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffNonce, creatorNonce)
	adrAndNonce := append(creatorAddress, buffNonce...)
	scAddress := keccak.Keccak{}.Compute(string(adrAndNonce))

	return scAddress
}

func createPrefixMask(vmType []byte) []byte {
	prefixMask := make([]byte, NumInitCharactersForScAddress-VMTypeLen)
	prefixMask = append(prefixMask, vmType...)

	return prefixMask
}

func createSuffixMask(creatorAddress []byte) []byte {
	return creatorAddress[len(creatorAddress)-2:]
}

func (vadb *VMAccountsDB) getAccountFromAddressBytes(address []byte) (state.AccountHandler, error) {
	tempAcc, success := vadb.getAccountFromTemporaryAccounts(address)
	if success {
		return tempAcc, nil
	}

	addr, err := vadb.addrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	return vadb.accounts.GetExistingAccount(addr)
}

func (vadb *VMAccountsDB) getShardAccountFromAddressBytes(address []byte) (*state.Account, error) {
	account, err := vadb.getAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	shardAccount, ok := account.(*state.Account)
	if !ok {
		return nil, state.ErrWrongTypeAssertion
	}

	return shardAccount, nil
}

func (vadb *VMAccountsDB) getAccountFromTemporaryAccounts(address []byte) (state.AccountHandler, bool) {
	vadb.mutTempAccounts.Lock()
	defer vadb.mutTempAccounts.Unlock()

	if tempAcc, ok := vadb.tempAccounts[string(address)]; ok {
		return tempAcc, true
	}

	return nil, false
}

// AddTempAccount will add a temporary account in temporary store
func (vadb *VMAccountsDB) AddTempAccount(address []byte, balance *big.Int, nonce uint64) {
	vadb.mutTempAccounts.Lock()
	vadb.tempAccounts[string(address)] = &state.Account{Balance: balance, Nonce: nonce}
	vadb.mutTempAccounts.Unlock()
}

// CleanTempAccounts cleans the map holding the temporary accounts
func (vadb *VMAccountsDB) CleanTempAccounts() {
	vadb.mutTempAccounts.Lock()
	vadb.tempAccounts = make(map[string]state.AccountHandler, 0)
	vadb.mutTempAccounts.Unlock()
}

// TempAccount can retrieve a temporary account from provided address
func (vadb *VMAccountsDB) TempAccount(address []byte) state.AccountHandler {
	tempAcc, success := vadb.getAccountFromTemporaryAccounts(address)
	if success {
		return tempAcc
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (vadb *VMAccountsDB) IsInterfaceNil() bool {
	if vadb == nil {
		return true
	}
	return false
}

// SetCurrentHeader sets current header to be used by smart contracts
func (vadb *VMAccountsDB) SetCurrentHeader(hdr data.HeaderHandler) {
	vadb.mutCurrentHdr.Lock()
	vadb.currentHdr = hdr
	vadb.mutCurrentHdr.Unlock()
}
