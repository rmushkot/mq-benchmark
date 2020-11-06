package pulsar

import (
	"fmt"
	"log"
	"os/exec"
)

const (
	pulsard       = "apachepulsar/pulsar"
	internalPort  = "6650"
	dataPath      = "../../../../results/data/pulsar"
	pulsarCommand = `docker run -d \
						-p %s:%s \
						-p 8080:8080 \
						-v %s \
						%s \
						bin/pulsar standalone`
)

// Broker implements the broker interface for NATS.
type Broker struct {
	containerID string
}

// Start will start the message broker and prepare it for testing.
func (n *Broker) Start(host, port string) (interface{}, error) {
	cmd := fmt.Sprintf(pulsarCommand, internalPort, internalPort, dataPath, pulsard)
	containerID, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		log.Printf("Failed to start container %s: %s", pulsard, err.Error())
		return "", err
	}

	log.Printf("Started container %s: %s", pulsard, containerID)
	n.containerID = string(containerID)
	return string(containerID), nil
}

// Stop will stop the message broker.
func (n *Broker) Stop() (interface{}, error) {
	containerID, err := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("sudo docker kill %s", n.containerID)).Output()
	if err != nil {
		log.Printf("Failed to stop container %s: %s", pulsard, err.Error())
		return "", err
	}

	log.Printf("Stopped container %s: %s", pulsard, n.containerID)
	n.containerID = ""
	return string(containerID), nil
}
