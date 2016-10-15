package shell

import (
	"encoding/base64"
	"encoding/json"
	//	"sync"
)

type b64String string

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

type PubSubRecord struct {
	From     string    `json:"from"`
	Data     b64String `json:"data"`
	SeqNo    b64String `json:"seqno"`
	TopicIDs []string  `json:"topicIDs"`
}

func (r PubSubRecord) DataString() string {
	return string(r.Data)
}

///

type PubSubSubscription struct {
	resp *Response
}

func newPubSubSubscription(resp *Response) *PubSubSubscription {
	return &PubSubSubscription{
		resp: resp,
	}
}

func (s *PubSubSubscription) Next() (*PubSubRecord, error) {
	if s.resp.Error != nil {
		return nil, s.resp.Error
	}

	d := json.NewDecoder(s.resp.Output)

	r := &PubSubRecord{}
	err := d.Decode(r)

	return r, err
}

func (s *PubSubSubscription) Cancel() error {
	if s.resp.Output == nil {
		return nil
	}

	return s.resp.Output.Close()
}

///

/*
type subscriptionHandler struct {
	topic string
	resp  *Response

	readers map[chan *PubSubRecord]struct{}

	stop      chan struct{}
	add, drop chan chan *PubSubRecord

	failReason error
}

func newPubSubSubscriptionHandler(topic string, resp *Response) *subscriptionHandler {
	sh := &subscriptionHandler{
		// the topic that is being handled
		topic: topic,
		// stop shuts down the subscription handler.
		stop: make(chan struct{}),
		// readers is the set of listeners
		readers: make(map[chan *PubSubRecord]struct{}),
		//add is the channel in which you add more listeners
		add: make(chan chan *PubSubRecord),
		//drop is the channel to which you send channels
		drop: make(chan chan *PubSubRecord),
		resp: resp,
	}

	go sh.work()

	return sh
}

func (sh *subscriptionHandler) work() {
	readOne := func(ch chan *PubSubRecord, errCh chan error) {
		d := json.NewDecoder(sh.resp.Output)
		if sh.resp.Error != nil {
			errCh <- sh.resp.Error
			return
		}

		r := PubSubRecord{}
		err := d.Decode(&r)
		if err != nil {
			errCh <- err
			return
		}

		ch <- &r
	}

	ch := make(chan *PubSubRecord)
	errCh := make(chan error)

	go readOne(ch, errCh)

L:
	for {
		select {
		// remove a rdCh from pool
		case ch := <-sh.drop:
			delete(sh.readers, ch)

			if len(sh.readers) == 0 {
				break L
			}

		// add a rdCh to pool
		case ch := <-sh.add:
			sh.readers[ch] = struct{}{}

		case r := <-ch:
			for rdCh := range sh.readers {
				rdCh <- r
			}

			go readOne(ch, errCh)

		case err := <-errCh:
			sh.failReason = err
			break L

		case <-sh.stop:
			break L
		}
	}

	for rdCh := range sh.readers {
		delete(sh.readers, rdCh)
		close(rdCh)
	}

	//sh.resp.Output.Close()
	sh = nil
}

func (sh *subscriptionHandler) Stop() {
	sh.stop <- struct{}{}
}

func (sh *subscriptionHandler) Sub() *PubSubSubscription {
	ch := make(chan *PubSubRecord)

	sh.add <- ch

	return newPubSubSubscription(sh.topic, ch)
}

func (sh *subscriptionHandler) Drop(s *PubSubSubscription) {
	sh.drop <- s.ch
}

func (sh *subscriptionHandler) Error() error {
	return sh.failReason
}

///

type subscriptionManager struct {
	sync.Mutex

	s    *Shell
	subs map[string]*subscriptionHandler
}

func newPubSubSubscriptionManager(s *Shell) *subscriptionManager {
	return &subscriptionManager{
		s:    s,
		subs: make(map[string]*subscriptionHandler),
	}
}

func (sm *subscriptionManager) Sub(topic string) (*PubSubSubscription, error) {
	// lock
	sm.Lock()
	defer sm.Unlock()

	// check if already subscribed
	sh := sm.subs[topic]
	if sh == nil { // if not, do so!
		// connect
		req := sm.s.newRequest("pubsub/sub", topic)
		resp, err := req.Send(sm.s.httpcli)
		if err != nil {
			return nil, err
		}

		// pass connection to handler and add handler to manager
		sh = newPubSubSubscriptionHandler(topic, resp)
		sm.subs[topic] = sh
	}

	// success
	return sh.Sub(), nil
}

func (sm *subscriptionManager) Drop(s *PubSubSubscription) {
	sm.Lock()
	defer sm.Unlock()

	sh := sm.subs[s.topic]
	if sh != nil {
		sh.Drop(s)
	}
}

func (sm *subscriptionManager) dropHandler(sh *subscriptionHandler) {
	sm.Lock()
	defer sm.Unlock()

	delete(sm.subs, sh.topic)
}
*/
