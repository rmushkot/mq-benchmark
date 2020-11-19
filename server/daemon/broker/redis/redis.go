package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"
)

const redisPort = 6379

var channelKey = broker.GenerateName()

// PubSubPeer implements the peer interface for Redis Pub/Sub
// for more info on RedisPubSub see https://redis.io/topics/pubsub
type PubSubPeer struct {
	conn     *redis.Client
	pubsub   *redis.PubSub
	messages <-chan *redis.Message
	send     chan []byte
	errors   chan error
	done     chan bool
}

// NewPeer creates a peer used for communicating with Redis
func NewPeer(host string) (*PubSubPeer, error) {
	p := PubSubPeer{}
	p.conn = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, redisPort),
	})
	p.send = make(chan []byte)
	p.errors = make(chan error)
	p.done = make(chan bool)
	return &p, nil
}

// Subscribe prepares the peer to consume messages.
func (p *PubSubPeer) Subscribe() error {
	p.pubsub = p.conn.Subscribe(context.Background(), channelKey)
	_, err := p.pubsub.Receive(context.Background())
	p.messages = p.pubsub.Channel()
	return err
}

// Recv returns a single message consumed by the peer. Subscribe must be
// called before this. It returns an error if the receive failed.
func (p *PubSubPeer) Recv() ([]byte, error) {
	msg := <-p.messages
	return []byte(msg.Payload), nil
}

// Send returns a channel on which messages can be sent for publishing.
func (p *PubSubPeer) Send() chan<- []byte {
	return p.send
}

// Errors returns the channel on which the peer sends publish errors.
func (p *PubSubPeer) Errors() <-chan error {
	return p.errors
}

// Done signals to the peer that message publishing has completed.
func (p *PubSubPeer) Done() {
	p.done <- true
}

// Setup prepares the peer for testing.
func (p *PubSubPeer) Setup() {
	publisherFn := func() {
		for {
			select {
			case msg := <-p.send:
				err := p.conn.Publish(context.Background(), channelKey, msg).Err()
				if err != nil {
					fmt.Println("Failed to publish", err)
					p.errors <- err
				}
			case <-p.done:
				p.done <- true // signal next peer to stop
				return
			}
		}
	}

	for i := 0; i < p.conn.Options().PoolSize; i++ {
		go publisherFn()
	}
}

// Teardown performs any cleanup logic that needs to be performed after the
// test is complete.
func (p *PubSubPeer) Teardown() {
	if p.pubsub != nil {
		p.pubsub.Close()
	}
	p.conn.Close()
}
