package kafka

const (
	zookeeper = "bitnami/zookeeper"
	// zookeeperCmd  = "docker run -d -p %s:%s -e ALLOW_ANONYMOUS_LOGIN=yes %s"
	zookeeperCmd = `docker run -d --name zookeeper-server \
							-p 127.0.0.1:2181:2181 \
							--network app-tier \
							-e ALLOW_ANONYMOUS_LOGIN=yes \
							bitnami/zookeeper:latest`
	zookeeperPort = "2181"
	kafka         = "bitnami/kafka"
	kafkaPort     = "9092"
	jmxPort       = "7203"
	// TODO: Use --link.
	kafkaCmd = `docker run -d --name kafka-server \
					-p 127.0.0.1:9092:9092 \
					--network app-tier \
					-e ALLOW_PLAINTEXT_LISTENER=yes \
					-e KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper-server:2181 \
					bitnami/kafka:latest`
	dockerCmd = "docker-compose up"
	docker
)

// Broker implements the broker interface for Kafka.
type Broker struct {
	kafkaContainerID     string
	zookeeperContainerID string
}

// Start will start the message broker and prepare it for testing.
func (k *Broker) Start(host, port string) (interface{}, error) {
	// fmt.Println(host, ":", port)
	// if port == zookeeperPort || port == jmxPort {
	// 	return nil, fmt.Errorf("Port %s is reserved", port)
	// }

	// // cmd := fmt.Sprintf(zookeeperCmd, zookeeperPort, zookeeperPort, zookeeper)
	// cmd := fmt.Sprintf(dockerCmd)
	// _, err := exec.Command("/bin/sh", "-c", cmd).Output()
	// if err != nil {
	// 	log.Printf("Failed to start container %s: %s", zookeeper, err.Error())
	// 	return "", err
	// }
	// log.Printf("Started container")

	// time.Sleep(time.Minute)
	// // cmd = fmt.Sprintf(kafkaCmd, host, kafkaPort, kafkaPort, jmxPort, jmxPort, host, host, kafka)
	// cmd = fmt.Sprintf(kafkaCmd)
	// kafkaContainerID, err := exec.Command("/bin/sh", "-c", cmd).Output()
	// if err != nil {
	// 	log.Printf("Failed to start container %s: %s", kafka, err.Error())
	// 	k.Stop()
	// 	return "", err
	// }

	// log.Printf("Started container %s: %s", kafka, kafkaContainerID)
	// k.kafkaContainerID = string(kafkaContainerID)
	// k.zookeeperContainerID = string(zkContainerID)

	// // NOTE: Leader election can take a while. For now, just sleep to try to
	// // ensure the cluster is ready. Is there a way to avoid this or make it
	// // better?
	// time.Sleep(time.Minute)
	kafkaContainerID := "123456"

	k.kafkaContainerID = string(kafkaContainerID)
	k.zookeeperContainerID = string(kafkaContainerID)
	return string(kafkaContainerID), nil
}

// Stop will stop the message broker.
func (k *Broker) Stop() (interface{}, error) {
	// _, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker kill %s", k.zookeeperContainerID)).Output()
	// if err != nil {
	// 	log.Printf("Failed to stop container %s: %s", zookeeper, err.Error())
	// } else {
	// 	log.Printf("Stopped container %s: %s", zookeeper, k.zookeeperContainerID)
	// }

	// kafkaContainerID, e := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker kill %s", k.kafkaContainerID)).Output()
	// if e != nil {
	// 	log.Printf("Failed to stop container %s: %s", kafka, err.Error())
	// 	err = e
	// } else {
	// 	log.Printf("Stopped container %s: %s", kafka, k.kafkaContainerID)
	// }
	kafkaContainerID := "123456"
	return string(kafkaContainerID), nil
}
