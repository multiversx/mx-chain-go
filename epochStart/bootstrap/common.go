package bootstrap

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

const baseErrorMessage = "error with epoch start bootstrapper arguments"

func checkArguments(args ArgsEpochStartBootstrap) error {
	if check.IfNil(args.GenesisShardCoordinator) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilShardCoordinator)
	}
	if check.IfNil(args.MainMessenger) {
		return fmt.Errorf("%s on main network: %w", baseErrorMessage, epochStart.ErrNilMessenger)
	}
	if check.IfNil(args.FullArchiveMessenger) {
		return fmt.Errorf("%s on full archive network: %w", baseErrorMessage, epochStart.ErrNilMessenger)
	}
	if check.IfNil(args.EconomicsData) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilEconomicsData)
	}
	if check.IfNil(args.CoreComponentsHolder) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilCoreComponentsHolder)
	}
	if check.IfNil(args.CryptoComponentsHolder) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilCryptoComponentsHolder)
	}
	if check.IfNil(args.CryptoComponentsHolder.PublicKey()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilPubKey)
	}
	if check.IfNil(args.CoreComponentsHolder.Hasher()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilHasher)
	}
	if check.IfNil(args.CoreComponentsHolder.InternalMarshalizer()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilMarshalizer)
	}
	if check.IfNil(args.CryptoComponentsHolder.BlockSignKeyGen()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilBlockKeyGen)
	}
	if check.IfNil(args.CryptoComponentsHolder.TxSignKeyGen()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilKeyGen)
	}
	if check.IfNil(args.CryptoComponentsHolder.TxSingleSigner()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilSingleSigner)
	}
	if check.IfNil(args.CryptoComponentsHolder.BlockSigner()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilBlockSingleSigner)
	}
	if check.IfNil(args.CoreComponentsHolder.TxMarshalizer()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilTxSignMarshalizer)
	}
	if check.IfNil(args.CoreComponentsHolder.PathHandler()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilPathManager)
	}
	if check.IfNil(args.TrieSyncStatisticsProvider) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilTrieSyncStatistics)
	}
	if args.GenesisNodesConfig == nil {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilGenesisNodesConfig)
	}
	if check.IfNil(args.Rater) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilRater)
	}
	if check.IfNil(args.CoreComponentsHolder.AddressPubKeyConverter()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilPubkeyConverter)
	}
	if check.IfNil(args.RoundHandler) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilRoundHandler)
	}
	if check.IfNil(args.StorageUnitOpener) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilStorageUnitOpener)
	}
	if check.IfNil(args.LatestStorageDataProvider) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilLatestStorageDataProvider)
	}
	if check.IfNil(args.CoreComponentsHolder.Uint64ByteSliceConverter()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilUint64Converter)
	}
	if check.IfNil(args.NodeShuffler) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilShuffler)
	}
	if args.GeneralConfig.EpochStartConfig.MinNumOfPeersToConsiderBlockValid < minNumPeersToConsiderMetaBlockValid {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNotEnoughNumOfPeersToConsiderBlockValid)
	}
	if args.GeneralConfig.EpochStartConfig.MinNumConnectedPeersToStart < minNumConnectedPeers {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNotEnoughNumConnectedPeers)
	}
	if check.IfNil(args.ArgumentsParser) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilArgumentsParser)
	}
	if check.IfNil(args.StatusHandler) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilStatusHandler)
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilHeaderIntegrityVerifier)
	}
	if check.IfNil(args.DataSyncerCreator) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilScheduledDataSyncerFactory)
	}
	if check.IfNil(args.CoreComponentsHolder.TxSignHasher()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilHasher)
	}
	if check.IfNil(args.CoreComponentsHolder.EpochNotifier()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilEpochNotifier)
	}
	if args.GeneralConfig.TrieSync.MaxHardCapForMissingNodes < 1 {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrInvalidMaxHardCapForMissingNodes)
	}
	if args.GeneralConfig.TrieSync.NumConcurrentTrieSyncers < 1 {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrInvalidNumConcurrentTrieSyncers)
	}
	if check.IfNil(args.CryptoComponentsHolder.ManagedPeersHolder()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, epochStart.ErrNilManagedPeersHolder)
	}
	if check.IfNil(args.StateStatsHandler) {
		return fmt.Errorf("%s: %w", baseErrorMessage, statistics.ErrNilStateStatsHandler)
	}
	if check.IfNil(args.NodesCoordinatorRegistryFactory) {
		return fmt.Errorf("%s: %w", baseErrorMessage, nodesCoordinator.ErrNilNodesCoordinatorRegistryFactory)
	}
	if check.IfNil(args.RunTypeComponents) {
		return fmt.Errorf("%s: %w", baseErrorMessage, errors.ErrNilRunTypeComponents)
	}
	if check.IfNil(args.RunTypeComponents.ShardCoordinatorCreator()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, errors.ErrNilShardCoordinatorFactory)
	}
	if check.IfNil(args.RunTypeComponents.AdditionalStorageServiceCreator()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, errors.ErrNilAdditionalStorageServiceCreator)
	}
	if check.IfNil(args.RunTypeComponents.NodesCoordinatorWithRaterCreator()) {
		return fmt.Errorf("%s: %w", baseErrorMessage, errors.ErrNilNodesCoordinatorFactory)
	}

	return nil
}
