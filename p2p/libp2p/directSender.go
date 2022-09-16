package libp2p

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	pubsub "github.com/ElrondNetwork/go-libp2p-pubsub"
	pubsubPb "github.com/ElrondNetwork/go-libp2p-pubsub/pb"
	ggio "github.com/gogo/protobuf/io"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/whyrusleeping/timecache"
)

var _ p2p.DirectSender = (*directSender)(nil)

const timeSeenMessages = time.Second * 120
const maxMutexes = 10000
const sequenceNumberSize = 8

type directSender struct {
	counter         uint64
	ctx             context.Context
	hostP2P         host.Host
	messageHandler  func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error
	mutSeenMessages sync.Mutex
	seenMessages    *timecache.TimeCache
	mutexForPeer    *MutexHolder
	// signer          p2p.SignerVerifier
}

// NewDirectSender returns a new instance of direct sender object
func NewDirectSender(
	ctx context.Context,
	h host.Host,
	messageHandler func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error,
	signer p2p.SignerVerifier,
) (*directSender, error) {

	if h == nil {
		return nil, p2p.ErrNilHost
	}
	if ctx == nil {
		return nil, p2p.ErrNilContext
	}
	if messageHandler == nil {
		return nil, p2p.ErrNilDirectSendMessageHandler
	}
	if check.IfNil(signer) {
		return nil, p2p.ErrNilP2PSigner
	}

	mutexForPeer, err := NewMutexHolder(maxMutexes)
	if err != nil {
		return nil, err
	}

	ds := &directSender{
		counter:        uint64(time.Now().UnixNano()),
		ctx:            ctx,
		hostP2P:        h,
		seenMessages:   timecache.NewTimeCache(timeSeenMessages),
		messageHandler: messageHandler,
		mutexForPeer:   mutexForPeer,
		// 	signer:         signer,
	}

	// wire-up a handler for direct messages
	h.SetStreamHandler(DirectSendID, ds.directStreamHandler)

	return ds, nil
}

func (ds *directSender) directStreamHandler(s network.Stream) {
	reader := ggio.NewDelimitedReader(s, maxSendBuffSize)

	go func(r ggio.ReadCloser) {
		for {
			msg := &pubsubPb.Message{}

			err := reader.ReadMsg(msg)
			if err != nil {
				// stream has encountered an error, close this go routine

				if err != io.EOF {
					_ = s.Reset()
					log.Trace("error reading rpc",
						"from", s.Conn().RemotePeer(),
						"error", err.Error(),
					)
				} else {
					// Just be nice. They probably won't read this
					// but it doesn't hurt to send it.
					_ = s.Close()
				}
				return
			}

			err = ds.processReceivedDirectMessage(msg, s.Conn().RemotePeer())
			if err != nil {
				log.Trace("p2p processReceivedDirectMessage", "error", err.Error())
			}
		}
	}(reader)
}

func (ds *directSender) processReceivedDirectMessage(message *pubsubPb.Message, fromConnectedPeer peer.ID) error {
	if message == nil {
		return p2p.ErrNilMessage
	}
	if message.Topic == nil {
		return p2p.ErrNilTopic
	}
	if !bytes.Equal(message.GetFrom(), []byte(fromConnectedPeer)) {
		return fmt.Errorf("%w mismatch between From and fromConnectedPeer values", p2p.ErrInvalidValue)
	}
	if message.Key != nil {
		return fmt.Errorf("%w for Key field as the node accepts only nil on this field", p2p.ErrInvalidValue)
	}
	if len(message.Seqno) > sequenceNumberSize {
		return fmt.Errorf("%w for SeqNo field as the node accepts only a maximum %d bytes", p2p.ErrInvalidValue, sequenceNumberSize)
	}
	if ds.checkAndSetSeenMessage(message) {
		return p2p.ErrAlreadySeenMessage
	}
	err := ds.checkSig(message)
	if err != nil {
		return err
	}

	pbMessage := &pubsub.Message{
		Message: message,
	}

	return ds.messageHandler(pbMessage, core.PeerID(fromConnectedPeer))
}

