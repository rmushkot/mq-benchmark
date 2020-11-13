package natsstreaming

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"
)

const (
	subject   = "test"
	clusterID = "test-cluster"
	clientID  = "stan-bench"
	// Maximum bytes we will get behind before we start slowing down publishing.
	maxBytesBehind = 1024 * 1024 // 1MB

	// Maximum msgs we will get behind before we start slowing down publishing.
	maxMsgsBehind = 65536 // 64k

	// Time to delay publishing when we are behind.
	delay = 0 * time.Millisecond
)

// Peer implements the peer interface for NATS.
type Peer struct {
	conn     *nats.Conn
	sconn    stan.Conn
	messages chan []byte
	send     chan []byte
	errors   chan error
	done     chan bool
}

// NewPeer creates and returns a new Peer for communicating with NATS.
func NewPeer(host string) (*Peer, error) {
	fmt.Println(host)
	conn, err := nats.Connect(fmt.Sprintf("nats://%s:4222", host)) // This needs to be the address of the host that has the docker container running ie "nats://13.58.149.44:4222"
	if err != nil {
		return nil, err
	}

	// We want to be alerted if we get disconnected, this will be due to Slow
	// Consumer.
	conn.Opts.AllowReconnect = false
	sc, err := stan.Connect(clusterID, broker.GenerateName(), stan.NatsConn(conn))
	if err != nil {
		return nil, err
	}

	return &Peer{
		conn:     conn,
		sconn:    sc,
		messages: make(chan []byte, 100000),
		send:     make(chan []byte),
		errors:   make(chan error, 1),
		done:     make(chan bool),
	}, nil
}

// Subscribe prepares the peer to consume messages.
func (n *Peer) Subscribe() error {
	n.sconn.Subscribe(subject, func(message *stan.Msg) {
		n.messages <- message.Data
	})
	return nil
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
				if err := n.sendMessage(msg); err != nil {
					n.errors <- err
				}
			case <-n.done:
				return
			}
		}
	}()
}

func (n *Peer) sendMessage(message []byte) error {
	// Check if we are behind by >= 1MB bytes.
	// bytesDeltaOver := n.conn.OutBytes-n.conn.InBytes >= maxBytesBehind

	// // Check if we are behind by >= 65k msgs.
	// msgsDeltaOver := n.conn.OutMsgs-n.conn.InMsgs >= maxMsgsBehind

	// // If we are behind on either condition, sleep a bit to catch up receiver.
	// if bytesDeltaOver || msgsDeltaOver {
	// 	time.Sleep(delay)
	// }
	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			log.Printf("Warning: error publishing msg id %s: %v\n", ackedNuid, err.Error())
		} else {
			return
		}
	}
	_, err := n.sconn.PublishAsync(subject, message, ackHandler)
	return err
}

// Teardown performs any cleanup logic that needs to be performed after the
// test is complete.
func (n *Peer) Teardown() {
	n.conn.Close()
	n.sconn.Close()
}
