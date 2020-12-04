package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"

)

var topic = broker.GenerateName() 

// Peer implements the peer interface for Kafka.
type Peer struct {
	writer *kafka.Writer
	reader *kafka.Reader
	host string
	send   chan []byte
	errors chan error
	done   chan bool
}

// NewPeer creates and returns a new Peer for communicating with Kafka.
func NewPeer(host string) (*Peer, error) {
	hostPort := fmt.Sprintf("%s:9092", host)
	// to create topics when auto.create.topics.enable='true'
	

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:       []string{hostPort},
		Topic:         topic,
		QueueCapacity: 100,
		BatchSize:     1000, // default 100
		BatchBytes:    100000,
		RequiredAcks:  1,
		Async:         true,
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{hostPort},
		Topic:   topic,
		// Partition:     0,
		// GroupID: "group1",
		QueueCapacity: 100, // default 100
		MinBytes:      10e4, // 10 KB
		MaxBytes:      10e6, // 10 MB
		// CommitInterval: time.Second,
		ReadBackoffMin: 5 * time.Millisecond,
	})

	return &Peer{
		writer: writer,
		reader: reader,
		host: hostPort,
		send:   make(chan []byte),
		errors: make(chan error, 1),
		done:   make(chan bool),
	}, nil
}

// Subscribe prepares the peer to consume messages.
func (k *Peer) Subscribe() error {
	// This ensures that a topic is created.
	conn, err := kafka.DialLeader(context.Background(), "tcp", k.host, topic, 0)
	conn.Close()
	return err
}

// Recv returns a single message consumed by the peer. Subscribe must be called
// before this. It returns an error if the receive failed.
func (k *Peer) Recv() ([]byte, error) {
	msg, err := k.reader.ReadMessage(context.Background())
	if err != nil {
		return nil, err
	}
	return msg.Value, nil
}

// Send returns a channel on which messages can be sent for publishing.
func (k *Peer) Send() chan<- []byte {
	return k.send
}

// Errors returns the channel on which the peer sends publish errors.
func (k *Peer) Errors() <-chan error {
	return k.errors
}

// Done signals to the peer that message publishing has completed.
func (k *Peer) Done() {
	k.done <- true
}

// Setup prepares the peer for testing.
func (k *Peer) Setup() {
	go func() {
		for {
			select {
			case msg := <-k.send:
				if err := k.writer.WriteMessages(context.Background(),
					kafka.Message{
						Value: msg}); err != nil {
					k.errors <- err
				}
			case <-k.done:
				return
			}
		}
	}()
}

// Teardown performs any cleanup logic that needs to be performed after the
// test is complete.
func (k *Peer) Teardown() {
	if err := k.reader.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
	if err := k.writer.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}
