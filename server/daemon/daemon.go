package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	"./broker/activemq"
	"./broker/amqp"
	"./broker/amqp/rabbitmq"
	"./broker/beanstalkd"
	"./broker/kafka"
	"./broker/nats"
	"./broker/nsq"
	// "./broker/redis"
	// "github.com/rmushkot/mq-benchmark/server/daemon/broker/broker/activemq"
	// "github.com/rmushkot/mq-benchmark/server/daemon/broker/broker/amqp"
	// "github.com/rmushkot/mq-benchmark/server/daemon/broker/broker/amqp/rabbitmq"
	// "github.com/rmushkot/mq-benchmark/server/daemon/broker/broker/beanstalkd"
	// "github.com/rmushkot/mq-benchmark/server/daemon/broker/broker/kafka"
	// "github.com/rmushkot/mq-benchmark/server/daemon/broker/broker/nats"
	// "github.com/rmushkot/mq-benchmark/server/daemon/broker/broker/nsq"
	// "github.com/rmushkot/mq-benchmark/server/daemon/broker/broker/redis"
)

type daemon string
type operation string

const (
	start    operation = "start"
	stop     operation = "stop"
	run      operation = "run"
	sub      operation = "subscribers"
	pub      operation = "publishers"
	results  operation = "results"
	teardown operation = "teardown"
)

// These are supported message brokers.
const (
	NATS       = "nats"
	Beanstalkd = "beanstalkd"
	Kafka      = "kafka"
	Kestrel    = "kestrel"
	ActiveMQ   = "activemq"
	RabbitMQ   = "rabbitmq"
	NSQ        = "nsq"
	Redis      = "redis"
)

type request struct {
	Operation   operation `json:"operation"`
	Broker      string    `json:"broker"`
	Port        string    `json:"port"`
	NumMessages int       `json:"num_messages"`
	MessageSize int64     `json:"message_size"`
	Count       int       `json:"count"`
	Host        string    `json:"host"`
}

type response struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Result     interface{} `json:"result"`
	PubResults []*result   `json:"pub_results,omitempty"`
	SubResults []*result   `json:"sub_results,omitempty"`
}

type result struct {
	Duration   float32         `json:"duration,omitempty"`
	Throughput float32         `json:"throughput,omitempty"`
	Latency    *latencyResults `json:"latency,omitempty"`
	Err        string          `json:"error,omitempty"`
}

// broker handles configuring the message broker for testing.
type broker interface {
	// Start will start the message broker and prepare it for testing.
	Start(string, string) (interface{}, error)

	// Stop will stop the message broker.
	Stop() (interface{}, error)
}

// peer is a single producer or consumer in the test.
type peer interface {
	// Subscribe prepares the peer to consume messages.
	Subscribe() error

	// Recv returns a single message consumed by the peer. Subscribe must be
	// called before this. It returns an error if the receive failed.
	Recv() ([]byte, error)

	// Send returns a channel on which messages can be sent for publishing.
	Send() chan<- []byte

	// Errors returns the channel on which the peer sends publish errors.
	Errors() <-chan error

	// Done signals to the peer that message publishing has completed.
	Done()

	// Setup prepares the peer for testing.
	Setup()

	// Teardown performs any cleanup logic that needs to be performed after the
	// test is complete.
	Teardown()
}

// Daemon is the server portion of Flotilla which runs on machines we want to
// communicate with and include in our benchmarks.
type Daemon struct {
	con         net.Conn
	lis         net.Listener
	broker      broker
	publishers  []*publisher
	subscribers []*subscriber
}

// NewDaemon creates and returns a new Daemon from the provided Config. An
// error is returned if the Daemon cannot be created.
func NewDaemon() (*Daemon, error) {
	var rep net.Listener
	var conn net.Conn
	// rep.AddTransport(tcp.NewTransport())
	return &Daemon{conn, rep, nil, []*publisher{}, []*subscriber{}}, nil
}

// Start will allow the Daemon to begin processing requests. This is a blocking
// call.
func (d *Daemon) Start(port int) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
	ln, err := net.ListenTCP("tcp", tcpAddr)
	// defer ln.Close()
	if err != nil {
		return err
	}
	d.lis = ln
	return d.loop()
}

func (d *Daemon) loop() error {
	for {
		c, err := d.lis.Accept()
		if err != nil {
			fmt.Println(err)
			continue
			// return err
		}
		defer c.Close()
		d.con = c

		var req request
		decoder := json.NewDecoder(c)
		if err := decoder.Decode(&req); err != nil {
			fmt.Println(err)
			return err
		}

		resp := d.processRequest(req)
		d.sendResponse(resp)
	}
}

