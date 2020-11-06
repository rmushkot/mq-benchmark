package pulsar

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

const (
	topic = "test"
)

// Peer implements the peer interface for pulsar.
type Peer struct {
	conn     *pulsar.Conn
	producer *pulsar.Producer
	consumer *pulsar.Consumer
	messages chan []byte
	send     chan []byte
	errors   chan error
	done     chan bool
}

// NewPeer creates and returns a new Peer for communicating with pulsar.
func NewPeer(host string) (*Peer, error) {
	conn, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               fmt.Sprintf("pulsar://%s:6650", host),
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})

	if err != nil {
		return nil, err
	}

	return &Peer{
		conn:     conn,
		producer: producer,
		consumer: nil,
		send:     make(chan []byte),
		errors:   make(chan error, 1),
		done:     make(chan bool),
	}, nil
}

// Subscribe prepares the peer to consume messages.
func (n *Peer) Subscribe() error {
	consumer, err := n.conn.Subscribe(pulsar.ConsumerOptions{
		Topic: topic,
	})
	n.consumer = consumer
}

// Recv returns a single message consumed by the peer. Subscribe must be called
// before this. It returns an error if the receive failed.
func (n *Peer) Recv() ([]byte, error) {
	msg, err = n.consumer.Receive(context.Background())
	return msg.Payload(), err
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
				_, err := n.producer.Send(context.Background(), &pulsar.ProducerMessage{
					Payload: msg,
				})
				if err != nil {
					n.errors <- err
				}
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
	n.producer.Close()
	n.consumer.Close()
}
