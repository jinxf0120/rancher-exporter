package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"prometheus-rancher-exporter/internal/utils"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"prometheus-rancher-exporter/collector"
	"prometheus-rancher-exporter/query/rancher"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	k8sClientBurst = 100
	k8sClientQPS   = 100
)

func getEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func main() {

	log.Info("Building Rancher Client")

	InClusterConfig := true

	var config *rest.Config
	var err error

	if strings.ToUpper(os.Getenv("RANCHER_EXPORTER_EXTERNAL_AUTH")) == "TRUE" {
		log.Info("RANCHER_EXPORTER_EXTERNAL_AUTH env variable set to true, using out of cluster config")
		InClusterConfig = false
	}

	Timer_GetLatestRancherVersion, err := strconv.Atoi(getEnv("TIMER_GET_LATEST_RANCHER_VERSION", "1"))
	InformerResyncPeriod, err := strconv.Atoi(getEnv("INFORMER_RESYNC_PERIOD", "300"))

	if InClusterConfig {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal("Unable to construct REST client")
		}

		config.Burst = k8sClientBurst
		config.QPS = k8sClientQPS
	} else {
		currentUser, err := user.Current()
		if err != nil {
			log.Fatal(err.Error())
		}

		kubeconfig := flag.String("kubeconfig", fmt.Sprintf("/home/%s/.kube/config", currentUser.Username), "absolute path to the kubeconfig file")
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			log.Fatal("Unable to construct Rancher client Config")
		}

		config.Burst = k8sClientBurst
		config.QPS = k8sClientQPS

	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to construct Rancher client")
	}

	RancherClient := rancher.Client{
		Config: config,
		Client: client,
	}

	rancherInstalled, rancherBackupsInstalled, err := utils.CheckInstalledRancherApps(RancherClient)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Printf("Detected Rancher: %s", strconv.FormatBool(rancherInstalled))
	log.Printf("Detected Rancher Backup Operator: %s", strconv.FormatBool(rancherBackupsInstalled))

	if !rancherInstalled && !rancherBackupsInstalled {
		log.Fatal("Neither Rancher nor Rancher Backup Operator detected, exiting")
	}

	log.Info("Using Informer mechanism for metric collection")
	if err := RancherClient.InitInformerManager(InformerResyncPeriod, rancherBackupsInstalled); err != nil {
		log.Fatalf("Failed to initialize informer manager: %v", err)
	}
	log.Infof("Informer manager initialized with resync period: %d seconds", InformerResyncPeriod)

	if rancherInstalled {
		log.Printf("Collecting Rancher Metrics")
		http.Handle("/metrics", promhttp.Handler())
		collector.RegisterHandlers(RancherClient)
	}

	if rancherBackupsInstalled {
		log.Printf("Collecting Rancher Backup Operator Metrics")
		reg := prometheus.NewRegistry()
		backupsHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
		http.Handle("/backup-metrics", backupsHandler)
		collector.RegisterBackupHandlers(RancherClient, reg)
	}

	if err := RancherClient.StartInformer(); err != nil {
		log.Fatalf("Failed to start informer: %v", err)
	}

	if rancherInstalled {
		go collector.RunBaseMetricsLoop(RancherClient, Timer_GetLatestRancherVersion)
	}

	log.Info("Beginning to serve on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
