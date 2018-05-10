package pkg

import (
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

type P8sTAPService struct {}

func (p *P8sTAPService) Start() {
	glog.V(0).Infof("Starting prometheus TAP service...")

	serviceConf := conf.DefaultConfPath

	if os.Getenv("PROMETURBO_LOCAL_DEBUG") == "1" {
		serviceConf = conf.LocalDebugConfPath
		fmt.Printf("Using config file %s for local debugging", serviceConf)
	}

	conf, err := conf.NewPrometurboConf(serviceConf)
	if err != nil {
		glog.Errorf("Error while parsing the turbo communicator config file %v: %v\n", serviceConf, err)
		os.Exit(1)
	}

	glog.V(2).Infof("conf: %++v\n", conf)

	communicator := conf.Communicator
	targetAddr := conf.TargetConf.Address
	scope := conf.TargetConf.Scope

	//builder := monitoring.NewEntityBuilder(scope)
	metricExporters := []monitoring.MetricExporter{monitoring.NewMetricExporter(conf.MetricExporterEndpoint)}

	registrationClient := &registration.P8sRegistrationClient{}
	discoveryClient := discovery.NewDiscoveryClient(targetAddr, scope, metricExporters)

	tapService, err := service.NewTAPServiceBuilder().
		WithTurboCommunicator(communicator).
		WithTurboProbe(probe.NewProbeBuilder(registration.TargetType, registration.ProbeCategory).
			//WithDiscoveryOptions(probe.FullRediscoveryIntervalSecondsOption(60)).
			RegisteredBy(registrationClient).
			DiscoversTarget(targetAddr, discoveryClient)).Create()

	if err != nil {
		glog.Errorf("Error while building turbo tap service on target %v: %v\n", targetAddr, err)
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


// TODO: Move the handle to turbo-sdk-probe as it should be common logic for similar probes
// handleExit disconnects the tap service from Turbo service when prometurbo is terminated
func handleExit(disconnectFunc disconnectFromTurboFunc) {
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
