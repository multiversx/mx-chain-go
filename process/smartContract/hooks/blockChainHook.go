package hooks

import (
	"encoding/binary"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// BlockChainHookImpl is a wrapper over AccountsAdapter that satisfy vmcommon.BlockchainHook interface
type BlockChainHookImpl struct {
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

// NewBlockChainHookImpl creates a new BlockChainHookImpl instance
func NewBlockChainHookImpl(
	args ArgBlockChainHook,
) (*BlockChainHookImpl, error) {
	err := checkForNil(args)
	if err != nil {
		return nil, err
	}

	blockChainHookImpl := &BlockChainHookImpl{
		accounts:         args.Accounts,
		addrConv:         args.AddrConv,
		storageService:   args.StorageService,
		blockChain:       args.BlockChain,
		shardCoordinator: args.ShardCoordinator,
		marshalizer:      args.Marshalizer,
		uint64Converter:  args.Uint64Converter,
	}

	blockChainHookImpl.currentHdr = &block.Header{}
	blockChainHookImpl.tempAccounts = make(map[string]state.AccountHandler, 0)

	return blockChainHookImpl, nil
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
func (bh *BlockChainHookImpl) AccountExists(address []byte) (bool, error) {
	_, err := bh.getAccountFromAddressBytes(address)
	if err != nil {
		if err == state.ErrAccNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetBalance returns the balance of a shard account
func (bh *BlockChainHookImpl) GetBalance(address []byte) (*big.Int, error) {
	exists, err := bh.AccountExists(address)
	if err != nil {
		return nil, err
	}
	if !exists {
		return big.NewInt(0), nil
	}

	shardAccount, err := bh.getShardAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	return shardAccount.Balance, nil
}

// GetNonce returns the nonce of a shard account
func (bh *BlockChainHookImpl) GetNonce(address []byte) (uint64, error) {
	exists, err := bh.AccountExists(address)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	shardAccount, err := bh.getShardAccountFromAddressBytes(address)
	if err != nil {
		return 0, err
	}

	return shardAccount.Nonce, nil
}

// GetStorageData returns the storage value of a variable held in account's data trie
func (bh *BlockChainHookImpl) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	exists, err := bh.AccountExists(accountAddress)
	if err != nil {
		return nil, err
	}
	if !exists {
		return make([]byte, 0), nil
	}

	account, err := bh.getAccountFromAddressBytes(accountAddress)
	if err != nil {
		return nil, err
	}

	return account.DataTrieTracker().RetrieveValue(index)
}

// IsCodeEmpty returns if the code is empty
func (bh *BlockChainHookImpl) IsCodeEmpty(address []byte) (bool, error) {
	exists, err := bh.AccountExists(address)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}

	account, err := bh.getAccountFromAddressBytes(address)
	if err != nil {
		return false, err
	}

	isCodeEmpty := len(account.GetCode()) == 0
	return isCodeEmpty, nil
}

// GetCode retrieves the account's code
func (bh *BlockChainHookImpl) GetCode(address []byte) ([]byte, error) {
	account, err := bh.getAccountFromAddressBytes(address)
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
func (bh *BlockChainHookImpl) GetBlockhash(offset *big.Int) ([]byte, error) {
	if offset.Cmp(big.NewInt(0)) > 0 {
		return nil, process.ErrInvalidNonceRequest
	}

	hdr := bh.blockChain.GetCurrentBlockHeader()

	requestedNonce := big.NewInt(0).SetUint64(hdr.GetNonce())
	requestedNonce.Sub(requestedNonce, offset)

	if requestedNonce.Cmp(big.NewInt(0)) < 0 {
		return nil, process.ErrInvalidNonceRequest
	}

	if offset.Cmp(big.NewInt(0)) == 0 {
		return bh.blockChain.GetCurrentBlockHeaderHash(), nil
	}

	_, hash, err := process.GetHeaderFromStorageWithNonce(
		requestedNonce.Uint64(),
		bh.shardCoordinator.SelfId(),
		bh.storageService,
		bh.uint64Converter,
		bh.marshalizer,
	)

	if err != nil {
		return nil, err
	}

	return hash, nil
}

// LastNonce returns the nonce from from the last committed block
func (bh *BlockChainHookImpl) LastNonce() uint64 {
	if bh.blockChain.GetCurrentBlockHeader() != nil {
		return bh.blockChain.GetCurrentBlockHeader().GetNonce()
	}
	return 0
}

// LastRound returns the round from the last committed block
func (bh *BlockChainHookImpl) LastRound() uint64 {
	if bh.blockChain.GetCurrentBlockHeader() != nil {
		return bh.blockChain.GetCurrentBlockHeader().GetRound()
	}
	return 0
}

// LastTimeStamp returns the timeStamp from the last committed block
func (bh *BlockChainHookImpl) LastTimeStamp() uint64 {
	if bh.blockChain.GetCurrentBlockHeader() != nil {
		return bh.blockChain.GetCurrentBlockHeader().GetTimeStamp()
	}
	return 0
}

// LastRandomSeed returns the random seed from the last committed block
func (bh *BlockChainHookImpl) LastRandomSeed() []byte {
	if bh.blockChain.GetCurrentBlockHeader() != nil {
		return bh.blockChain.GetCurrentBlockHeader().GetRandSeed()
	}
	return []byte{}
}

// LastEpoch returns the epoch from the last committed block
func (bh *BlockChainHookImpl) LastEpoch() uint32 {
	if bh.blockChain.GetCurrentBlockHeader() != nil {
		return bh.blockChain.GetCurrentBlockHeader().GetEpoch()
	}
	return 0
}

// GetStateRootHash returns the state root hash from the last committed block
func (bh *BlockChainHookImpl) GetStateRootHash() []byte {
	if bh.blockChain.GetCurrentBlockHeader() != nil {
		return bh.blockChain.GetCurrentBlockHeader().GetRootHash()
	}
	return []byte{}
}

// CurrentNonce returns the nonce from the current block
func (bh *BlockChainHookImpl) CurrentNonce() uint64 {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()

	return bh.currentHdr.GetNonce()
}

// CurrentRound returns the round from the current block
func (bh *BlockChainHookImpl) CurrentRound() uint64 {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()
	return bh.currentHdr.GetRound()
}

// CurrentTimeStamp return the timestamp from the current block
func (bh *BlockChainHookImpl) CurrentTimeStamp() uint64 {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()
	return bh.currentHdr.GetTimeStamp()
}

// CurrentRandomSeed returns the random seed from the current header
func (bh *BlockChainHookImpl) CurrentRandomSeed() []byte {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()
	return bh.currentHdr.GetRandSeed()
}

// CurrentEpoch returns the current epoch
func (bh *BlockChainHookImpl) CurrentEpoch() uint32 {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()
	return bh.currentHdr.GetEpoch()
}

// NewAddress is a hook which creates a new smart contract address from the creators address and nonce
// The address is created by applied keccak256 on the appended value off creator address and nonce
// Prefix mask is applied for first 8 bytes 0, and for bytes 9-10 - VM type
// Suffix mask is applied - last 2 bytes are for the shard ID - mask is applied as suffix mask
func (bh *BlockChainHookImpl) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	addressLength := bh.addrConv.AddressLen()
	if len(creatorAddress) != addressLength {
		return nil, ErrAddressLengthNotCorrect
	}

	if len(vmType) != core.VMTypeLen {
		return nil, ErrVMTypeLengthIsNotCorrect
	}

	_, err := bh.getShardAccountFromAddressBytes(creatorAddress)
	if err != nil {
		return nil, err
	}

	base := hashFromAddressAndNonce(creatorAddress, creatorNonce)
	prefixMask := createPrefixMask(vmType)
	suffixMask := createSuffixMask(creatorAddress)

	copy(base[:core.NumInitCharactersForScAddress], prefixMask)
	copy(base[len(base)-core.ShardIdentiferLen:], suffixMask)

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
	prefixMask := make([]byte, core.NumInitCharactersForScAddress-core.VMTypeLen)
	prefixMask = append(prefixMask, vmType...)

	return prefixMask
}

func createSuffixMask(creatorAddress []byte) []byte {
	return creatorAddress[len(creatorAddress)-2:]
}

func (bh *BlockChainHookImpl) getAccountFromAddressBytes(address []byte) (state.AccountHandler, error) {
	tempAcc, success := bh.getAccountFromTemporaryAccounts(address)
	if success {
		return tempAcc, nil
	}

	addr, err := bh.addrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	return bh.accounts.GetExistingAccount(addr)
}

func (bh *BlockChainHookImpl) getShardAccountFromAddressBytes(address []byte) (*state.Account, error) {
	account, err := bh.getAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	shardAccount, ok := account.(*state.Account)
	if !ok {
		return nil, state.ErrWrongTypeAssertion
	}

	return shardAccount, nil
}

func (bh *BlockChainHookImpl) getAccountFromTemporaryAccounts(address []byte) (state.AccountHandler, bool) {
	bh.mutTempAccounts.Lock()
	defer bh.mutTempAccounts.Unlock()

	if tempAcc, ok := bh.tempAccounts[string(address)]; ok {
		return tempAcc, true
	}

	return nil, false
}

// AddTempAccount will add a temporary account in temporary store
func (bh *BlockChainHookImpl) AddTempAccount(address []byte, balance *big.Int, nonce uint64) {
	bh.mutTempAccounts.Lock()
	defer bh.mutTempAccounts.Unlock()

	accTracker := &TempAccountTracker{}
	addrContainer := state.NewAddress(address)
	account, err := state.NewAccount(addrContainer, accTracker)
	if err != nil {
		return
	}

	account.Balance = balance
	account.Nonce = nonce

	bh.tempAccounts[string(address)] = account
}

// CleanTempAccounts cleans the map holding the temporary accounts
func (bh *BlockChainHookImpl) CleanTempAccounts() {
	bh.mutTempAccounts.Lock()
	bh.tempAccounts = make(map[string]state.AccountHandler, 0)
	bh.mutTempAccounts.Unlock()
}

// TempAccount can retrieve a temporary account from provided address
func (bh *BlockChainHookImpl) TempAccount(address []byte) state.AccountHandler {
	tempAcc, success := bh.getAccountFromTemporaryAccounts(address)
	if success {
		return tempAcc
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bh *BlockChainHookImpl) IsInterfaceNil() bool {
	if bh == nil {
		return true
	}
	return false
}

// SetCurrentHeader sets current header to be used by smart contracts
func (bh *BlockChainHookImpl) SetCurrentHeader(hdr data.HeaderHandler) {
	if hdr == nil || hdr.IsInterfaceNil() {
		return
	}

	bh.mutCurrentHdr.Lock()
	bh.currentHdr = hdr
	bh.mutCurrentHdr.Unlock()
}