func (ds *directSender) checkAndSetSeenMessage(msg *pubsubPb.Message) bool {
	msgId := string(msg.GetFrom()) + string(msg.GetSeqno())

	ds.mutSeenMessages.Lock()
	defer ds.mutSeenMessages.Unlock()

	if ds.seenMessages.Has(msgId) {
		return true
	}

	ds.seenMessages.Add(msgId)
	return false
}

// NextSequenceNumber returns the next uint64 found in *counter as byte slice
func (ds *directSender) NextSequenceNumber() []byte {
	seqno := make([]byte, sequenceNumberSize)
	newVal := atomic.AddUint64(&ds.counter, 1)
	binary.BigEndian.PutUint64(seqno, newVal)
	return seqno
}

// Send will send a direct message to the connected peer
func (ds *directSender) Send(topic string, buff []byte, peer core.PeerID) error {
	if len(buff) >= maxSendBuffSize {
		return fmt.Errorf("%w, to be sent: %d, maximum: %d", p2p.ErrMessageTooLarge, len(buff), maxSendBuffSize)
	}

	mut := ds.mutexForPeer.Get(string(peer))
	mut.Lock()
	defer mut.Unlock()

	conn, err := ds.getConnection(peer)
	if err != nil {
		return err
	}

	stream, err := ds.getOrCreateStream(conn)
	if err != nil {
		return err
	}

	msg, err := ds.createMessage(topic, buff, conn)
	if err != nil {
		return err
	}

	bufw := bufio.NewWriter(stream)
	w := ggio.NewDelimitedWriter(bufw)

	err = w.WriteMsg(msg)
	if err != nil {
		_ = stream.Reset()
		_ = stream.Close()
		return err
	}

	err = bufw.Flush()
	if err != nil {
		_ = stream.Reset()
		_ = stream.Close()
		return err
	}

	return nil
}

func (ds *directSender) getConnection(p core.PeerID) (network.Conn, error) {
	conns := ds.hostP2P.Network().ConnsToPeer(peer.ID(p))
	if len(conns) == 0 {
		return nil, p2p.ErrPeerNotDirectlyConnected
	}

	// return the connection that has the highest number of streams
	lStreams := 0
	var conn network.Conn
	for _, c := range conns {
		length := len(c.GetStreams())
		if length >= lStreams {
			lStreams = length
			conn = c
		}
	}

	return conn, nil
}

func (ds *directSender) getOrCreateStream(conn network.Conn) (network.Stream, error) {
	streams := conn.GetStreams()
	var foundStream network.Stream
	for i := 0; i < len(streams); i++ {
		isExpectedStream := streams[i].Protocol() == DirectSendID
		isSendableStream := streams[i].Stat().Direction == network.DirOutbound

		if isExpectedStream && isSendableStream {
			foundStream = streams[i]
			break
		}
	}

	var err error

	if foundStream == nil {
		foundStream, err = ds.hostP2P.NewStream(ds.ctx, conn.RemotePeer(), DirectSendID)
		if err != nil {
			return nil, err
		}
	}

	return foundStream, nil
}

func (ds *directSender) createMessage(topic string, buff []byte, conn network.Conn) (*pubsubPb.Message, error) {
	seqno := ds.NextSequenceNumber()
	mes := pubsubPb.Message{}
	mes.Data = buff
	mes.Topic = &topic
	mes.From = []byte(conn.LocalPeer())
	mes.Seqno = seqno
	mes.Key = nil

	//buff, err := mes.Marshal()
	//if err != nil {
	//	return nil, err
	//}
	//
	//buff = withSignPrefix(buff)
	//
	//mes.Signature, err = ds.signer.Sign(buff)
	//if err != nil {
	//	return nil, err
	//}

	return &mes, nil
}

func (ds *directSender) checkSig(message *pubsubPb.Message) error {
	if len(message.Signature) == 0 {
		return nil // TODO will remove this in the future
	}

	//copyMessage := *message
	//copyMessage.Signature = nil
	//copyMessage.Key = nil
	//
	//buff, err := copyMessage.Marshal()
	//if err != nil {
	//	return err
	//}
	//
	//buff = withSignPrefix(buff)
	//
	//return ds.signer.Verify(buff, core.PeerID(message.From), message.Signature)

	return nil
}

func withSignPrefix(bytes []byte) []byte {
	return append([]byte(pubsub.SignPrefix), bytes...)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ds *directSender) IsInterfaceNil() bool {
	return ds == nil
}
