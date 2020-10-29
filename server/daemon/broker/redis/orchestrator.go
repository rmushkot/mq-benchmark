package redis

import (
	"fmt"
	"log"
	"os/exec"
)

const (
	redisd = "redis"
)

// Broker implements the broker interface for NATS.
type Broker struct {
	containerID string
}

// Start will start the message broker and prepare it for testing.
func (n *Broker) Start(host, port string) (interface{}, error) {
	containerID, err := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("docker run -d %s", redisd)).Output()
	if err != nil {
		log.Printf("Failed to start container %s: %s", redisd, err.Error())
		return "", err
	}

	log.Printf("Started container %s: %s", redisd, containerID)
	n.containerID = string(containerID)
	return string(containerID), nil
}

// Stop will stop the message broker.
func (n *Broker) Stop() (interface{}, error) {
	containerID, err := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("docker kill %s", n.containerID)).Output()
	if err != nil {
		log.Printf("Failed to stop container %s: %s", redisd, err.Error())
		return "", err
	}

	log.Printf("Stopped container %s: %s", redisd, n.containerID)
	n.containerID = ""
	return string(containerID), nil
}
