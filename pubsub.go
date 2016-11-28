package shell

import (
	"encoding/base64"
	"encoding/json"
)

// floodsub uses base64 encoding for just about everything.
// To Decode the base64 while also decoding the JSON the type b64String was added.
type b64String string

// UnmarshalJSON implements the json.Unmarshaler interface.
func (bs *b64String) UnmarshalJSON(in []byte) error {
	var b64 string

	err := json.Unmarshal(in, &b64)
	if err != nil {
		return err
	}

	bsStr, err := base64.StdEncoding.DecodeString(b64)

	*bs = b64String(bsStr)
	return err
}

func (bs *b64String) Marshal() (string, error) {
	jsonBytes, err := json.Marshal(
		base64.StdEncoding.EncodeToString(
			[]byte(*bs)))

	return string(jsonBytes), err
}

///

// PubSubRecord is a record received via PubSub.
type PubSubRecord struct {
	From     string    `json:"from"`
	Data     b64String `json:"data"`
	SeqNo    b64String `json:"seqno"`
	TopicIDs []string  `json:"topicIDs"`
}

// DataString returns the string representation of the data field.
func (r PubSubRecord) DataString() string {
	return string(r.Data)
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

	sub.Next() // skip empty element used for flushing
	return sub
}

// Next waits for the next record and returns that.
func (s *PubSubSubscription) Next() (*PubSubRecord, error) {
	if s.resp.Error != nil {
		return nil, s.resp.Error
	}

	d := json.NewDecoder(s.resp.Output)

	r := &PubSubRecord{}
	err := d.Decode(r)

	return r, err
}

// Cancel cancels the given subscription.
func (s *PubSubSubscription) Cancel() error {
	if s.resp.Output == nil {
		return nil
	}

	return s.resp.Output.Close()
}
