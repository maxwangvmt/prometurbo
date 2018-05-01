package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
	"github.com/turbonomic/prometurbo/pkg/conf"
	"github.com/turbonomic/prometurbo/pkg/discovery"
	"github.com/turbonomic/prometurbo/pkg/discovery/monitoring"
	"github.com/turbonomic/prometurbo/pkg/registration"
	"os"
	"time"
	"os/signal"
	"syscall"
)

type disconnectFromTurboFunc func()

func main() {
	fmt.Printf("Starting server...\n")
	flag.Parse()

	serviceConf := conf.DefaultConfPath

	if os.Getenv("PROMETURBO_LOCAL_DEBUG") == "1" {
		serviceConf = conf.LocalDebugConfPath
		fmt.Printf("Using config file %s for local debugging", serviceConf)
	}

	conf, err := conf.NewPrometheusConf(serviceConf)
	if err != nil {
		fmt.Printf("Error while parsing the turbo communicator config file %v: %v\n", serviceConf, err)
		glog.Infof("Error while parsing the turbo communicator config file %v: %v\n", serviceConf, err)
		os.Exit(1)
	}

	fmt.Printf("conf: %++v\n", conf)

	communicator := conf.Communicator
	targetAddr := conf.TargetConf.Address //"target-address"
	scope := conf.TargetConf.Scope        //"scope"

	builder := monitoring.NewAppEntityBuilder(scope)
	metricExporters := []monitoring.IMetricExporter{monitoring.NewMetricExporter(conf.MetricExporterEndpoint)}

	registrationClient := &registration.PrometheusRegistrationClient{}
	discoveryClient := discovery.NewDiscoveryClient(targetAddr, builder, metricExporters)

	tapService, err := service.NewTAPServiceBuilder().
		WithTurboCommunicator(communicator).
		WithTurboProbe(probe.NewProbeBuilder(registration.TargetType, registration.ProbeCategory).
			RegisteredBy(registrationClient).
			DiscoversTarget(targetAddr, discoveryClient)).Create()

	if err != nil {
		glog.Infof("Error while building turbo tap service on target %v: %v\n", targetAddr, err)
		os.Exit(1)
	}

	// Before running service, wait for the exporter to start up
	// TODO: Check the readiness of the exporter
	time.Sleep(5 * time.Second)

	// Disconnect from Turbo server when Kubeturbo is shutdown
	handleExit(func() { tapService.DisconnectFromTurbo() })

	// Connect to the Turbo server
	tapService.ConnectToTurbo()

	select {}
}


// handleExit disconnects the tap service from Turbo service when Kubeturbo is shotdown
func handleExit(disconnectFunc disconnectFromTurboFunc) { //k8sTAPService *kubeturbo.K8sTAPService) {
	glog.V(4).Infof("*** Handling Prometurbo Termination ***")
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGHUP)

	go func() {
		select {
		case sig := <-sigChan:
		// Close the mediation container including the endpoints. It avoids the
		// invalid endpoints remaining in the server side. See OM-28801.
			glog.V(2).Infof("Signal %s received. Disconnecting from Turbo server...\n", sig)
			disconnectFunc()
		}
	}()
}
