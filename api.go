package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type HelmRequest struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	ChartUrl  string                 `json:"chart_url"`
	Values    map[string]interface{} `json:"values"`
	Version   int                    `json:"version"`
}
type HelmCharts struct {
	Charts []HelmRequest `json:"charts"`
}

//sent request of charts to be installed
func installCharts(w http.ResponseWriter, r *http.Request) {
	var req HelmCharts
	err := json.NewDecoder(r.Body).Decode(&req)
	fmt.Println(err)
	topic := "helm_charts"
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		fmt.Printf("Failed to create producer: %s", err)
		panic(err)
	}

	// Go-routine to handle message delivery reports and
	// possibly other event types (errors, stats, etc)
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	for _, val := range req.Charts {
		stringJson, _ := json.Marshal(val)
		data := []byte(string(stringJson))
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          data,
		}, nil)
	}
	// Wait for all messages to be delivered
	p.Flush(15 * 1000)
	p.Close()

}
