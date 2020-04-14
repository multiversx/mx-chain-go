package bootstrap

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

const baseErrorMessage = "error with epoch start bootstrapper arguments"

func checkArguments(args ArgsEpochStartBootstrap) error {
	if check.IfNil(args.PathManager) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilPathManager)
	}
	if check.IfNil(args.GenesisShardCoordinator) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilShardCoordinator)
	}
	if check.IfNil(args.Messenger) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilMessenger)
	}
	if check.IfNil(args.EconomicsData) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilEconomicsData)
	}
	if check.IfNil(args.PublicKey) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilPubKey)
	}
	if check.IfNil(args.Hasher) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilHasher)
	}
	if check.IfNil(args.Marshalizer) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilMarshalizer)
	}
	if check.IfNil(args.BlockKeyGen) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilBlockKeyGen)
	}
	if check.IfNil(args.KeyGen) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilKeyGen)
	}
	if check.IfNil(args.SingleSigner) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilSingleSigner)
	}
	if check.IfNil(args.BlockSingleSigner) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilBlockSingleSigner)
	}
	if check.IfNil(args.TxSignMarshalizer) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilTxSignMarshalizer)
	}
	if args.GenesisNodesConfig == nil {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilGenesisNodesConfig)
	}
	if check.IfNil(args.Rater) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilRater)
	}
	if args.TrieStorageManagers == nil {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilTrieStorageManager)
	}
	if check.IfNil(args.TrieContainer) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilTriesContainer)
	}
	if len(args.DefaultDBPath) == 0 {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrInvalidDefaultDBPath)
	}
	if check.IfNil(args.AddressPubkeyConverter) {
		return fmt.Errorf("%w: %s", epochStart.ErrNilPubkeyConverter, baseErrorMessage)
	}
	if len(args.DefaultEpochString) == 0 {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrInvalidDefaultEpochString)
	}
	if len(args.DefaultShardString) == 0 {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrInvalidDefaultShardString)
	}
	if len(args.WorkingDir) == 0 {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrInvalidWorkingDir)
	}
	if check.IfNil(args.Rounder) {
		return epochStart.ErrNilRounder
	}

	return nil
}
