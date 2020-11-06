package natsstreaming

import (
	"fmt"
	"log"
	"os/exec"
)

const (
	gnatsd       = "nats-streaming"
	internalPort = "4223"
)

// Broker implements the broker interface for NATS.
type Broker struct {
	containerID string
}

// Start will start the message broker and prepare it for testing.
func (n *Broker) Start(host, port string) (interface{}, error) {
	containerID, err := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("sudo docker run -d %s -store file -dir /datastore", gnatsd)).Output()
	if err != nil {
		log.Printf("Failed to start container %s: %s", gnatsd, err.Error())
		return "", err
	}

	log.Printf("Started container %s: %s", gnatsd, containerID)
	n.containerID = string(containerID)
	return string(containerID), nil
}

// Stop will stop the message broker.
func (n *Broker) Stop() (interface{}, error) {
	containerID, err := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("sudo docker kill %s", n.containerID)).Output()
	if err != nil {
		log.Printf("Failed to stop container %s: %s", gnatsd, err.Error())
		return "", err
	}

	log.Printf("Stopped container %s: %s", gnatsd, n.containerID)
	n.containerID = ""
	return string(containerID), nil
}
