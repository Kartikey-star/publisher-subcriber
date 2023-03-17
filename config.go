package main

import (
	"net/http"

	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var settings = initSettings()

type configFlagsWithTransport struct {
	*genericclioptions.ConfigFlags
	Transport *http.RoundTripper
}

func initSettings() *cli.EnvSettings {
	conf := cli.New()
	return conf
}
