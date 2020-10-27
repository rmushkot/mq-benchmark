package kafka

import (
	"context"
	"strings"
	"time"

	kafkaP "github.com/segmentio/kafka-go"
)

const topic = "test"

// Peer implements the peer interface for Kafka.
type Peer struct {
	producer *kafkaP.Conn
	consumer *kafkaP.Conn
	send     chan []byte
	errors   chan error
	done     chan bool
}

// // NewPeer creates and returns a new Peer for communicating with Kafka.
// func NewPeer(host string) (*Peer, error) {
// 	host = strings.Split(host, ":")[0] + ":9092" //localhost:5000 into 9092
// 	config := sarama.NewConfig()
// 	fmt.Println(host)
// 	client, err := sarama.NewClient([]string{host}, config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	producer, err := sarama.NewAsyncProducer([]string{host}, config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	consumer, err := sarama.NewConsumer([]string{host}, config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Peer{
// 		client:   client,
// 		producer: producer,
// 		consumer: partitionConsumer,
// 		send:     make(chan []byte),
// 		errors:   make(chan error, 1),
// 		done:     make(chan bool),
// 	}, nil
// }

// NewPeer creates and returns a new Peer for communicating with Kafka.
func NewPeer(host string) (*Peer, error) {
	host = strings.Split(host, ":")[0] + ":9092" //localhost:5000 into 9092

	// producer := kafkaP.NewWriter(kafkaP.WriterConfig{
	// 	Brokers:  []string{host},
	// 	Topic:    topic,
	// 	Balancer: &kafkaP.LeastBytes{},
	// })

	// consumer := kafkaP.NewReader(kafkaP.ReaderConfig{
	// 	Brokers:   []string{host},
	// 	Topic:     topic,
	// 	Partition: 0,
	// 	MinBytes:  10e1,
	// })
	prod, err := kafkaP.DialLeader(context.Background(), "tcp", host, topic, 0)
	if err != nil {
		return nil, err
	}
	prod.SetWriteDeadline(time.Now().Add(10 * time.Second))

	conn, err := kafkaP.DialLeader(context.Background(), "tcp", host, topic, 0)
	if err != nil {
		return nil, err
	}
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	return &Peer{
		producer: prod,
		consumer: conn,
		send:     make(chan []byte),
		errors:   make(chan error, 1),
		done:     make(chan bool),
	}, nil
}

// Subscribe prepares the peer to consume messages.
func (k *Peer) Subscribe() error {
	return nil
}

// Recv returns a single message consumed by the peer. Subscribe must be called
// before this. It returns an error if the receive failed.
func (k *Peer) Recv() ([]byte, error) {
	// msg, err := k.consumer.ReadMessage(context.Background())
	// if err != nil {
	// 	return nil, err
	// }
	// return msg.Value, nil
	batch := k.consumer.ReadBatch(10, 1e6) // fetch 10KB min, 1MB max
	var err error
	b := make([]byte, 10e3)
	for {
		_, err := batch.Read(b)
		if err != nil {
			break
		}
	}
	return b, err
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
				if err := k.sendMessage(msg); err != nil {
					k.errors <- err
				}
			case <-k.done:
				return
			}
		}
	}()
}

func (k *Peer) sendMessage(message []byte) error {
	_, err := k.producer.WriteMessages(
		kafkaP.Message{
			Value: message,
		})
	if err != nil {
		return nil
	}
	return nil

}

// Teardown performs any cleanup logic that needs to be performed after the
// test is complete.
func (k *Peer) Teardown() {
	k.producer.Close()
	if k.consumer != nil {
		k.consumer.Close()
	}
}
