package shell

import (
	"encoding/binary"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
)

// Message is a pubsub message.
type Message struct {
	From     peer.ID
	Data     []byte
	Seqno    []byte
	TopicIDs []string
}

type message struct {
	*pb.Message
}

func (m *message) GetFrom() peer.ID {
	return peer.ID(m.Message.GetFrom())
}

type floodsubRecord struct {
	msg *message
}

func (r floodsubRecord) From() peer.ID {
	return r.msg.GetFrom()
}

func (r floodsubRecord) Data() []byte {
	return r.msg.GetData()
}

func (r floodsubRecord) SeqNo() int64 {
	return int64(binary.BigEndian.Uint64(r.msg.GetSeqno()))
}

func (r floodsubRecord) TopicIDs() []string {
	return r.msg.GetTopicIDs()
}

///

// PubSubSubscription allow you to receive pubsub records that where published on the network.
type PubSubSubscription struct {
	resp *Response
}

func newPubSubSubscription(resp *Response) *PubSubSubscription {
	sub := &PubSubSubscription{
		resp: resp,
	}

	return sub
}

// Next waits for the next record and returns that.
func (s *PubSubSubscription) Next() (PubSubRecord, error) {
	if s.resp.Error != nil {
		return nil, s.resp.Error
	}

	d := json.NewDecoder(s.resp.Output)

	var r struct {
		From     []byte   `json:"from,omitempty"`
		Data     []byte   `json:"data,omitempty"`
		Seqno    []byte   `json:"seqno,omitempty"`
		TopicIDs []string `json:"topicIDs,omitempty"`
	}

	err := d.Decode(&r)
	if err != nil {
		return nil, err
	}

	from, err := peer.IDFromBytes(r.From)
	if err != nil {
		return nil, err
	}
	return &Message{
		From:     from,
		Data:     r.Data,
		Seqno:    r.Seqno,
		TopicIDs: r.TopicIDs,
	}, nil
}

// Cancel cancels the given subscription.
func (s *PubSubSubscription) Cancel() error {
	if s.resp.Output == nil {
		return nil
	}

	return s.resp.Output.Close()
}
