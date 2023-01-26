package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor/chunk"
	"github.com/multiversx/mx-chain-go/storage"
)

const minimumRequestTimeInterval = time.Millisecond * 200

type chunkHandler interface {
	Put(chunkIndex uint32, buff []byte)
	TryAssembleAllChunks() []byte
	GetAllMissingChunkIndexes() []uint32
	Size() int
	IsInterfaceNil() bool
}

type checkRequest struct {
	batch        *batch.Batch
	chanResponse chan process.CheckedChunkResult
}

// TrieNodesChunksProcessorArgs is the argument DTO used in the trieNodeChunksProcessor constructor
type TrieNodesChunksProcessorArgs struct {
	Hasher          hashing.Hasher
	ChunksCacher    storage.Cacher
	RequestInterval time.Duration
	RequestHandler  process.RequestHandler
	Topic           string
}

type trieNodeChunksProcessor struct {
	hasher            hashing.Hasher
	chunksCacher      storage.Cacher
	chanCheckRequests chan checkRequest
	requestInterval   time.Duration
	requestHandler    process.RequestHandler
	topic             string
	cancel            func()
	chanClose         chan struct{}
}

// NewTrieNodeChunksProcessor creates a new trieNodeChunksProcessor instance
func NewTrieNodeChunksProcessor(arg TrieNodesChunksProcessorArgs) (*trieNodeChunksProcessor, error) {
	if check.IfNil(arg.Hasher) {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor", process.ErrNilHasher)
	}
	if check.IfNil(arg.ChunksCacher) {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor", process.ErrNilCacher)
	}
	if arg.RequestInterval < minimumRequestTimeInterval {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor, minimum request interval is %v",
			process.ErrInvalidValue, minimumRequestTimeInterval)
	}
	if check.IfNil(arg.RequestHandler) {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor", process.ErrNilRequestHandler)
	}
	if len(arg.Topic) == 0 {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor", process.ErrEmptyTopic)
	}

	tncp := &trieNodeChunksProcessor{
		hasher:            arg.Hasher,
		chunksCacher:      arg.ChunksCacher,
		chanCheckRequests: make(chan checkRequest),
		requestInterval:   arg.RequestInterval,
		requestHandler:    arg.RequestHandler,
		topic:             arg.Topic,
		chanClose:         make(chan struct{}),
	}
	var ctx context.Context
	ctx, tncp.cancel = context.WithCancel(context.Background())
	go tncp.processLoop(ctx)

	log.Debug("NewTrieNodeChunksProcessor", "key", arg.Topic)

	return tncp, nil
}

func (proc *trieNodeChunksProcessor) processLoop(ctx context.Context) {
	chanDoRequests := time.After(proc.requestInterval)
	for {
		select {
		case <-ctx.Done():
			log.Debug("trieNodeChunksProcessor.processLoop go routine is stopping...")
			return
		case request := <-proc.chanCheckRequests:
			proc.processCheckRequest(request)
		case <-chanDoRequests:
			proc.doRequests(ctx)
			chanDoRequests = time.After(proc.requestInterval)
		}
	}
}

// CheckBatch will check the batch returning a checked chunk result containing result processing
func (proc *trieNodeChunksProcessor) CheckBatch(b *batch.Batch, whiteListHandler process.WhiteListHandler) (process.CheckedChunkResult, error) {
	batchValid, err := proc.batchIsValid(b, whiteListHandler)
	if !batchValid {
		return process.CheckedChunkResult{
			IsChunk:        false,
			HaveAllChunks:  false,
			CompleteBuffer: nil,
		}, err
	}

	respChan := make(chan process.CheckedChunkResult, 1)
	req := checkRequest{
		batch:        b,
		chanResponse: respChan,
	}

	select {
	case proc.chanCheckRequests <- req:
	case <-proc.chanClose:
		return process.CheckedChunkResult{}, process.ErrProcessClosed
	}

	select {
	case response := <-respChan:
		return response, nil
	case <-proc.chanClose:
		return process.CheckedChunkResult{}, process.ErrProcessClosed
	}
}

