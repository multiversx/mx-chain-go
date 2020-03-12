package block

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ValidatorInfoProcessor implements validator info processing for miniblocks of type peerMiniblock
type ValidatorInfoProcessor struct {
	state                        *startOfEpochPeerMiniblocks
	miniBlocksPool               storage.Cacher
	marshalizer                  marshal.Marshalizer
	Hasher                       hashing.Hasher
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor
	requestHandler               process.RequestHandler
}

// NewValidatorInfoProcessor creates a new ValidatorInfoProcessor object
func NewValidatorInfoProcessor(arguments ArgValidatorInfoProcessor) (*ValidatorInfoProcessor, error) {
	if check.IfNil(arguments.ValidatorStatisticsProcessor) {
		return nil, process.ErrNilValidatorStatistics
	}
	if check.IfNil(arguments.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(arguments.MiniBlocksPool) {
		return nil, process.ErrNilMiniBlockPool
	}
	if check.IfNil(arguments.Requesthandler) {
		return nil, process.ErrNilRequestHandler
	}

	vip := &ValidatorInfoProcessor{
		state:                        &startOfEpochPeerMiniblocks{},
		miniBlocksPool:               arguments.MiniBlocksPool,
		marshalizer:                  arguments.Marshalizer,
		validatorStatisticsProcessor: arguments.ValidatorStatisticsProcessor,
		requestHandler:               arguments.Requesthandler,
		Hasher:                       arguments.Hasher,
	}

	vip.miniBlocksPool.RegisterHandler(vip.receivedMiniBlock)

	return vip, nil
}
func (vip *ValidatorInfoProcessor) TryProcessMetaBlock(metaBlock *block.MetaBlock, metablockHash []byte) error {
	vip.state.init(metaBlock, metablockHash)

	vip.computeMissingPeerBlocks(metaBlock)

	err := vip.retrieveMissingBlocks()
	if err != nil {
		return err
	}
	err = vip.tryProcessAllPeerMiniBlocks(metaBlock)
	if err != nil {
		return err
	}
	return nil
}

func (vip *ValidatorInfoProcessor) receivedMiniBlock(key []byte) {
	log.Debug("miniblock key", "key", key)
	mb, ok := vip.miniBlocksPool.Get(key)
	if ok {
		peerMb, ok := mb.(*block.MiniBlock)
		if ok && peerMb.Type == block.PeerBlock {
			log.Trace(fmt.Sprintf("received miniblock of type %s", peerMb.Type))

			vip.state.mutMiniBlocksForBlock.Lock()
			defer vip.state.mutMiniBlocksForBlock.Unlock()

			mb, exists := vip.miniBlocksPool.Peek(key)
			if exists {
				peerMb, ok := mb.(*block.MiniBlock)
				if ok {
					marshalledMiniblock, err := vip.marshalizer.Marshal(peerMb)
					if err != nil {
						log.Debug("receivedMiniBlock", "error", err.Error())
					} else {
						hash := vip.Hasher.Compute(string(marshalledMiniblock))
						if bytes.Equal(hash, key) {
							log.Debug("hash and key do not match")
						} else {
							vip.state.allPeerMiniblocks[string(key)] = peerMb
						}
					}
				}
			}

			for _, peerMb := range vip.state.allPeerMiniblocks {
				if peerMb == nil {
					return
				}
			}

			select {
			case vip.state.chRcvAllMiniblocks <- struct{}{}:
			default:
			}

		}
	}
}

func (vip *ValidatorInfoProcessor) tryProcessAllPeerMiniBlocks(metaBlock *block.MetaBlock) error {
	for _, peerMiniBlock := range metaBlock.MiniBlockHeaders {
		if peerMiniBlock.Type != block.PeerBlock {
			continue
		}

		mb, found := vip.state.allPeerMiniblocks[string(peerMiniBlock.Hash)]
		if !found {
			return process.ErrNilMiniBlock
		}

		for _, txHash := range mb.TxHashes {
			vid := &state.ValidatorInfo{}
			err := vip.marshalizer.Unmarshal(vid, txHash)
			if err != nil {
				return err
			}
			err = vip.validatorStatisticsProcessor.Process(vid)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vip *ValidatorInfoProcessor) computeMissingPeerBlocks(metaBlock *block.MetaBlock) {
	log.Trace("got a startEpochBlock", "miniblockHeaders len", len(metaBlock.MiniBlockHeaders))
	peerMiniBlocks := make(map[string]*block.MiniBlock)
	missingNumber := 0
	for _, mb := range metaBlock.MiniBlockHeaders {
		if mb.Type == block.PeerBlock {
			mbObjectFound, ok := vip.miniBlocksPool.Peek(mb.Hash)
			if ok {
				mbFound, isMB := mbObjectFound.(*block.MiniBlock)
				if isMB {
					peerMiniBlocks[string(mb.Hash)] = mbFound
					continue
				}
			}
			peerMiniBlocks[string(mb.Hash)] = nil
			missingNumber++
		}
	}

	log.Trace("missing allPeerMiniblocks", "len", len(peerMiniBlocks))

	vip.state.setPeerMiniblocks(peerMiniBlocks)

}

func (vip *ValidatorInfoProcessor) retrieveMissingBlocks() error {
	vip.state.mutMiniBlocksForBlock.Lock()
	currentState := vip.state
	vip.state.mutMiniBlocksForBlock.Unlock()
	missingMiniblocks := make([][]byte, 0)
	for mbHash, mb := range currentState.allPeerMiniblocks {
		if mb == nil {
			missingMiniblocks = append(missingMiniblocks, []byte(mbHash))
		}
	}
	if len(missingMiniblocks) > 0 {
		go vip.requestHandler.RequestMiniBlocks(core.MetachainShardId, missingMiniblocks)
	}

	if len(missingMiniblocks) > 0 {
		select {
		case <-currentState.chRcvAllMiniblocks:
			return nil
		case <-time.After(time.Second * 5):
			return process.ErrTimeIsOut
		}
	}

	return nil
}

func (vip *ValidatorInfoProcessor) IsInterfaceNil() bool {
	return vip == nil
}
