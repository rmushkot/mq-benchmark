package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"
)

var channelKey = broker.GenerateName()

// PubSubPeer implements the peer interface for Redis Pub/Sub
// for more info on RedisPubSub see https://redis.io/topics/pubsub
type PubSubPeer struct {
	conn     *redis.Client
	pipe     redis.Pipeliner
	pubsub   *redis.PubSub
	messages <-chan *redis.Message
	send     chan []byte
	errors   chan error
	done     chan bool
}

// NewPeer creates a peer used for communicating with Redis
func NewPubSubPeer(host string) (*PubSubPeer, error) {
	p := PubSubPeer{}
	p.conn = newClient(host)
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
	p.pipe.Exec(context.Background())
	p.done <- true
}

// Setup prepares the peer for testing.
func (p *PubSubPeer) Setup() {
	p.pipe = p.conn.Pipeline()

	// Flush the pipeline after this many messages have been put into it
	const msgFlushCount = 50_000

	// helper function to clear the redis pipeline
	flushPipe := func() {
		_, err := p.pipe.Exec(context.Background())
		if err != nil {
			p.errors <- err
		}
	}

	go func() {
		msgs := 0
		for {
			select {
			case msg := <-p.send:
				err := p.pipe.Publish(context.Background(), channelKey, msg).Err()
				if err != nil {
					fmt.Println("Failed to publish", err)
					p.errors <- err
				}
				msgs = (msgs + 1) % msgFlushCount
				if msgs == 0 {
					go flushPipe()
				}
			case <-p.done:
				return
			}
		}
	}()
}

// Teardown performs any cleanup logic that needs to be performed after the
// test is complete.
func (p *PubSubPeer) Teardown() {
	if p.pubsub != nil {
		p.pubsub.Close()
	}
	p.conn.Close()
}
