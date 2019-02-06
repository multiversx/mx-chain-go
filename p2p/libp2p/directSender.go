package libp2p

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/whyrusleeping/timecache"
)

const timeSeenMessages = time.Second * 120

type directSender struct {
	counter            uint64
	ctx                context.Context
	hostP2P            host.Host
	receivedDirectChan chan *pubsub_pb.Message
	messageHandler     func(msg p2p.MessageP2P) error
	seenMessages       *timecache.TimeCache
}

// NewDirectSender returns a new instance of direct sender object
func NewDirectSender(
	ctx context.Context,
	h host.Host,
	messageHandler func(msg p2p.MessageP2P) error,
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

	ds := &directSender{
		counter:            uint64(time.Now().UnixNano()),
		ctx:                ctx,
		hostP2P:            h,
		receivedDirectChan: make(chan *pubsub_pb.Message),
		seenMessages:       timecache.NewTimeCache(timeSeenMessages),
		messageHandler:     messageHandler,
	}

	//wire-up a handler for direct messages
	h.SetStreamHandler(DirectSendID, ds.directStreamHandler)

	go ds.processReceivedDirect()

	return ds, nil
}

func (ds *directSender) directStreamHandler(s net.Stream) {
	reader := ggio.NewDelimitedReader(s, 1<<20)

	go func(r ggio.ReadCloser) {
		for {
			msg := &pubsub_pb.Message{}

			err := reader.ReadMsg(msg)
			if err != nil {
				//stream has encountered an error, close this go routine

				if err != io.EOF {
					_ = s.Reset()
					log.Debug(fmt.Sprintf("error reading rpc from %s: %s", s.Conn().RemotePeer(), err))
				} else {
					// Just be nice. They probably won't read this
					// but it doesn't hurt to send it.
					_ = s.Close()
				}
				return
			}

			ds.receivedDirectChan <- msg
		}
	}(reader)
}

func (ds *directSender) processReceivedDirect() {
	for {
		select {
		case msg := <-ds.receivedDirectChan:
			err := ds.processReceivedDirectMessage(msg)
			if err != nil {
				log.Debug(err.Error())
			}
		}
	}
}

func (ds *directSender) processReceivedDirectMessage(message *pubsub_pb.Message) error {
	if message == nil {
		return p2p.ErrNilMessage
	}

	if message.TopicIDs == nil {
		return p2p.ErrNilTopic
	}

	if len(message.TopicIDs) == 0 {
		return p2p.ErrEmptyTopicList
	}

	if ds.checkAndSetSeenMessage(message) {
		return p2p.ErrAlreadySeenMessage
	}

	p2pMsg := p2p.NewMessage(&pubsub.Message{Message: message})
	return ds.messageHandler(p2pMsg)
}

func (ds *directSender) checkAndSetSeenMessage(msg *pubsub_pb.Message) bool {
	msgId := string(msg.GetFrom()) + string(msg.GetSeqno())

	if ds.seenMessages.Has(msgId) {
		return true
	}

	ds.seenMessages.Add(msgId)
	return false
}

// NextSeqno returns the next uint64 found in *counter as byte slice
func (ds *directSender) NextSeqno(counter *uint64) []byte {
	seqno := make([]byte, 8)
	newVal := atomic.AddUint64(counter, 1)
	binary.BigEndian.PutUint64(seqno, newVal)
	return seqno
}

// SendDirectToConnectedPeer will send a direct message to the connected peer
func (ds *directSender) SendDirectToConnectedPeer(topic string, buff []byte, peer p2p.PeerID) error {
	conn, err := ds.getConnection(peer)
	if err != nil {
		return err
	}

	stream, err := ds.getOrCreateStream(conn)
	if err != nil {
		return err
	}

	msg := ds.createMessage(topic, buff, conn)

	bufw := bufio.NewWriter(stream)
	w := ggio.NewDelimitedWriter(bufw)

	go func(msg proto.Message) {
		err := w.WriteMsg(msg)
		log.LogIfError(err)

		err = bufw.Flush()
		log.LogIfError(err)
	}(msg)

	return nil
}

func (ds *directSender) getConnection(p p2p.PeerID) (net.Conn, error) {
	conns := ds.hostP2P.Network().ConnsToPeer(peer.ID(p))

	if len(conns) == 0 {
		return nil, p2p.ErrPeerNotDirectlyConnected
	}

	return conns[0], nil
}

func (ds *directSender) getOrCreateStream(conn net.Conn) (net.Stream, error) {
	//find or create stream
	streams := conn.GetStreams()
	var foundStream net.Stream
	for i := 0; i < len(streams); i++ {
		if streams[i].Protocol() == DirectSendID {
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

func (ds *directSender) createMessage(topic string, buff []byte, conn net.Conn) *pubsub_pb.Message {
	//create message
	seqno := ds.NextSeqno(&ds.counter)
	mes := pubsub_pb.Message{}
	mes.Data = buff
	mes.TopicIDs = []string{topic}
	mes.From = []byte(conn.LocalPeer())
	mes.Seqno = seqno

	return &mes
}
