package redis

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

const (
	redisDockerImg = "redis"
	redisDockerCmd = "docker run -d -p %d:%d %s"
)

// Broker implements broker interface for Redis
type Broker struct {
	containerID string
}

// Start will start the message broker and prepare it for testing.
func (b *Broker) Start(host, port string) (interface{}, error) {
	cmd := exec.Command("docker", "run", "-d", "-p",
		fmt.Sprintf("%d:%d", redisPort, redisPort), redisDockerImg)
	containerID, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to start container %s: %s", redisDockerImg, err.Error())
		return "", err
	}
	b.containerID = strings.TrimSpace(string(containerID))
	log.Printf("Started container %s: %s", redisDockerImg, b.containerID)
	time.Sleep(10 * time.Second)
	return b.containerID, nil
}

// Stop will stop the message broker.
func (b *Broker) Stop() (interface{}, error) {
	cmd := exec.Command("docker", "kill", b.containerID)
	containerID, err := cmd.Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			log.Println(string(err.Stderr))
		}
		log.Printf("Failed to stop container %s: %s", redisDockerImg, err.Error())
		return "", err
	}

	log.Printf("Stopped container %s: %s", redisDockerImg, b.containerID)
	b.containerID = ""
	return string(containerID), nil
}
