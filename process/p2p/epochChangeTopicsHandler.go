package p2p

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type topicResolver interface {
	p2p.MessageProcessor
	RequestTopic() string
}

var log = logger.GetOrCreate("epochChangeTopicsHandler")

// ArgEpochChangeTopicsHandler is the DTO used to create a new instance of epochChangeTopicsHandler
type ArgEpochChangeTopicsHandler struct {
	MainMessenger                    p2p.Messenger
	FullArchiveMessenger             p2p.Messenger
	EnableEpochsHandler              common.EnableEpochsHandler
	EpochNotifier                    process.EpochNotifier
	ShardCoordinator                 sharding.Coordinator
	ResolversContainer               dataRetriever.ResolversContainer
	MainInterceptorsContainer        process.InterceptorsContainer
	FullArchiveInterceptorsContainer process.InterceptorsContainer
	IsFullArchive                    bool
}

// TODO[Sorin]: add unittests
type epochChangeTopicsHandler struct {
	mainMessenger                    p2p.Messenger
	fullArchiveMessenger             p2p.Messenger
	enableEpochsHandler              common.EnableEpochsHandler
	shardCoordinator                 sharding.Coordinator
	resolversContainer               dataRetriever.ResolversContainer
	mainInterceptorsContainer        process.InterceptorsContainer
	fullArchiveInterceptorsContainer process.InterceptorsContainer
	isFullArchive                    bool
}

// NewEpochChangeTopicsHandler creates a new instance of epochChangeTopicsHandler
func NewEpochChangeTopicsHandler(args ArgEpochChangeTopicsHandler) (*epochChangeTopicsHandler, error) {
	err := checkArgEpochChangeTopicsHandler(args)
	if err != nil {
		return nil, err
	}

	handler := &epochChangeTopicsHandler{
		mainMessenger:                    args.MainMessenger,
		fullArchiveMessenger:             args.FullArchiveMessenger,
		enableEpochsHandler:              args.EnableEpochsHandler,
		shardCoordinator:                 args.ShardCoordinator,
		resolversContainer:               args.ResolversContainer,
		mainInterceptorsContainer:        args.MainInterceptorsContainer,
		fullArchiveInterceptorsContainer: args.FullArchiveInterceptorsContainer,
		isFullArchive:                    args.IsFullArchive,
	}
	args.EpochNotifier.RegisterNotifyHandler(handler)

	return handler, nil
}

