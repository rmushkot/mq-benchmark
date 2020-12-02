package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"
)

var (
	subject = broker.GenerateName()
)

// Peer implements the peer interface for NATS.
type Peer struct {
	conn         *nats.Conn
	subscription *nats.Subscription
	messages     chan []byte
	send         chan []byte
	errors       chan error
	done         chan bool
}

// NewPeer creates and returns a new Peer for communicating with NATS.
func NewPeer(host string) (*Peer, error) {
	conn, err := nats.Connect(fmt.Sprintf("nats://%s:4222", host)) // This needs to be the address of the host that has the docker container running ie "nats://13.58.149.44:4222"
	if err != nil {
		return nil, err
	}

	// We want to be alerted if we get disconnected, this will be due to Slow
	// Consumer.
	conn.Opts.AllowReconnect = false

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
	sub, err := n.conn.Subscribe(subject, func(message *nats.Msg) {
		n.messages <- message.Data
	})
	sub.SetPendingLimits(-1, -1)
	n.conn.Flush()
	n.subscription = sub
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
				// go func() {
					if err := n.conn.Publish(subject, msg); err != nil {
						n.errors <- err
					}
				// }()
			case <-n.done:
				n.conn.Flush()
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
