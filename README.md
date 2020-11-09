# mq-benchmark
Mq-benchmark is a tool for testing message queues in more realistic environments based on [Flotilla](https://github.com/tylertreat/flotilla) with some minor tweaks. This [blog post](http://www.bravenewgeek.com/benchmark-responsibly/) provides some more background on the methodology behind this project.

Testing anything at scale can be difficult to achieve in practice. It generally takes a lot of resources and often requires ad hoc solutions. Flotilla attempts to provide automated orchestration for benchmarking message queues in scaled-up configurations. Simply put, we can benchmark a message broker with arbitrarily many producers and consumers distributed across arbitrarily many machines with a single command.

```shell
./client \
    --broker=kafka \
    --host=192.168.59.100:9500 \
    --peer-hosts=localhost:9500,192.168.59.101:9500,192.168.59.102:9500,192.168.59.103:9500 \
    --producers=5 \
    --consumers=3 \
    --num-messages=1000000
    --message-size=5000
```

Mq-benchmark relies on [HDR Histogram](http://hdrhistogram.github.io/HdrHistogram/) (or rather a [Go variant](https://github.com/codahale/hdrhistogram) of it) which supports recording and analyzing sampled data value counts at extremely low latencies.

Flotilla supports several message brokers out of the box:

- [Beanstalkd](http://kr.github.io/beanstalkd/)
- [NATS](http://nats.io/)
- [NATS Streaming](https://docs.nats.io/nats-streaming-concepts/intro)
- [Kafka](http://kafka.apache.org/) **- See caveats below -**
- [ActiveMQ](http://activemq.apache.org/)
- [RabbitMQ](http://www.rabbitmq.com/)
- [NSQ](http://nsq.io/)

## Installation

Flotilla consists of two binaries: the server daemon and client. The daemon runs on any machines you wish to include in your tests. The client orchestrates and executes the tests. Note that the daemon makes use of [Docker](https://www.docker.com/) for running many of the brokers, so it must be installed on the host machine.

<!-- To install the daemon, run:

```bash
$ go get github.com/tylertreat/flotilla/flotilla-server
```

To install the client, run:

```bash
$ go get github.com/tylertreat/flotilla/flotilla-client
``` -->

## Usage

Ensure the daemon is running on any machines you wish mq-benchmark to communicate with:

Navigate to the /server directory
```bash
$ go build -o server main.go 
$ ./server
Benchmark daemon started on port 9500...
```

### Local Configuration

Flotilla can be run locally to perform benchmarks on a single machine. First, start the daemon with `./server`. Next, run a benchmark using the client:

Navigate to the /client directory
```bash
$ go build -o client main.go
$ ./client --broker=nats
```

Mq-benchmark will run everything on localhost.

### Distributed Configuration

With all daemons started, run a benchmark using the client and provide the peers you wish to communicate with:

```bash
$ ./client --broker=rabbitmq --host=<ip> --peer-hosts=<list of ips>
```

For full usage details, run:

```bash
$ ./client --help
```

## Caveats
- There is currently no security built in. Use this tool *at your own risk*. The daemon runs on port 9500 by default.
- Currently kafka can not be orchestrated through this project, so I recommend using docker-compose from [wurstmeister/kafka-docker](https://github.com/wurstmeister/kafka-docker).
