package sync

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
)

type requestHandlerWithSetEpochStub struct {
	testscommon.RequestHandlerStub
	SetEpochCalled func(epoch uint32)
}

func (rhs *requestHandlerWithSetEpochStub) SetEpoch(epoch uint32) {
	if rhs.SetEpochCalled != nil {
		rhs.SetEpochCalled(epoch)
		return
	}

	rhs.RequestHandlerStub.SetEpoch(epoch)
}

type blockBootstrapperStub struct {
	getCurrHeaderCalled        func() (data.HeaderHandler, error)
	getPrevHeaderCalled        func(data.HeaderHandler, storage.Storer) (data.HeaderHandler, error)
	getBlockBodyCalled         func(data.HeaderHandler) (data.BodyHandler, error)
	isForkTriggeredByMetaFunc  func() bool
	requestHeaderByNonceCalled func(nonce uint64)
	requestProofByNonceCalled  func(nonce uint64)
}

func (bbs *blockBootstrapperStub) getCurrHeader() (data.HeaderHandler, error) {
	if bbs.getCurrHeaderCalled != nil {
		return bbs.getCurrHeaderCalled()
	}
	return nil, nil
}

func (bbs *blockBootstrapperStub) getPrevHeader(header data.HeaderHandler, storer storage.Storer) (data.HeaderHandler, error) {
	if bbs.getPrevHeaderCalled != nil {
		return bbs.getPrevHeaderCalled(header, storer)
	}
	return nil, nil
}

func (bbs *blockBootstrapperStub) getBlockBody(header data.HeaderHandler) (data.BodyHandler, error) {
	if bbs.getBlockBodyCalled != nil {
		return bbs.getBlockBodyCalled(header)
	}
	return nil, nil
}

func (bbs *blockBootstrapperStub) getHeaderWithHashRequestingIfMissing([]byte) (data.HeaderHandler, error) {
	return nil, nil
}

func (bbs *blockBootstrapperStub) getHeaderWithNonceRequestingIfMissing(uint64) (data.HeaderHandler, []byte, error) {
	return nil, nil, nil
}

func (bbs *blockBootstrapperStub) getBlockBodyRequestingIfMissing(data.HeaderHandler) (data.BodyHandler, error) {
	return nil, nil
}

func (bbs *blockBootstrapperStub) isForkTriggeredByMeta() bool {
	if bbs.isForkTriggeredByMetaFunc != nil {
		return bbs.isForkTriggeredByMetaFunc()
	}
	return false
}

func (bbs *blockBootstrapperStub) requestHeaderByNonce(nonce uint64) {
	if bbs.requestHeaderByNonceCalled != nil {
		bbs.requestHeaderByNonceCalled(nonce)
	}
}

func (bbs *blockBootstrapperStub) requestProofByNonce(nonce uint64) {
	if bbs.requestProofByNonceCalled != nil {
		bbs.requestProofByNonceCalled(nonce)
	}
}

func getMockChainHandler() data.ChainHandler {
	return &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch: 0,
			}
		},
	}
}
