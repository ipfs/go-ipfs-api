package shell

import (
	"encoding/json"
	"io"

	"github.com/libp2p/go-libp2p-peer"
)

// Message is a pubsub message.
type Message struct {
	From     peer.ID
	Data     []byte
	Seqno    []byte
	TopicIDs []string
}

// PubSubSubscription allow you to receive pubsub records that where published on the network.
type PubSubSubscription struct {
	resp io.Closer
	dec  *json.Decoder
}

func newPubSubSubscription(resp io.ReadCloser) *PubSubSubscription {
	return &PubSubSubscription{
		resp: resp,
		dec:  json.NewDecoder(resp),
	}
}

// Next waits for the next record and returns that.
func (s *PubSubSubscription) Next() (*Message, error) {
	var r struct {
		From     []byte   `json:"from,omitempty"`
		Data     []byte   `json:"data,omitempty"`
		Seqno    []byte   `json:"seqno,omitempty"`
		TopicIDs []string `json:"topicIDs,omitempty"`
	}

	err := s.dec.Decode(&r)
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
	return s.resp.Close()
}
