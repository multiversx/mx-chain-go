package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// TODO: increase code coverage with unit tests

type peer struct {
	storage          dataRetriever.StorageService
	accounts         state.AccountsAdapter
	blockType        block.Type
	store            dataRetriever.StorageService
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	addressConverter state.AddressConverter
}

// NewTransactionPreprocessor creates a new transaction preprocessor object
func NewPeerPreprocessor(
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accounts state.AccountsAdapter,
	addressConverter state.AddressConverter,
	blockType block.Type,
) (*peer, error) {

	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(store) {
		return nil, process.ErrNilTxStorage
	}

	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}

	txs := peer{
		storage:          store,
		accounts:         accounts,
		blockType:        blockType,
		store:            store,
		hasher:           hasher,
		marshalizer:      marshalizer,
		addressConverter: addressConverter,
	}

	return &txs, nil
}

// IsDataPrepared returns non error if all the requested transactions arrived and were saved into the pool
func (txs *peer) IsDataPrepared(requestedTxs int, haveTime func() time.Duration) error {
	return nil
}

// RemoveTxBlockFromPools removes transactions and miniblocks from associated pools
func (txs *peer) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	return nil
}

// RestoreTxBlockIntoPools restores the transactions and miniblocks to associated pools
func (txs *peer) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
	return 0, nil
}

// ProcessBlockTransactions processes all the transaction from the block.Body, updates the state
func (txs *peer) ProcessBlockTransactions(
	body block.Body,
	haveTime func() bool,
) error {
	// basic validation already done in interceptors
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != txs.blockType {
			continue
		}

		log.Trace("Processing peerBlock")
		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if !haveTime() {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]

			vid := state.ValidatorInfoData{}
			err := txs.marshalizer.Unmarshal(&vid, txHash)
			if err != nil {
				return err
			}

			addressContainer, err := txs.addressConverter.CreateAddressFromPublicKeyBytes(vid.PublicKey)
			if err != nil {
				return err
			}

			account, err := txs.accounts.GetAccountWithJournal(addressContainer)
			if err != nil {
				return err
			}

			peerAccount, ok := account.(state.PeerAccountHandler)
			if !ok {
				return process.ErrInvalidPeerAccount
			}

			err = peerAccount.SetTempRatingWithJournal(vid.TempRating)
			err = peerAccount.SetRatingWithJournal(vid.Rating)

			log.Trace("Setting infos", "pk", vid.PublicKey, "rating", vid.Rating, "tempRating", vid.TempRating)

		}

		log.Trace("Finished processing peerBlock")
	}

	return nil
}

func (txs *peer) CreateBlockStarted() {}

func (txs *peer) SaveTxBlockToStorage(body block.Body) error { return nil }

func (txs *peer) RequestBlockTransactions(body block.Body) int { return 0 }

func (txs *peer) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) { return nil, nil }

func (txs *peer) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int          { return 0 }
func (txs *peer) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool) error { return nil }
func (txs *peer) CreateAndProcessMiniBlocks(maxTxSpaceRemained uint32, maxMbSpaceRemained uint32, haveTime func() bool) (block.MiniBlockSlice, error) {
	return nil, nil
}

func (txs *peer) GetAllCurrentUsedTxs() map[string]data.TransactionHandler { return nil }

// IsInterfaceNil returns true if there is no value under the interface
func (txs *peer) IsInterfaceNil() bool {
	return txs == nil
}