func checkArgEpochChangeTopicsHandler(args ArgEpochChangeTopicsHandler) error {
	if check.IfNil(args.MainMessenger) {
		return fmt.Errorf("%w for main messenger", process.ErrNilMessenger)
	}
	if check.IfNil(args.FullArchiveMessenger) {
		return fmt.Errorf("%w for full archive messenger", process.ErrNilMessenger)
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.ResolversContainer) {
		return process.ErrNilResolversContainer
	}
	if check.IfNil(args.MainInterceptorsContainer) {
		return fmt.Errorf("%w for main container", process.ErrNilInterceptorsContainer)
	}
	if check.IfNil(args.FullArchiveInterceptorsContainer) {
		return fmt.Errorf("%w for full archive container", process.ErrNilInterceptorsContainer)
	}

	return nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (handler *epochChangeTopicsHandler) EpochConfirmed(epoch uint32, _ uint64) {
	// TODO[Sorin]: proper Supernova flag
	supernovaEpoch := handler.enableEpochsHandler.GetActivationEpoch(common.AndromedaFlag)
	if epoch != supernovaEpoch {
		return
	}

	log.Debug("moving transaction topics on the new network...")

	err := handler.replaceProcessorsForTopic(common.TransactionTopic)
	if err != nil {
		log.Warn("replaceProcessorsForTopic failed", "topic", common.TransactionTopic, "err", err)
		return
	}

	err = handler.replaceProcessorsForTopic(common.UnsignedTransactionTopic)
	if err != nil {
		log.Warn("replaceProcessorsForTopic failed", "topic", common.UnsignedTransactionTopic, "err", err)
		return
	}

	err = handler.replaceProcessorsForTopic(common.RewardsTransactionTopic)
	if err != nil {
		log.Warn("replaceProcessorsForTopic failed", "topic", common.RewardsTransactionTopic, "err", err)
		return
	}
}

func (handler *epochChangeTopicsHandler) replaceProcessorsForTopic(topic string) error {
	err := handler.replaceTxInterceptorsOnTxNetwork(topic)
	if err != nil {
		return err
	}

	return handler.replaceTxResolversOnTxNetwork(topic)
}

func removeTopicFromMessenger(messenger p2p.MessageHandler, topic string) error {
	err := messenger.UnregisterMessageProcessor(topic, common.DefaultInterceptorsIdentifier)
	if err != nil {
		return err
	}
	err = messenger.UnregisterMessageProcessor(topic, common.HardforkInterceptorsIdentifier)
	if err != nil {
		return err
	}
	err = messenger.UnJoinTopic(topic)
	if err != nil {
		return err
	}

	return nil
}

func (handler *epochChangeTopicsHandler) replaceTxResolversOnTxNetwork(
	topic string,
) error {
	if topic == common.RewardsTransactionTopic {
		return handler.registerRewardTxResolversOnTxNetwork()
	}

	shardC := handler.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		// same old resolver will be used, already unregistered from the messengers
		oldResolver, err := handler.resolversContainer.Get(identifierTx)
		if err != nil {
			return err
		}

		resolver, ok := oldResolver.(topicResolver)
		if !ok {
			return fmt.Errorf("%w for topic %s", process.ErrCannotCastTxResolver, identifierTx)
		}

		err = handler.mainMessenger.RegisterMessageProcessor(p2p.TransactionsNetwork, resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
		if err != nil {
			return err
		}

		err = handler.fullArchiveMessenger.RegisterMessageProcessor(p2p.TransactionsNetwork, resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler *epochChangeTopicsHandler) replaceTxInterceptorsOnTxNetwork(
	topic string,
) error {
	if topic == common.RewardsTransactionTopic {
		return handler.replaceRewardTxInterceptorsOnTxNetwork()
	}

	shardC := handler.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)

		err := replaceTxTopicAndAssignHandlerOnMessenger(
			identifierTx,
			true,
			handler.mainMessenger,
			handler.mainInterceptorsContainer,
		)
		if err != nil {
			return err
		}

		if !handler.isFullArchive {
			continue
		}

		err = replaceTxTopicAndAssignHandlerOnMessenger(
			identifierTx,
			true,
			handler.fullArchiveMessenger,
			handler.fullArchiveInterceptorsContainer,
		)
		if err != nil {
			return err
		}
	}

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	err := replaceTxTopicAndAssignHandlerOnMessenger(
		identifierTx,
		true,
		handler.mainMessenger,
		handler.mainInterceptorsContainer,
	)
	if err != nil {
		return err
	}

	if !handler.isFullArchive {
		return nil
	}

	return replaceTxTopicAndAssignHandlerOnMessenger(
		identifierTx,
		true,
		handler.fullArchiveMessenger,
		handler.fullArchiveInterceptorsContainer,
	)
}

func (handler *epochChangeTopicsHandler) registerRewardTxResolversOnTxNetwork() error {
	shardC := handler.shardCoordinator
	if shardC.SelfId() != core.MetachainShardId {
		return handler.replaceShardRewardTxResolversOnTxNetwork()
	}

	return handler.replaceMetaRewardTxResolversOnTxNetwork()
}

func (handler *epochChangeTopicsHandler) replaceShardRewardTxResolversOnTxNetwork() error {
	shardC := handler.shardCoordinator
	topic := common.RewardsTransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	oldResolver, err := handler.resolversContainer.Get(topic)
	if err != nil {
		return err
	}

	resolver, ok := oldResolver.(topicResolver)
	if !ok {
		return fmt.Errorf("%w for topic %s", process.ErrCannotCastTxResolver, topic)
	}

	err = handler.mainMessenger.RegisterMessageProcessor(p2p.TransactionsNetwork, resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return err
	}

	return handler.fullArchiveMessenger.RegisterMessageProcessor(p2p.TransactionsNetwork, resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
}

func (handler *epochChangeTopicsHandler) replaceMetaRewardTxResolversOnTxNetwork() error {
	shardC := handler.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := common.RewardsTransactionTopic + shardC.CommunicationIdentifier(idx)
		// same old resolver will be used, already unregistered from the messengers
		oldResolver, err := handler.resolversContainer.Get(identifierTx)
		if err != nil {
			return err
		}

		resolver, ok := oldResolver.(topicResolver)
		if !ok {
			return fmt.Errorf("%w for topic %s", process.ErrCannotCastTxResolver, identifierTx)
		}

		err = handler.mainMessenger.RegisterMessageProcessor(p2p.TransactionsNetwork, resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
		if err != nil {
			return err
		}

		err = handler.fullArchiveMessenger.RegisterMessageProcessor(p2p.TransactionsNetwork, resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler *epochChangeTopicsHandler) replaceRewardTxInterceptorsOnTxNetwork() error {
	shardC := handler.shardCoordinator
	if shardC.SelfId() != core.MetachainShardId {
		return handler.replaceShardRewardTxInterceptorsOnTxNetwork()
	}

	return handler.replaceMetaRewardTxInterceptorsOnTxNetwork()
}

func (handler *epochChangeTopicsHandler) replaceShardRewardTxInterceptorsOnTxNetwork() error {
	shardC := handler.shardCoordinator
	topic := common.RewardsTransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	err := replaceTxTopicAndAssignHandlerOnMessenger(
		topic,
		true,
		handler.mainMessenger,
		handler.mainInterceptorsContainer,
	)
	if err != nil {
		return err
	}

	return replaceTxTopicAndAssignHandlerOnMessenger(
		topic,
		true,
		handler.fullArchiveMessenger,
		handler.fullArchiveInterceptorsContainer,
	)
}

func (handler *epochChangeTopicsHandler) replaceMetaRewardTxInterceptorsOnTxNetwork() error {
	shardC := handler.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	for idx := uint32(0); idx < noOfShards; idx++ {
		topic := common.RewardsTransactionTopic + shardC.CommunicationIdentifier(idx)
		err := replaceTxTopicAndAssignHandlerOnMessenger(
			topic,
			true,
			handler.mainMessenger,
			handler.mainInterceptorsContainer,
		)
		if err != nil {
			return err
		}

		err = replaceTxTopicAndAssignHandlerOnMessenger(
			topic,
			true,
			handler.fullArchiveMessenger,
			handler.fullArchiveInterceptorsContainer,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func replaceTxTopicAndAssignHandlerOnMessenger(
	topic string,
	createChannel bool,
	messenger p2p.Messenger,
	container process.InterceptorsContainer,
) error {
	err := removeTopicFromMessenger(messenger, topic)
	if err != nil {
		return err
	}

	err = messenger.CreateTopic(p2p.TransactionsNetwork, topic, createChannel)
	if err != nil {
		return err
	}

	interceptor, err := container.Get(topic)
	if err != nil {
		return err
	}

	return messenger.RegisterMessageProcessor(p2p.TransactionsNetwork, topic, common.DefaultInterceptorsIdentifier, interceptor)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *epochChangeTopicsHandler) IsInterfaceNil() bool {
	return handler == nil
}
