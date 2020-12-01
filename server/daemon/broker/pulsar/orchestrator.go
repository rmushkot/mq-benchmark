package pulsar

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

const (
	pulsard       = "apachepulsar/pulsar"
	internalPort  = "6650"
	dataPath      = "/mq-benchmark/results/data/pulsar"
	pulsarCommand = `docker run -d \
						-p %s:%s \
						%s \
						bin/pulsar standalone`
)

// Broker implements the broker interface for NATS.
type Broker struct {
	containerID string
}

// Start will start the message broker and prepare it for testing.
func (n *Broker) Start(host, port string) (interface{}, error) {
	cmd := fmt.Sprintf(pulsarCommand, internalPort, internalPort, pulsard)
	containerID, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		log.Printf("Failed to start container %s: %s", pulsard, err.Error())
		return "", err
	}

	log.Printf("Started container %s: %s", pulsard, containerID)
	n.containerID = string(containerID)
	time.Sleep(10 * time.Second)
	return string(containerID), nil
}

// Stop will stop the message broker.
func (n *Broker) Stop() (interface{}, error) {
	containerID, err := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("docker kill %s", n.containerID)).Output()
	if err != nil {
		log.Printf("Failed to stop container %s: %s", pulsard, err.Error())
		return "", err
	}

	log.Printf("Stopped container %s: %s", pulsard, n.containerID)
	n.containerID = ""
	return string(containerID), nil
}
