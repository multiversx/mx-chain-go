package pruning

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	lastEpochIndex             = 1
	currentEpochIndex          = 0
	minNumOfActiveDBsNecessary = 2
)

type triePruningStorer struct {
	*PruningStorer
}

// NewTriePruningStorer will return a new instance of NewTriePruningStorer
func NewTriePruningStorer(args *StorerArgs) (*triePruningStorer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	activePersisters, persistersMapByEpoch, err := initTriePersisterInEpoch(args, "")
	if err != nil {
		return nil, err
	}

	ps, err := initPruningStorer(args, "", activePersisters, persistersMapByEpoch)
	if err != nil {
		return nil, err
	}

	tps := &triePruningStorer{ps}
	ps.extendPersisterLifeHandler = tps.extendPersisterLife
	tps.registerHandler(args.Notifier)

	return tps, nil
}

func (ps *triePruningStorer) extendPersisterLife() bool {
	numActiveDBs := 0
	for i := 0; i < int(ps.numOfActivePersisters); i++ {
		if i >= len(ps.activePersisters) {
			continue
		}

		val, err := ps.activePersisters[i].persister.Get([]byte(common.ActiveDBKey))
		if err != nil {
			continue
		}

		if bytes.Equal(val, []byte(common.ActiveDBVal)) {
			numActiveDBs++
		}
	}

	if numActiveDBs < minNumOfActiveDBsNecessary {
		log.Debug("extendPersisterLife", "path", ps.dbPath)
		return true
	}

	return false
}

func initTriePersisterInEpoch(
	args *StorerArgs,
	shardIDStr string,
) ([]*persisterData, map[uint32]*persisterData, error) {
	if !args.PruningEnabled {
		return createPersisterIfPruningDisabled(args, shardIDStr)
	}

	if args.NumOfEpochsToKeep < args.NumOfActivePersisters {
		return nil, nil, fmt.Errorf("invalid epochs configuration")
	}

	oldestEpochActive, oldestEpochKeep := computeOldestEpochActiveAndToKeep(args)
	var persisters []*persisterData
	persistersMapByEpoch := make(map[uint32]*persisterData)

	closeOldPersisters := false
	for epoch := int64(args.StartingEpoch); epoch >= oldestEpochKeep; epoch-- {
		log.Debug("initTriePersisterInEpoch(): createPersisterDataForEpoch", "identifier", args.Identifier, "epoch", epoch, "shardID", shardIDStr)
		p, err := createPersisterDataForEpoch(args, uint32(epoch), shardIDStr)
		if err != nil {
			return nil, nil, err
		}

		persistersMapByEpoch[uint32(epoch)] = p

		if epoch < oldestEpochActive && closeOldPersisters {
			err = p.Close()
			if err != nil {
				log.Debug("persister.Close()", "identifier", args.Identifier, "error", err.Error())
			}
		} else {
			persisters = append(persisters, p)
			log.Debug("appended a pruning active persister", "epoch", epoch, "identifier", args.Identifier)
		}

		if !closeOldPersisters {
			closeOldPersisters = shouldCloseOldPersisters(p)
		}
	}

	return persisters, persistersMapByEpoch, nil
}

func shouldCloseOldPersisters(pd *persisterData) bool {
	val, err := pd.persister.Get([]byte(common.ActiveDBKey))
	if err != nil {
		return false
	}

	if bytes.Equal(val, []byte(common.ActiveDBVal)) {
		return true
	}

	return false
}

// PutInEpochWithoutCache adds data to persistence medium related to the specified epoch
func (ps *triePruningStorer) PutInEpochWithoutCache(key []byte, data []byte, epoch uint32) error {
	ps.lock.RLock()
	pd, exists := ps.persistersMapByEpoch[epoch]
	ps.lock.RUnlock()
	if !exists {
		return fmt.Errorf("put in epoch: persister for epoch %d not found", epoch)
	}

	persister, closePersister, err := ps.createAndInitPersisterIfClosedProtected(pd)
	if err != nil {
		return err
	}
	defer closePersister()

	err = persister.Put(key, data)
	if err != nil {
		return err
	}

	return nil
}

// GetFromOldEpochsWithoutAddingToCache searches the old epochs for the given key without adding to the cache
func (ps *triePruningStorer) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, error) {
	v, ok := ps.cacher.Get(key)
	if ok && !bytes.Equal([]byte(common.ActiveDBKey), key) {
		return v.([]byte), nil
	}

	ps.lock.RLock()
	defer ps.lock.RUnlock()

	numClosedDbs := 0
	for idx := 1; idx < len(ps.activePersisters); idx++ {
		val, err := ps.activePersisters[idx].persister.Get(key)
		if err != nil {
			if err == storage.ErrDBIsClosed {
				numClosedDbs++
			}

			continue
		}

		return val, nil
	}

	if numClosedDbs+1 == len(ps.activePersisters) && len(ps.activePersisters) > 1 {
		return nil, storage.ErrDBIsClosed
	}

	return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
}

// GetFromLastEpoch searches only the last epoch storer for the given key
func (ps *triePruningStorer) GetFromLastEpoch(key []byte) ([]byte, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if len(ps.activePersisters) < 2 {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	return ps.activePersisters[lastEpochIndex].persister.Get(key)
}

// GetFromCurrentEpoch searches only the current epoch storer for the given key
func (ps *triePruningStorer) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	ps.lock.RLock()

	if len(ps.activePersisters) == 0 {
		ps.lock.RUnlock()
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	persister := ps.activePersisters[currentEpochIndex].persister
	ps.lock.RUnlock()

	return persister.Get(key)
}

// GetLatestStorageEpoch returns the epoch for the latest opened persister
func (ps *triePruningStorer) GetLatestStorageEpoch() (uint32, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if len(ps.activePersisters) == 0 {
		return 0, fmt.Errorf("there are no active persisters")
	}

	return ps.activePersisters[currentEpochIndex].epoch, nil
}
