package shell

import (
	"context"
	"testing"

	"github.com/cheekybits/is"
)

func TestP2PListener(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	listener, err := s.P2POpenListener(context.Background(), "p2p-open-listener-test", "/ip4/127.0.0.1/udp/3000")
	is.Nil(err)
	is.Equal(listener.Address, "/ip4/127.0.0.1/udp/3000")
	is.Equal(listener.Protocol, "/p2p/p2p-open-listener-test")

	listenerList, err := s.P2PListListeners(context.Background())
	is.Nil(err)
	is.Equal(len(listenerList.Listeners), 1)
	is.Equal(listenerList.Listeners[0].Address, "/ip4/127.0.0.1/udp/3000")
	is.Equal(listenerList.Listeners[0].Protocol, "/p2p/p2p-open-listener-test")

	is.Nil(s.P2PCloseListener(context.Background(), "p2p-open-listener-test", false))

	listenerList, err = s.P2PListListeners(context.Background())
	is.Nil(err)
	is.Equal(len(listenerList.Listeners), 0)
}

func TestP2PStreams(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	streamsList, err := s.P2PListStreams(context.Background())
	is.Nil(err)
	is.Equal(len(streamsList.Streams), 0)
}
