package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/fetcher"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"math/big"

	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

// ChainHandler defines what a chain handler should be able to do
type ChainHandler interface {
	IncrementRound()
	CreateNewBlock() error
	IsInterfaceNil() bool
}

// ChainFetcher defines what a chain fetcher should be able to do
type ChainFetcher interface {
	GetNetworkInfo() (*fetcher.NetworkInfo, error)
	GetAddressNonce(address string) (uint64, error)
	FetchAccount(address []byte, newAccount vmcommon.AccountHandler) (state.UserAccountHandler, error)
	FetchKeyFromGateway(address []byte, key []byte) ([]byte, error)
}

// ChainSimulator defines what a chain simulator should be able to do
type ChainSimulator interface {
	GenerateBlocks(numOfBlocks int) error
	GetNodeHandler(shardID uint32) process.NodeHandler
	GenerateAddressInShard(providedShardID uint32) dtos.WalletAddress
	GenerateAndMintWalletAddress(targetShardID uint32, value *big.Int) (dtos.WalletAddress, error)
	IsInterfaceNil() bool
}
