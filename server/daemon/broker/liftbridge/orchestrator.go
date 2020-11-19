package liftbridge

import (
	"fmt"
	"log"
	"os/exec"
)

const (
	liftbridged  = "liftbridge/standalone-dev"
	internalPort = "4222"
	liftport     = "9292"
)

// Broker implements the broker interface for NATS.
type Broker struct {
	containerID string
}

// Start will start the message broker and prepare it for testing.
func (n *Broker) Start(host, port string) (interface{}, error) {
	containerID, err := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("docker run -d -p %s:%s -p %s:%s %s", internalPort, internalPort, liftport, liftport, liftbridged)).Output()
	if err != nil {
		log.Printf("Failed to start container %s: %s", liftbridged, err.Error())
		return "", err
	}

	log.Printf("Started container %s: %s", liftbridged, containerID)
	n.containerID = string(containerID)
	return string(containerID), nil
}

// Stop will stop the message broker.
func (n *Broker) Stop() (interface{}, error) {
	containerID, err := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("docker kill %s", n.containerID)).Output()
	if err != nil {
		log.Printf("Failed to stop container %s: %s", liftbridged, err.Error())
		return "", err
	}

	log.Printf("Stopped container %s: %s", liftbridged, n.containerID)
	n.containerID = ""
	return string(containerID), nil
}
