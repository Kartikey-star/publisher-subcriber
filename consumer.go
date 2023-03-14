package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/santhosh-tekuri/jsonschema/loader"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/tools/clientcmd"
)

func consumer() {
	fmt.Println("Reaching Here")
	topic := "helm_charts"
	c, err := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092", "group.id": "kafka-go-getting-started"})
	if err != nil {
		fmt.Printf("Failed to create consumer: %s", err)
		panic(err)
	}
	err = c.SubscribeTopics([]string{topic}, nil)
	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev, err := c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Errors are informational and automatically handled by the consumer
				continue
			}
			var rel *release.Release
			if rel, err = InstallChart(string(ev.Value)); err != nil {
				// Errors are informational and automatically handled by the consumer
				fmt.Println(err)
				continue
			}
			fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n and create release %s",
				*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value), rel)
		}
	}

	c.Close()

}

func initSettings() *cli.EnvSettings {
	conf := cli.New()
	conf.RepositoryCache = "/tmp"
	return conf
}

var settings = initSettings()

func InstallChart(message string) (*release.Release, error) {
	var req HelmRequest
	json.Unmarshal([]byte(message), req)
	config, _ := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	cmd := action.NewInstall(&conf)

	releaseName, chartName, err := cmd.NameAndChart([]string{req.Name, req.ChartUrl})
	if err != nil {
		return nil, err
	}
	cmd.ReleaseName = releaseName

	cp, err := cmd.ChartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		return nil, err
	}

	ch, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	cmd.Namespace = req.Namespace
	release, err := cmd.Run(ch, req.Values)
	if err != nil {
		return nil, err
	}
	return release, nil
}