func (proc *trieNodeChunksProcessor) processCheckRequest(cr checkRequest) {
	shouldNotCreateChunk := cr.batch.ChunkIndex != 0
	result := process.CheckedChunkResult{
		IsChunk: true,
	}
	chunkObject, found := proc.chunksCacher.Get(cr.batch.Reference)
	if !found {
		if shouldNotCreateChunk {
			//we received other chunks from a previous, completed large trie node, return

			proc.writeCheckedChunkResultOnChan(cr, result)
			return
		}

		chunkObject = chunk.NewChunk(cr.batch.MaxChunks, cr.batch.Reference)
	}
	chunkData, ok := chunkObject.(chunkHandler)
	if !ok {
		if shouldNotCreateChunk {
			//we received other chunks from a previous, completed large trie node, return
			proc.writeCheckedChunkResultOnChan(cr, result)
			return
		}

		chunkData = chunk.NewChunk(cr.batch.MaxChunks, cr.batch.Reference)
	}

	chunkData.Put(cr.batch.ChunkIndex, cr.batch.Data[0])

	result.CompleteBuffer = chunkData.TryAssembleAllChunks()
	result.HaveAllChunks = len(result.CompleteBuffer) > 0
	if result.HaveAllChunks {
		proc.chunksCacher.Remove(cr.batch.Reference)
	} else {
		proc.chunksCacher.Put(cr.batch.Reference, chunkData, chunkData.Size())
	}

	proc.writeCheckedChunkResultOnChan(cr, result)
}

func (proc *trieNodeChunksProcessor) writeCheckedChunkResultOnChan(cr checkRequest, result process.CheckedChunkResult) {
	select {
	case cr.chanResponse <- result:
	default:
		log.Trace("trieNodeChunksProcessor.processCheckRequest - no one is listening on the end chan")
	}
}

func (proc *trieNodeChunksProcessor) batchIsValid(b *batch.Batch, whiteListHandler process.WhiteListHandler) (bool, error) {
	if b.MaxChunks < 2 {
		return false, nil
	}
	if len(b.Reference) != proc.hasher.Size() {
		return false, process.ErrIncompatibleReference
	}
	if len(b.Data) != 1 {
		return false, nil
	}
	if check.IfNil(whiteListHandler) {
		return false, process.ErrNilWhiteListHandler
	}

	identifier := proc.requestHandler.CreateTrieNodeIdentifier(b.Reference, b.ChunkIndex)
	isWhiteListed := whiteListHandler.IsWhiteListedAtLeastOne([][]byte{identifier}) ||
		whiteListHandler.IsWhiteListedAtLeastOne([][]byte{b.Reference})

	if !isWhiteListed {
		return false, process.ErrTrieNodeIsNotWhitelisted
	}

	return true, nil
}

func (proc *trieNodeChunksProcessor) doRequests(ctx context.Context) {
	references := proc.chunksCacher.Keys()
	for _, ref := range references {
		select {
		case <-ctx.Done():
			//early exit
			return
		default:
		}

		proc.requestMissingForReference(ref, ctx)
	}
}

func (proc *trieNodeChunksProcessor) requestMissingForReference(reference []byte, ctx context.Context) {
	data, found := proc.chunksCacher.Get(reference)
	if !found {
		return
	}

	chunkData, ok := data.(chunkHandler)
	if !ok {
		return
	}

	missing := chunkData.GetAllMissingChunkIndexes()
	for _, missingChunkIndex := range missing {
		select {
		case <-ctx.Done():
			//early exit
			return
		default:
		}

		proc.requestHandler.RequestTrieNode(reference, proc.topic, missingChunkIndex)
	}
}

// Close will close the process go routine
func (proc *trieNodeChunksProcessor) Close() error {
	log.Debug("trieNodeChunkProcessor.Close()", "key", proc.topic)
	defer func() {
		//this instruction should be called last as to release hanging go routines
		close(proc.chanClose)
	}()

	proc.cancel()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (proc *trieNodeChunksProcessor) IsInterfaceNil() bool {
	return proc == nil
}
