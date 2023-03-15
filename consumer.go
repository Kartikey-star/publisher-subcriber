package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
)

func consumer() {
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
			fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n and create release %v",
				*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value), rel)
		}
	}

	c.Close()

}

func InstallChart(message string) (*release.Release, error) {
	var req HelmRequest
	json.Unmarshal([]byte(message), &req)
	actionConfig := new(action.Configuration)
	helmDriver := os.Getenv("HELM_DRIVER")
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), helmDriver, nil); err != nil {
		fmt.Println(err)
	}
	actionConfig.KubeClient = &kubefake.PrintingKubeClient{Out: ioutil.Discard}
	cmd := action.NewInstall(actionConfig)
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
	// Add chart URL as an annotation before installation
	if ch.Metadata == nil {
		ch.Metadata = new(chart.Metadata)
	}
	if ch.Metadata.Annotations == nil {
		ch.Metadata.Annotations = make(map[string]string)
	}
	ch.Metadata.Annotations["chart_url"] = req.ChartUrl

	cmd.Namespace = req.Namespace
	release, err := cmd.Run(ch, req.Values)
	if err != nil {
		return nil, err
	}
	return release, nil
}
