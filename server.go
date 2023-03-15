package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"helm.sh/helm/v3/pkg/action"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
)

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
	}
}

func main() {
	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/", handler).Methods("GET")
	router.HandleFunc("/installCharts", installCharts).Methods("POST")
	router.HandleFunc("/uninstallCharts", uninstallCharts).Methods("DELETE")

	actionConfig := new(action.Configuration)
	helmDriver := os.Getenv("HELM_DRIVER")
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), helmDriver, debug); err != nil {
		fmt.Println(err)
	}
	actionConfig.KubeClient = &kubefake.PrintingKubeClient{Out: ioutil.Discard}

	go installChartsConsumer(actionConfig)
	go uninstallChartsConsumer(actionConfig)
	http.ListenAndServe(":8080", router)
}

func handler(w http.ResponseWriter, r *http.Request) {
	responseJSON(w, http.StatusOK, "listening")
}

func responseJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func responseError(w http.ResponseWriter, status int, payload string) {
	responseJSON(w, status, payload)
}
