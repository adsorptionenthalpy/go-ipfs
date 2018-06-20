package p2p

import (
	"errors"
	"sync"

	ma "gx/ipfs/QmUxSEGbv2nmYNnfXi7839wwQqTN3kwQeUxe8dTjZWZs7J/go-multiaddr"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

// Listener listens for connections and proxies them to a target
type Listener interface {
	Protocol() protocol.ID
	ListenAddress() ma.Multiaddr
	TargetAddress() ma.Multiaddr

	start() error

	// Close closes the listener. Does not affect child streams
	Close() error
}

type listenerKey struct {
	proto  string
	listen string
	target string
}

// ListenerRegistry is a collection of local application proto listeners.
type ListenerRegistry struct {
	Listeners map[listenerKey]Listener
	lk        sync.Mutex
}

// Register registers listenerInfo into this registry and starts it
func (r *ListenerRegistry) Register(l Listener) error {
	r.lk.Lock()

	if _, ok := r.Listeners[getListenerKey(l)]; ok {
		r.lk.Unlock()
		return errors.New("listener already registered")
	}

	r.Listeners[getListenerKey(l)] = l

	r.lk.Unlock()

	if err := l.start(); err != nil {
		r.lk.Lock()
		defer r.lk.Lock()

		delete(r.Listeners, getListenerKey(l))
		return err
	}

	return nil
}

// Deregister removes p2p listener from this registry
func (r *ListenerRegistry) Deregister(k listenerKey) {
	r.lk.Lock()
	defer r.lk.Unlock()

	delete(r.Listeners, k)
}

func getListenerKey(l Listener) listenerKey {
	return listenerKey{
		proto:  string(l.Protocol()),
		listen: l.ListenAddress().String(),
		target: l.TargetAddress().String(),
	}
}
