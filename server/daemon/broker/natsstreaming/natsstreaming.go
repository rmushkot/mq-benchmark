package natsstreaming

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/rmushkot/mq-benchmark/server/daemon/broker"
)

var (
	subject   = broker.GenerateName()
	clusterID = "test-cluster"
	clientID  = broker.GenerateName()
	// Maximum bytes we will get behind before we start slowing down publishing.
	maxBytesBehind uint64 = 1024 * 1024 // 1MB

	// Maximum msgs we will get behind before we start slowing down publishing.
	maxMsgsBehind uint64 = 65536 // 64k

	// Time to delay publishing when we are behind.
	delay = 0 * time.Millisecond
)

// Peer implements the peer interface for NATS.
type Peer struct {
	conn     *nats.Conn
	sconn    stan.Conn
	sub      stan.Subscription
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
	sc, err := stan.Connect(clusterID, broker.GenerateName(), stan.MaxPubAcksInflight(10000), stan.NatsConn(conn))
	if err != nil {
		return nil, err
	}

	return &Peer{
		conn:     conn,
		sconn:    sc,
		sub:      nil,
		messages: make(chan []byte, 100000),
		send:     make(chan []byte),
		errors:   make(chan error, 1),
		done:     make(chan bool),
	}, nil
}

// Subscribe prepares the peer to consume messages.
func (n *Peer) Subscribe() error {
	s, err := n.sconn.Subscribe(subject, func(message *stan.Msg) {
		n.messages <- message.Data
	})
	s.SetPendingLimits(-1, -1)
	n.sub = s
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
				go func() {
					if err := n.sendMessage(msg); err != nil {
						n.errors <- err
					}
				}()
			case <-n.done:
				return
			}
		}
	}()
}

func (n *Peer) sendMessage(message []byte) error {
	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			log.Printf("Warning: error publishing msg id %s: %v\n", ackedNuid, err.Error())
		}
	}
	_, err := n.sconn.PublishAsync(subject, message, ackHandler)
	if err != nil {
		log.Printf("Error Publishing a message. %v", err.Error())
	}
	return err
}

// Teardown performs any cleanup logic that needs to be performed after the
// test is complete.
func (n *Peer) Teardown() {
	if n.sub != nil {
		n.sub.Unsubscribe()
	}
	n.conn.Close()
	n.sconn.Close()
}
