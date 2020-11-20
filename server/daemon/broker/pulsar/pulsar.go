package pulsar

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"
)

var (
	topic = broker.GenerateName()
)

// Peer implements the peer interface for pulsar.
type Peer struct {
	conn     pulsar.Client
	producer pulsar.Producer
	consumer pulsar.Consumer
	messages chan pulsar.ConsumerMessage
	send     chan []byte
	errors   chan error
	done     chan bool
}

// NewPeer creates and returns a new Peer for communicating with pulsar.
func NewPeer(host string) (*Peer, error) {
	conn, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: fmt.Sprintf("pulsar://%s:6650", host),
	})
	if err != nil {
		return nil, err
	}
	producer, err := conn.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})

	if err != nil {
		return nil, err
	}

	channel := make(chan pulsar.ConsumerMessage)
	return &Peer{
		conn:     conn,
		producer: producer,
		consumer: nil,
		messages: channel,
		send:     make(chan []byte),
		errors:   make(chan error, 1),
		done:     make(chan bool),
	}, nil
}

// Subscribe prepares the peer to consume messages.
func (n *Peer) Subscribe() error {
	consumer, err := n.conn.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: "my-sub",
		Type:             pulsar.Shared,
		MessageChannel:   n.messages,
	})
	n.consumer = consumer
	return err
}

// Recv returns a single message consumed by the peer. Subscribe must be called
// before this. It returns an error if the receive failed.
func (n *Peer) Recv() ([]byte, error) {
	msg := <-n.messages
	n.consumer.Ack(msg)
	return msg.Payload(), nil
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
				n.producer.SendAsync(context.Background(), &pulsar.ProducerMessage{
					Payload: msg,
				}, func(id pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
					if err != nil {
						fmt.Println("Failed to publish", err)
						n.errors <- err
					}
				})
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
	if n.producer != nil {
		n.producer.Close()
	}
	if n.consumer != nil {
		n.consumer.Close()
	}
}
