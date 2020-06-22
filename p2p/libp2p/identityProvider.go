package libp2p

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/data"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

const authProtocolID = "elrond-node-auth/0.0.1"
const maxBytesToReceive = 1024

type identityProvider struct {
	host                     host.Host
	networkShardingCollector p2p.NetworkShardingCollector
	signerVerifier           p2p.SignerVerifier
	marshalizer              p2p.Marshalizer
	receiveTimeout           time.Duration
}

// NewIdentityProvider creates a wrapper over libp2p's network.Notifiee implementation
// that is able to do a supplemental authentication against each peer connected
func NewIdentityProvider(
	host host.Host,
	networkShardingCollector p2p.NetworkShardingCollector,
	signerVerifier p2p.SignerVerifier,
	marshalizer p2p.Marshalizer,
	receiveTimeout time.Duration,
) (*identityProvider, error) {

	if host == nil {
		return nil, p2p.ErrNilHost
	}
	if check.IfNil(networkShardingCollector) {
		return nil, p2p.ErrNilNetworkShardingCollector
	}
	if check.IfNil(signerVerifier) {
		return nil, p2p.ErrNilSignerVerifier
	}
	if check.IfNil(marshalizer) {
		return nil, p2p.ErrNilMarshalizer
	}

	ip := &identityProvider{
		host:                     host,
		networkShardingCollector: networkShardingCollector,
		signerVerifier:           signerVerifier,
		marshalizer:              marshalizer,
		receiveTimeout:           receiveTimeout,
	}
	ip.host.SetStreamHandler(authProtocolID, ip.handleStreams)

	return ip, nil
}

// Listen is called when network starts listening on an addr
func (ip *identityProvider) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (ip *identityProvider) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (ip *identityProvider) Connected(_ network.Network, conn network.Conn) {
	go func() {
		ctx := context.Background()
		s, err := ip.host.NewStream(ctx, conn.RemotePeer(), authProtocolID)
		if err != nil {
			log.Debug("identity provider handleStreams", "error", err.Error())
			return
		}

		buff, err := ip.createPayload()
		if err != nil {
			log.Warn("identity provider can not create payload", "error", err.Error())
			return
		}

		_, err = s.Write(buff)
		if err != nil {
			log.Debug("identity provider write", "error", err.Error())
			return
		}

		log.Trace("message sent", "payload hex", hex.EncodeToString(buff))
	}()
}

// Disconnected is called when a connection closed
func (ip *identityProvider) Disconnected(network.Network, network.Conn) {}

// OpenedStream is called when a stream opened
func (ip *identityProvider) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (ip *identityProvider) ClosedStream(network.Network, network.Stream) {}

func (ip *identityProvider) createPayload() ([]byte, error) {
	am := data.AuthMessage{}
	am.Message = []byte(ip.host.ID())
	am.Timestamp = time.Now().Unix()
	am.Pubkey = ip.signerVerifier.PublicKey()

	buffWithoutSig, err := ip.marshalizer.Marshal(&am)
	if err != nil {
		return nil, err
	}

	sig, err := ip.signerVerifier.Sign(buffWithoutSig)
	if err != nil {
		return nil, err
	}

	am.Sig = sig

	payload, err := ip.marshalizer.Marshal(&am)
	if err != nil {
		return nil, err
	}

	return payload, err
}

func (ip *identityProvider) handleStreams(s network.Stream) {
	chError := make(chan error)
	chData := make(chan []byte)

	go func() {
		buff := make([]byte, maxBytesToReceive)
		n, err := s.Read(buff)
		if err != nil {
			chError <- err
			return
		}

		recvBuff := buff[:n]
		chData <- recvBuff
	}()

	select {
	case recvBuff := <-chData:
		log.Trace("message received", "payload hex", hex.EncodeToString(recvBuff))
		err := ip.processReceivedData(recvBuff)
		if err != nil {
			log.Debug("identity provider processReceivedData", "error", err)
		}
	case err := <-chError:
		log.Debug("identity provider read", "error", err.Error())
		_ = s.Close()
	case <-time.After(ip.receiveTimeout):
		log.Debug("identity provider read", "error", "timeout")
	}
}

func (ip *identityProvider) processReceivedData(recvBuff []byte) error {
	receivedAm := &data.AuthMessage{}
	err := ip.marshalizer.Unmarshal(receivedAm, recvBuff)
	if err != nil {
		return err
	}

	copiedAm := *receivedAm
	copiedAm.Sig = nil
	bufWithoutSig, err := ip.marshalizer.Marshal(&copiedAm)
	if err != nil {
		return err
	}

	err = ip.signerVerifier.Verify(bufWithoutSig, receivedAm.Sig, receivedAm.Pubkey)
	if err != nil {
		return err
	}

	pid := core.PeerID(receivedAm.Message)
	ip.networkShardingCollector.UpdatePeerIdPublicKey(pid, receivedAm.Pubkey)
	log.Trace("received authentication", "from", pid.Pretty(), "pk", hex.EncodeToString(receivedAm.Pubkey))

	return nil
}
