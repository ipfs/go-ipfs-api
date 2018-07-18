package shell

import (
	"context"
	"strconv"

	ma "github.com/multiformats/go-multiaddr"
)

// P2PListener describes a P2P listener.
type P2PListener struct {
	Protocol string
	Address  string
}

// P2POpenListener forwards P2P connections to a network multiaddr.
func (s *Shell) P2POpenListener(ctx context.Context, protocol, maddr string) (*P2PListener, error) {
	if _, err := ma.NewMultiaddr(maddr); err != nil {
		return nil, err
	}

	var response *P2PListener
	err := s.Request("p2p/listener/open").
		Arguments(protocol, maddr).Exec(ctx, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// P2PCloseListener closes one or all active P2P listeners.
func (s *Shell) P2PCloseListener(ctx context.Context, protocol string, closeAll bool) error {
	req := s.Request("p2p/listener/close").
		Option("all", strconv.FormatBool(closeAll))

	if protocol != "" {
		req.Arguments(protocol)
	}

	if err := req.Exec(ctx, nil); err != nil {
		return err
	}

	return nil
}

// P2PListenerList contains a slice of P2PListeners.
type P2PListenerList struct {
	Listeners []*P2PListener
}

// P2PListListeners lists all P2P listeners.
func (s *Shell) P2PListListeners(ctx context.Context) (*P2PListenerList, error) {
	var response *P2PListenerList

	if err := s.Request("p2p/listener/ls").Exec(ctx, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// P2PStream describes a P2P stream.
type P2PStream struct {
	Protocol string
	Address  string
}

// P2PStreamDial dials to a peer's P2P listener.
func (s *Shell) P2PStreamDial(ctx context.Context, peerID, protocol, listenerMaddr string) (*P2PStream, error) {
	var response *P2PStream
	req := s.Request("p2p/stream/dial").
		Arguments(peerID, protocol)

	if listenerMaddr != "" {
		if _, err := ma.NewMultiaddr(listenerMaddr); err != nil {
			return nil, err
		}
		req.Arguments(listenerMaddr)
	}

	if err := req.Exec(ctx, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// P2PCloseStream closes one or all active P2P streams.
func (s *Shell) P2PCloseStream(ctx context.Context, handlerID string, closeAll bool) error {
	req := s.Request("p2p/stream/close").
		Option("all", strconv.FormatBool(closeAll))

	if handlerID != "" {
		req.Arguments(handlerID)
	}

	if err := req.Exec(ctx, nil); err != nil {
		return err
	}

	return nil
}

// P2PStreamsList contains a slice of streams.
type P2PStreamsList struct {
	Streams []*struct {
		HandlerID     string
		Protocol      string
		LocalPeer     string
		LocalAddress  string
		RemotePeer    string
		RemoteAddress string
	}
}

// P2PListStreams lists all P2P streams.
func (s *Shell) P2PListStreams(ctx context.Context) (*P2PStreamsList, error) {
	var response *P2PStreamsList
	req := s.Request("p2p/stream/ls").
		Option("headers", strconv.FormatBool(true))

	if err := req.Exec(ctx, &response); err != nil {
		return nil, err
	}

	return response, nil
}
