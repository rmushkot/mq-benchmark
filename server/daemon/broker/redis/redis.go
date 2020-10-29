package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

const (
	key = "key"
)

// Peer implements the peer interface for redis.
type Peer struct {
	conn   *redis.Client
	ctx    context.Context
	send   chan []byte
	errors chan error
	done   chan bool
}

// NewPeer creates and returns a new Peer for communicating with redis.
func NewPeer(host string) (*Peer, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	return &Peer{
		conn:   rdb,
		ctx:    ctx,
		send:   make(chan []byte),
		errors: make(chan error, 1),
		done:   make(chan bool),
	}, nil
}

// Subscribe prepares the peer to consume messages.
func (n *Peer) Subscribe() error {
	return nil
}

// Recv returns a single message consumed by the peer. Subscribe must be called
// before this. It returns an error if the receive failed.
func (n *Peer) Recv() ([]byte, error) {
	val, err := n.conn.Get(n.ctx, key).Result()
	b := []byte(val)
	return b, err
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
				if err := n.conn.Set(n.ctx, key, msg, 0).Err(); err != nil {
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
}