func (d *Daemon) sendResponse(rep response) {
	encoder := json.NewEncoder(d.con)
	encoder.Encode(rep)
}

func (d *Daemon) processRequest(req request) response {
	var (
		response response
		err      error
	)
	switch req.Operation {
	case start:
		response.Result, err = d.processBrokerStart(req.Broker, req.Host, req.Port)
	case stop:
		response.Result, err = d.processBrokerStop()
	case pub:
		err = d.processPub(req)
	case sub:
		err = d.processSub(req)
	case run:
		err = d.processPublisherStart()
	case results:
		response.PubResults, response.SubResults, err = d.processResults()
		if err != nil {
			response.Message = err.Error()
			err = nil
		}
	case teardown:
		d.processTeardown()
	default:
		err = fmt.Errorf("Invalid operation %s", req.Operation)
	}

	if err != nil {
		response.Message = err.Error()
	} else {
		response.Success = true
	}

	return response
}
func (d *Daemon) processBrokerStart(broker, host, port string) (interface{}, error) {
	if d.broker != nil {
		return "", errors.New("Broker already running")
	}

	switch broker {
	case NATS:
		d.broker = &nats.Broker{}
	case Beanstalkd:
		d.broker = &beanstalkd.Broker{}
	case Kafka:
		d.broker = &kafka.Broker{}
	case ActiveMQ:
		d.broker = &activemq.Broker{}
	case RabbitMQ:
		d.broker = &rabbitmq.Broker{}
	case NSQ:
		d.broker = &nsq.Broker{}
	// case Redis:
	// 	d.broker = &nsq.Broker{}
	default:
		return "", fmt.Errorf("Invalid broker %s", broker)
	}

	result, err := d.broker.Start(host, port)
	if err != nil {
		d.broker = nil
	}
	return result, err
}

func (d *Daemon) processBrokerStop() (interface{}, error) {
	if d.broker == nil {
		return "", errors.New("No broker running")
	}

	result, err := d.broker.Stop()
	if err == nil {
		d.broker = nil
	}
	return result, err
}

func (d *Daemon) processPub(req request) error {
	for i := 0; i < req.Count; i++ {
		sender, err := d.newPeer(req.Broker, req.Host)
		if err != nil {
			return err
		}

		d.publishers = append(d.publishers, &publisher{
			peer:        sender,
			id:          i,
			numMessages: req.NumMessages,
			messageSize: req.MessageSize,
		})
	}

	return nil
}

func (d *Daemon) processSub(req request) error {
	for i := 0; i < req.Count; i++ {
		receiver, err := d.newPeer(req.Broker, req.Host)
		if err != nil {
			return err
		}

		if err := receiver.Subscribe(); err != nil {
			return err
		}

		subscriber := &subscriber{
			peer:        receiver,
			id:          i,
			numMessages: req.NumMessages,
			messageSize: req.MessageSize,
		}
		d.subscribers = append(d.subscribers, subscriber)
		go subscriber.start()
	}

	return nil
}

func (d *Daemon) processPublisherStart() error {
	for _, publisher := range d.publishers {
		go publisher.start()
	}

	return nil
}

func (d *Daemon) processResults() ([]*result, []*result, error) {
	subResults := make([]*result, 0, len(d.subscribers))
	for _, subscriber := range d.subscribers {
		result, err := subscriber.getResults()
		if err != nil {
			return nil, nil, err
		}
		subResults = append(subResults, result)
	}

	pubResults := make([]*result, 0, len(d.publishers))
	for _, publisher := range d.publishers {
		result, err := publisher.getResults()
		if err != nil {
			return nil, nil, err
		}
		pubResults = append(pubResults, result)
	}

	log.Println("Benchmark completed")
	return pubResults, subResults, nil
}

func (d *Daemon) processTeardown() {
	for _, subscriber := range d.subscribers {
		subscriber.Teardown()
	}
	d.subscribers = d.subscribers[:0]

	for _, publisher := range d.publishers {
		publisher.Teardown()
	}
	d.publishers = d.publishers[:0]
}

func (d *Daemon) newPeer(broker, host string) (peer, error) {
	switch broker {
	case NATS:
		return nats.NewPeer(host)
	case Beanstalkd:
		return beanstalkd.NewPeer(host)
	case Kafka:
		return kafka.NewPeer(host)
	case ActiveMQ:
		return activemq.NewPeer(host)
	case RabbitMQ:
		return amqp.NewPeer(host)
	case NSQ:
		return nsq.NewPeer(host)
	// case Redis:
	// 	return redis.NewPeer(host)
	default:
		return nil, fmt.Errorf("Invalid broker: %s", broker)
	}
}
