package factory

import (
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

func CreateShardCoordinator(
	nodesConfig NodesSetupHandler,
	pubKey crypto.PublicKey,
	prefsConfig config.PreferencesConfig,
	log logger.Logger,
) (sharding.Coordinator, core.NodeType, error) {

	selfShardId, err := getShardIdFromNodePubKey(pubKey, nodesConfig)
	nodeType := core.NodeTypeValidator
	if err == sharding.ErrPublicKeyNotFoundInGenesis {
		nodeType = core.NodeTypeObserver
		log.Info("starting as observer node")

		selfShardId, err = core.ProcessDestinationShardAsObserver(prefsConfig.DestinationShardAsObserver)
		if err != nil {
			return nil, "", err
		}
		if selfShardId == core.DisabledShardIDAsObserver {
			selfShardId = uint32(0)
		}
	}
	if err != nil {
		return nil, "", err
	}

	var shardName string
	if selfShardId == core.MetachainShardId {
		shardName = core.MetachainShardName
	} else {
		shardName = fmt.Sprintf("%d", selfShardId)
	}
	log.Info("shard info", "started in shard", shardName)

	shardCoordinator, err := sharding.NewMultiShardCoordinator(nodesConfig.NumberOfShards(), selfShardId)
	if err != nil {
		return nil, "", err
	}

	return shardCoordinator, nodeType, nil
}

func getShardIdFromNodePubKey(pubKey crypto.PublicKey, nodesConfig NodesSetupHandler) (uint32, error) {
	if pubKey == nil {
		return 0, errors.New("nil public key")
	}

	publicKey, err := pubKey.ToByteArray()
	if err != nil {
		return 0, err
	}

	selfShardId, err := nodesConfig.GetShardIDForPubKey(publicKey)
	if err != nil {
		return 0, err
	}

	return selfShardId, err
}
