package shell

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
)

type Message struct {
	From     peer.ID  `json:"from,omitempty"`
	Data     []byte   `json:"data,omitempty"`
	Seqno    []byte   `json:"seqno,omitempty"`
	TopicIDs []string `json:"topicIDs,omitempty"`
}

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
func (s *PubSubSubscription) Next() (*Message, error) {
	if s.resp.Error != nil {
		return nil, s.resp.Error
	}

	d := json.NewDecoder(s.resp.Output)

	var r Message
	err := d.Decode(&r)

	return &r, err
}

// Cancel cancels the given subscription.
func (s *PubSubSubscription) Cancel() error {
	if s.resp.Output == nil {
		return nil
	}

	return s.resp.Output.Close()
}
