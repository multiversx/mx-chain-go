package state

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	chainData "github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsRepository struct {
	chainHandler       chainData.ChainHandler
	strategyForFinal   *accountsRepositoryStrategy
	strategyForCurrent *accountsRepositoryStrategy
}

// NewAccountsRepository creates a new accountsRepository
func NewAccountsRepository(
	chainHandler chainData.ChainHandler,
	innerAccountsAdapterForFinal AccountsAdapter,
	innerAccountsAdapterForCurrent AccountsAdapter,
) (*accountsRepository, error) {
	if check.IfNil(chainHandler) {
		return nil, ErrNilChainHandler
	}
	if check.IfNil(innerAccountsAdapterForFinal) {
		return nil, ErrNilAccountsAdapter
	}
	if check.IfNil(innerAccountsAdapterForCurrent) {
		return nil, ErrNilAccountsAdapter
	}

	repository := &accountsRepository{chainHandler: chainHandler}
	repository.strategyForFinal = newAccountsRepositoryStrategy(repository.getFinalBlockInfo, innerAccountsAdapterForFinal)
	repository.strategyForCurrent = newAccountsRepositoryStrategy(repository.getCurrentBlockInfo, innerAccountsAdapterForCurrent)

	return repository, nil
}

func (repository *accountsRepository) getFinalBlockInfo() (*accountBlockInfo, error) {
	nonce, hash, rootHash := repository.chainHandler.GetFinalBlockInfo()
	if len(hash) == 0 || len(rootHash) == 0 {
		return nil, ErrBlockInfoNotAvailable
	}

	return &accountBlockInfo{
		nonce:    nonce,
		hash:     hash,
		rootHash: rootHash,
	}, nil
}

func (repository *accountsRepository) getCurrentBlockInfo() (*accountBlockInfo, error) {
	block := repository.chainHandler.GetCurrentBlockHeader()
	if check.IfNil(block) {
		return nil, ErrBlockInfoNotAvailable
	}

	hash := repository.chainHandler.GetCurrentBlockHeaderHash()
	if len(hash) == 0 {
		return nil, ErrBlockInfoNotAvailable
	}

	rootHash := repository.chainHandler.GetCurrentBlockRootHash()
	if len(rootHash) == 0 {
		return nil, ErrBlockInfoNotAvailable
	}

	return &accountBlockInfo{
		nonce:    block.GetNonce(),
		hash:     hash,
		rootHash: rootHash,
	}, nil
}

// GetAccountOnFinal returns the account data as found on the latest final rootHash
func (repository *accountsRepository) GetAccountOnFinal(address []byte) (vmcommon.AccountHandler, AccountBlockInfo, error) {
	return repository.strategyForFinal.getAccount(address)
}

// GetCodeOnFinal returns the code as found on the latest final rootHash
func (repository *accountsRepository) GetCodeOnFinal(codeHash []byte) ([]byte, AccountBlockInfo) {
	return repository.strategyForFinal.getCode(codeHash)
}

// GetAccountOnCurrent returns the account data as found on the current rootHash
func (repository *accountsRepository) GetAccountOnCurrent(address []byte) (vmcommon.AccountHandler, AccountBlockInfo, error) {
	return repository.strategyForCurrent.getAccount(address)
}

// GetCodeOnCurrent returns the code as found on the the current rootHash
func (repository *accountsRepository) GetCodeOnCurrent(codeHash []byte) ([]byte, AccountBlockInfo) {
	return repository.strategyForCurrent.getCode(codeHash)
}

// Close will handle the closing of the underlying components
func (repository *accountsRepository) Close() error {
	// Question for review: is it all right?
	errFinal := repository.strategyForFinal.close()
	errCurrent := repository.strategyForCurrent.close()

	if errFinal != nil {
		return errFinal
	}
	if errCurrent != nil {
		return errCurrent
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (repository *accountsRepository) IsInterfaceNil() bool {
	return repository == nil
}
