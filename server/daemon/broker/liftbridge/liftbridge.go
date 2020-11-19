package liftbridge

import (
	"context"
	"fmt"

	lift "github.com/liftbridge-io/go-liftbridge/v2"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"
)

var (
	subject = broker.GenerateName()
	name    = subject + "-stream"
)

// Peer implements the peer interface for pulsar.
type Peer struct {
	conn     lift.Client
	messages chan []byte
	send     chan []byte
	errors   chan error
	done     chan bool
}

// NewPeer creates and returns a new Peer for communicating with pulsar.
func NewPeer(host string) (*Peer, error) {
	addr := []string{fmt.Sprintf("%s:9292", host)}
	conn, err := lift.Connect(addr)
	if err != nil {
		return nil, err
	}

	if err := conn.CreateStream(context.Background(), subject, name); err != nil {
		if err != lift.ErrStreamExists {
			return nil, err
		}
	}

	return &Peer{
		conn:     conn,
		messages: make(chan []byte),
		send:     make(chan []byte),
		errors:   make(chan error, 1),
		done:     make(chan bool),
	}, nil
}

// Subscribe prepares the peer to consume messages.
func (n *Peer) Subscribe() error {
	err := n.conn.Subscribe(context.Background(), name, func(msg *lift.Message, err error) {
		n.messages <- msg.Value()
	})

	return err
}

// Recv returns a single message consumed by the peer. Subscribe must be called
// before this. It returns an error if the receive failed.
func (n *Peer) Recv() ([]byte, error) {
	return <-n.messages, nil
}

// Send returns a channel on which messages can be sent for publishing.
func (n *Peer) Send() chan<- []byte {
	return n.send
}

// Errors returns the channel on which the peer sends publish errors.
func (n *Peer) Errors() <-chan error {
	return n.errors
}

// Done signals to the peer that message publishing has completed.
func (n *Peer) Done() {
	n.done <- true
}

// Setup prepares the peer for testing.
func (n *Peer) Setup() {
	go func() {
		for {
			select {
			case msg := <-n.send:
				n.conn.PublishAsync(context.Background(), name, msg, nil)
			case <-n.done:
				return
			}
		}
	}()
}

// Teardown performs any cleanup logic that needs to be performed after the
// test is complete.
func (n *Peer) Teardown() {
	n.conn.Close()
}
