package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"
)

var streamKey = broker.GenerateName()

const (
	dataKey          = "data"
	streamXReadCount = 5_000
)

// StreamsPeer implements the peer interface using Redis Streams
// for more info on Redis Streams see: https://redis.io/topics/streams-intro
type StreamsPeer struct {
	conn     *redis.Client
	pipe     redis.Pipeliner
	recvMsgs chan []byte
	recvErrs chan error
	lastID   string // last ID read from the Stream in order to not miss any messages
	send     chan []byte
	errors   chan error
	done     chan bool
}

func NewStreamsPeer(host string) (*StreamsPeer, error) {
	s := StreamsPeer{}
	s.conn = newClient(host)
	s.send = make(chan []byte)
	s.errors = make(chan error)
	s.done = make(chan bool)
	return &s, nil
}

// Subscribe prepares the peer to consume messages.
func (s *StreamsPeer) Subscribe() error {
	s.recvMsgs = make(chan []byte, streamXReadCount)
	s.recvErrs = make(chan error)
	// initial ID of 0 means that we want to start receiving messages from the beginning
	// of the stream's history.
	// this will probably not be the cause for other Stream clients, but for the
	// benchmark tool the stream will be empty at the start and reset each time its run
	s.lastID = "0"
	go func() {
		for {
			args := redis.XReadArgs{
				Streams: []string{streamKey, s.lastID},
				Count:   streamXReadCount,
				Block:   time.Duration(0),
			}
			res, err := s.conn.XRead(context.Background(), &args).Result()

			if err != nil {
				if err == redis.ErrClosed {
					log.Println("Consumer: Exiting XREAD loop")
					return
				}
				s.recvErrs <- err
			}

			for _, msg := range res[0].Messages {
				// set ID so that we will get every message sent by the producers
				s.lastID = msg.ID
				data := []byte(msg.Values[dataKey].(string))
				s.recvMsgs <- data
			}
		}
	}()
	return nil
}

// Recv returns a single message consumed by the peer. Subscribe must be
// called before this. It returns an error if the receive failed.
func (s *StreamsPeer) Recv() ([]byte, error) {
	select {
	case data := <-s.recvMsgs:
		return data, nil
	case err := <-s.recvErrs:
		return nil, err
	}
}

// Send returns a channel on which messages can be sent for publishing.
func (s *StreamsPeer) Send() chan<- []byte {
	return s.send
}

// Errors returns the channel on which the peer sends publish errors.
func (s *StreamsPeer) Errors() <-chan error {
	return s.errors
}

// Done signals to the peer that message publishing has completed.
func (s *StreamsPeer) Done() {
	s.pipe.Exec(context.Background())
	s.done <- true
}

// Setup prepares the peer for testing.
func (s *StreamsPeer) Setup() {
	s.pipe = s.conn.Pipeline()
	// Flush the pipeline after this many messages have been put into it
	const msgFlushCount = 50_000

	// helper function to clear the redis pipeline
	flushPipe := func() {
		_, err := s.pipe.Exec(context.Background())
		if err != nil {
			s.errors <- err
		}
	}

	go func() {
		msgs := 0
		for {
			select {
			case msg := <-s.send:
				args := redis.XAddArgs{
					Stream: streamKey,
					ID:     "*",
					Values: map[string]interface{}{dataKey: msg},
				}
				err := s.pipe.XAdd(context.Background(), &args).Err()
				if err != nil {
					fmt.Println("Failed to publish", err)
					s.errors <- err
				}
				msgs = (msgs + 1) % msgFlushCount
				if msgs == 0 {
					go flushPipe()
				}
			case <-s.done:
				return
			}
		}
	}()
}

// Teardown performs any cleanup logic that needs to be performed after the
// test is complete.
func (s *StreamsPeer) Teardown() {
	s.conn.Close()
}
